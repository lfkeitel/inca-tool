package parser

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

// TaskFile represents a parsed task file
type TaskFile struct {
	Name        string
	Description string
	Author      string
	Date        string
	Version     string
	Concurrent  int32

	DeviceList string
	Devices    []string

	currentBlock string
	Commands     map[string]*CommandBlock

	Mode string
}

// CommandBlock contains all the settings for a block of commands
type CommandBlock struct {
	Name     string
	Type     string
	Commands []string
	Template string
}

const (
	modeRoot = iota
	modeCommand
	modeDevices
)

var (
	runningMode  = modeRoot
	wsRegex      = regexp.MustCompile(`^(\s+)`)
	currentSigWs = ""
)

// Parse will load the file filename and put it into a TaskFile struct or return and error if something goes wrong
func Parse(filename string) (*TaskFile, error) {
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return nil, fmt.Errorf("Task file does not exist: %s\n", filename)
	}

	return parse(filename)
}

func parse(filename string) (*TaskFile, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)
	task := &TaskFile{}
	lineNum := 0

	for scanner.Scan() {
		// Get next line
		line := scanner.Text()
		lineNum++

		if len(line) < 1 || line[0] == '#' {
			continue
		}

		if runningMode == modeCommand {
			if err := parseCommandLine(line, task, lineNum); err != nil {
				return nil, err
			}
		} else if runningMode == modeDevices {
			if err := parseDeviceLine(line, task, lineNum); err != nil {
				return nil, err
			}
		} else {
			if err := parseLine(line, task, lineNum); err != nil {
				return nil, err
			}
		}
	}

	if err := finishUp(task); err != nil {
		return nil, err
	}

	return task, nil
}

func parseLine(line string, task *TaskFile, lineNum int) error {
	// Split only on the first colon
	runningMode = modeRoot
	parts := strings.SplitN(line, ":", 2)

	if len(parts) != 2 {
		return fmt.Errorf("Error on line %d of task file\n", lineNum)
	}
	setting := strings.ToLower(parts[0])
	setting = strings.Title(setting)
	setting = strings.Replace(setting, " ", "", -1)
	setting = strings.TrimSpace(setting)
	settingVal := strings.TrimSpace(parts[1])

	switch setting {
	case "Commands":
		return parseCommandBlockStart(setting, settingVal, task, lineNum)
	case "Devices":
		runningMode = modeDevices
		return nil
	}

	taskReflect := reflect.ValueOf(task)
	// struct
	s := taskReflect.Elem()
	// exported field
	f := s.FieldByName(setting)
	if f.IsValid() {
		// A Value can be changed only if it is
		// addressable and was not obtained by
		// the use of unexported struct fields.
		if f.CanSet() {
			// change value of N
			if f.Kind() == reflect.String {
				if f.String() != "" {
					return fmt.Errorf("Cannot redeclare setting '%s'. Line %d\n", setting, lineNum)
				}
				f.SetString(settingVal)
			} else if f.Kind() == reflect.Int32 {
				if f.Int() > 0 {
					return fmt.Errorf("Cannot redeclare setting '%s'. Line %d\n", setting, lineNum)
				}

				i, err := strconv.Atoi(settingVal)
				if err != nil {
					return fmt.Errorf("Expected integer on line %d\n", lineNum)
				}
				f.SetInt(int64(i))
			}
		}
	} else {
		return fmt.Errorf("Invalid setting \"%s\". Line %d\n", setting, lineNum)
	}

	return nil
}

func parseCommandBlockStart(cmd, opts string, task *TaskFile, lineNum int) error {
	if task.Commands == nil {
		task.Commands = make(map[string]*CommandBlock)
	}

	if opts == "" {
		return fmt.Errorf("%s blocks must have a name. Line %d\n", cmd, lineNum)
	}

	pieces := strings.Split(opts, " ")
	name := pieces[0]

	_, set := task.Commands[name]
	if set {
		return fmt.Errorf("%s block with name '%s' already exists. Line %d\n", cmd, opts, lineNum)
	}

	task.Commands[name] = &CommandBlock{
		Name: name,
	}

	if len(pieces) > 1 {
		for _, setting := range pieces[1:] {
			parts := strings.Split(setting, "=")
			if len(parts) < 2 {
				continue
			}

			taskReflect := reflect.ValueOf(task.Commands[name])
			// struct
			s := taskReflect.Elem()
			// exported field
			f := s.FieldByName(strings.Title(parts[0]))
			if f.IsValid() {
				// A Value can be changed only if it is
				// addressable and was not obtained by
				// the use of unexported struct fields.
				if f.CanSet() {
					// change value of N
					if f.Kind() == reflect.String {
						if f.String() != "" {
							return fmt.Errorf("Cannot redeclare setting '%s'. Line %d\n", parts[0], lineNum)
						}
						f.SetString(parts[1])
					}
				}
			} else {
				return fmt.Errorf("Invalid block setting \"%s\". Line %d\n", parts[0], lineNum)
			}
		}
	}
	task.currentBlock = name
	runningMode = modeCommand
	return nil
}

func parseCommandLine(line string, task *TaskFile, lineNum int) error {
	matches := wsRegex.FindStringSubmatch(line)
	if len(matches) == 0 {
		return parseLine(line, task, lineNum)
	}
	sigWs := matches[0]
	current := task.Commands[task.currentBlock]

	if len(current.Commands) == 0 {
		currentSigWs = sigWs
	} else {
		if sigWs != currentSigWs {
			return fmt.Errorf("Command not in block, check indention. Line %d\n", lineNum)
		}
	}

	line = strings.TrimSpace(line)
	current.Commands = append(current.Commands, line)
	return nil
}

func parseDeviceLine(line string, task *TaskFile, lineNum int) error {
	matches := wsRegex.FindStringSubmatch(line)
	if len(matches) == 0 {
		return parseLine(line, task, lineNum)
	}
	sigWs := matches[0]

	if len(task.Devices) == 0 {
		currentSigWs = sigWs
	} else {
		if sigWs != currentSigWs {
			return fmt.Errorf("Device not in block, check indention. Line %d\n", lineNum)
		}
	}

	line = strings.TrimSpace(line)
	task.Devices = append(task.Devices, line)
	return nil
}

func finishUp(task *TaskFile) error {
	if task.Concurrent <= 0 {
		task.Concurrent = 300
	}

	if _, ok := task.Commands["main"]; !ok {
		return errors.New("No main command block declared")
	}
	return nil
}

package taskfileparser

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
	Description string
	Author      string
	Date        string
	Version     string
	Concurrent  int32

	DeviceList   string
	Manufacturer string
	Group        string
	Filter       string

	Script               string
	AdditionalArgs       string
	AdditionalArgsParsed []string

	Username       string
	Password       string
	EnablePassword string

	Commands          []string
	commandWhitespace string

	Expect           string
	expectWhitespace string

	Mode string
}

const (
	modeRoot = iota
	modeCommand
	modeExpect
)

var (
	runningMode = modeRoot
	wsRegex     = regexp.MustCompile(`^(\s+)`)
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
		} else if runningMode == modeExpect {
			if err := parseExpectLine(line, task, lineNum); err != nil {
				return nil, err
			}
		} else {
			if err := parseLine(line, task, lineNum); err != nil {
				return nil, err
			}
		}
	}

	if err := checkFilterSettings(task); err != nil {
		return nil, err
	}

	parseAdditionalArgs(task)

	return task, nil
}

func parseLine(line string, task *TaskFile, lineNum int) error {
	// Split only on the first colon
	parts := strings.SplitN(line, ":", 2)

	if len(parts) != 2 {
		return fmt.Errorf("Error on line %d of task file\n", lineNum)
	}
	setting := strings.ToLower(parts[0])
	setting = strings.Title(setting)
	setting = strings.Replace(setting, " ", "", -1)
	setting = strings.TrimSpace(setting)
	settingVal := strings.TrimSpace(parts[1])

	err, end := checkForCollision(setting, task, lineNum)
	if end {
		return err
	}

	taskReflect := reflect.ValueOf(task)
	// struct
	s := taskReflect.Elem()
	if s.Kind() == reflect.Struct {
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
	}

	return nil
}

func checkForCollision(setting string, task *TaskFile, lineNum int) (error, bool) {
	end := false
	if setting == "Commands" {
		end = true
		if task.Script != "" || task.Expect != "" {
			return fmt.Errorf("Commands cannot be used together with script or expect. Line %d\n", lineNum), end
		}
		runningMode = modeCommand
		task.Mode = "command"
	}

	if setting == "Expect" {
		end = true
		if task.Script != "" || len(task.Commands) != 0 {
			return fmt.Errorf("Expect cannot be used with commands or script. Line %d\n", lineNum), end
		}
		runningMode = modeExpect
		task.Mode = "expect"
	}

	if setting == "Script" {
		if task.Expect != "" || len(task.Commands) != 0 {
			return fmt.Errorf("Script cannot be used with commands or expect. Line %d\n", lineNum), true
		}
		task.Mode = "script"
	}

	return nil, end
}

func parseCommandLine(line string, task *TaskFile, lineNum int) error {
	matches := wsRegex.FindStringSubmatch(line)
	if len(matches) == 0 {
		return parseLine(line, task, lineNum)
	}
	sigWs := matches[0]

	if len(task.Commands) == 0 {
		task.commandWhitespace = sigWs
	} else {
		if sigWs != task.commandWhitespace {
			return fmt.Errorf("Command not in block, check indention. Line %d\n", lineNum)
		}
	}

	line = strings.TrimSpace(line)
	task.Commands = append(task.Commands, line)
	return nil
}

func parseExpectLine(line string, task *TaskFile, lineNum int) error {
	matches := wsRegex.FindStringSubmatch(line)
	if len(matches) == 0 {
		return parseLine(line, task, lineNum)
	}
	sigWs := matches[0]

	if task.Expect == "" {
		task.expectWhitespace = sigWs
	} else {
		if sigWs != task.expectWhitespace {
			return fmt.Errorf("Expect line not in block, check indention. Line %d\n", lineNum)
		}
	}

	line = strings.TrimSpace(line)
	task.Expect += line + "\n"
	return nil
}

func checkFilterSettings(task *TaskFile) error {
	if task.Filter != "" {
		if task.Manufacturer != "" || task.Group != "" {
			return errors.New("Cannot use Filter with Group or Manufacturer\n")
		}
	}

	if task.Manufacturer == "" || strings.ToLower(task.Manufacturer) == "all" {
		task.Manufacturer = "*"
	}
	if task.Group == "" || strings.ToLower(task.Group) == "all" {
		task.Group = "*"
	}

	task.Filter = task.Manufacturer + ":" + task.Group
	return nil
}

func parseAdditionalArgs(task *TaskFile) {
	if task.AdditionalArgs != "" {
		task.AdditionalArgsParsed = strings.Split(task.AdditionalArgs, ";")

		for i, arg := range task.AdditionalArgsParsed {
			task.AdditionalArgsParsed[i] = strings.TrimSpace(arg)
		}
	}
}

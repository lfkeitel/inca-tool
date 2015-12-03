package parser

import (
	"bufio"
	"errors"
	"fmt"
	"io"
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
	Template    string
	Prompt      string

	DeviceList string
	Devices    []string

	currentBlock string
	Commands     map[string]*CommandBlock
}

// CommandBlock contains all the settings for a block of commands
type CommandBlock struct {
	Name     string
	Type     string
	Commands []string
}

const (
	modeRoot = iota
	modeCommand
	modeDevices
)

var (
	wsRegex = regexp.MustCompile(`^(\s+)`)
)

type Parser struct {
	runningMode  int
	currentSigWs string
	mainReflect  reflect.Value
	reflected    bool
	task         *TaskFile
}

func NewParser() *Parser {
	p := &Parser{}
	p.Clean()
	return p
}

func (p *Parser) Clean() {
	p.runningMode = modeRoot
	p.currentSigWs = ""
	p.reflected = false
	p.task = &TaskFile{}
}

// ParseFile will load the file filename and put it into a TaskFile struct or return an error if something goes wrong
func (p *Parser) ParseFile(filename string) (*TaskFile, error) {
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return nil, fmt.Errorf("Task file does not exist: %s\n", filename)
	}

	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	if err := p.parse(file); err != nil {
		return nil, err
	}
	return p.task, nil
}

func (p *Parser) ParseString(data string) (*TaskFile, error) {
	if err := p.parse(strings.NewReader(data)); err != nil {
		return nil, err
	}
	return p.task, nil
}

func (p *Parser) parse(reader io.Reader) error {
	p.Clean()

	// Create scanner
	scanner := bufio.NewScanner(reader)
	scanner.Split(bufio.ScanLines)
	lineNum := 0
	p.reflected = false

	for scanner.Scan() {
		// Get next line
		lineRaw := scanner.Text()
		lineTrimmed := strings.TrimSpace(lineRaw)
		lineNum++

		// Check for blank lines and comments
		if len(lineTrimmed) < 1 || lineTrimmed[0] == '#' {
			continue
		}

		if p.runningMode == modeCommand {
			if err := p.parseCommandLine(lineRaw, lineNum); err != nil {
				return err
			}
		} else if p.runningMode == modeDevices {
			if err := p.parseDeviceLine(lineRaw, lineNum); err != nil {
				return err
			}
		} else {
			if err := p.parseLine(lineRaw, lineNum); err != nil {
				return err
			}
		}
	}

	if err := p.finishUp(); err != nil {
		return err
	}

	return nil
}

func (p *Parser) parseLine(line string, lineNum int) error {
	// Split only on the first colon
	p.runningMode = modeRoot
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
		return p.parseCommandBlockStart(setting, settingVal, lineNum)
	case "Devices":
		p.runningMode = modeDevices
		return nil
	}

	if !p.reflected {
		p.mainReflect = reflect.ValueOf(p.task)
		p.reflected = true
	}
	// struct
	s := p.mainReflect.Elem()
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

func (p *Parser) parseCommandBlockStart(cmd, opts string, lineNum int) error {
	if p.task.Commands == nil {
		p.task.Commands = make(map[string]*CommandBlock)
	}

	if opts == "" {
		return fmt.Errorf("%s blocks must have a name. Line %d\n", cmd, lineNum)
	}

	pieces := strings.Split(opts, " ")
	name := pieces[0]

	_, set := p.task.Commands[name]
	if set {
		return fmt.Errorf("%s block with name '%s' already exists. Line %d\n", cmd, opts, lineNum)
	}

	p.task.Commands[name] = &CommandBlock{
		Name: name,
	}

	if len(pieces) > 1 {
		for _, setting := range pieces[1:] {
			parts := strings.Split(setting, "=")
			if len(parts) < 2 {
				continue
			}

			taskReflect := reflect.ValueOf(p.task.Commands[name])
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
	p.task.currentBlock = name
	p.runningMode = modeCommand
	return nil
}

func (p *Parser) parseCommandLine(line string, lineNum int) error {
	matches := wsRegex.FindStringSubmatch(line)
	if len(matches) == 0 {
		return p.parseLine(line, lineNum)
	}
	sigWs := matches[0]
	current := p.task.Commands[p.task.currentBlock]

	if len(current.Commands) == 0 {
		p.currentSigWs = sigWs
	} else {
		if sigWs != p.currentSigWs {
			return fmt.Errorf("Command not in block, check indention. Line %d\n", lineNum)
		}
	}

	line = strings.TrimSpace(line)
	current.Commands = append(current.Commands, line)
	return nil
}

func (p *Parser) parseDeviceLine(line string, lineNum int) error {
	matches := wsRegex.FindStringSubmatch(line)
	if len(matches) == 0 {
		return p.parseLine(line, lineNum)
	}
	sigWs := matches[0]

	if len(p.task.Devices) == 0 {
		p.currentSigWs = sigWs
	} else {
		if sigWs != p.currentSigWs {
			return fmt.Errorf("Device not in block, check indention. Line %d\n", lineNum)
		}
	}

	line = strings.TrimSpace(line)
	p.task.Devices = append(p.task.Devices, line)
	return nil
}

func (p *Parser) finishUp() error {
	if p.task.Concurrent <= 0 {
		p.task.Concurrent = 300
	}

	if p.task.DeviceList == "" {
		p.task.DeviceList = "devices.conf"
	}

	if _, ok := p.task.Commands["main"]; !ok {
		return errors.New("No main command block declared")
	}
	return nil
}

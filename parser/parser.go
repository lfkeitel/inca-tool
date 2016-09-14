package parser

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

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
	p.task.Metadata = make(map[string]string)
}

// ParseFile will load the file filename and put it into a TaskFile struct or return an error if something goes wrong
func ParseFile(filename string) (*TaskFile, error) {
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return nil, fmt.Errorf("Task file does not exist: %s\n", filename)
	}

	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	p := NewParser()
	if err := p.parse(file); err != nil {
		return nil, err
	}
	return p.task, nil
}

func ParseString(data string) (*TaskFile, error) {
	p := NewParser()
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
		lineRaw := scanner.Bytes()
		lineTrimmed := bytes.TrimSpace(lineRaw)
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

func (p *Parser) parseLine(line []byte, lineNum int) error {
	// Split only on the first colon
	p.runningMode = modeRoot
	parts := bytes.SplitN(line, []byte(":"), 2)

	if len(parts) != 2 {
		return fmt.Errorf("Error on line %d of task file\n", lineNum)
	}
	setting := parts[0]
	settingVal := bytes.TrimSpace(parts[1])

	if bytes.Equal(setting, []byte("commands")) {
		return p.parseCommandBlockStart(setting, settingVal, lineNum)
	}
	if bytes.Equal(setting, []byte("devices")) {
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
	f := s.FieldByName(normalizeKeyToField(setting))
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
				f.SetString(string(settingVal))
			} else if f.Kind() == reflect.Int32 {
				if f.Int() > 0 {
					return fmt.Errorf("Cannot redeclare setting '%s'. Line %d\n", setting, lineNum)
				}

				i, err := strconv.Atoi(string(settingVal))
				if err != nil {
					return fmt.Errorf("Expected integer on line %d\n", lineNum)
				}
				f.SetInt(int64(i))
			}
		}

		return nil
	}

	// Custom data
	if setting[0] == '$' {
		p.task.Metadata["_"+string(setting[1:])] = string(settingVal)
		return nil
	}

	// Standard metadata
	if isStandardMetadata(string(setting)) {
		p.task.Metadata[string(setting)] = string(settingVal)
		return nil
	}
	return fmt.Errorf("Invalid setting \"%s\". Line %d\n", setting, lineNum)
}

func (p *Parser) parseCommandBlockStart(cmd, opts []byte, lineNum int) error {
	if p.task.Commands == nil {
		p.task.Commands = make(map[string]*CommandBlock)
	}

	if bytes.Equal(opts, []byte("")) {
		return fmt.Errorf("%s blocks must have a name. Line %d\n", cmd, lineNum)
	}

	pieces := bytes.Split(opts, []byte(" "))
	name := string(pieces[0])

	_, set := p.task.Commands[name]
	if set {
		return fmt.Errorf("%s block with name '%s' already exists. Line %d\n", cmd, opts, lineNum)
	}

	p.task.Commands[name] = &CommandBlock{
		Name: name,
	}

	if len(pieces) > 1 {
		for _, setting := range pieces[1:] {
			parts := bytes.Split(setting, []byte("="))
			if len(parts) < 2 {
				continue
			}

			taskReflect := reflect.ValueOf(p.task.Commands[name])
			// struct
			s := taskReflect.Elem()
			// exported field
			f := s.FieldByName(string(bytes.Title(parts[0])))
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
						f.SetString(string(parts[1]))
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

func (p *Parser) parseCommandLine(line []byte, lineNum int) error {
	matches := wsRegex.FindSubmatch(line)
	if len(matches) == 0 {
		return p.parseLine(line, lineNum)
	}
	sigWs := string(matches[0])
	current := p.task.Commands[p.task.currentBlock]

	if len(current.Commands) == 0 {
		p.currentSigWs = sigWs
	} else {
		if sigWs != p.currentSigWs {
			return fmt.Errorf("Command not in block, check indention. Line %d\n", lineNum)
		}
	}

	line = bytes.TrimSpace(line)
	current.Commands = append(current.Commands, string(line))
	return nil
}

func (p *Parser) parseDeviceLine(line []byte, lineNum int) error {
	matches := wsRegex.FindSubmatch(line)
	if len(matches) == 0 {
		return p.parseLine(line, lineNum)
	}
	sigWs := string(matches[0])

	if len(p.task.Devices) == 0 {
		p.currentSigWs = sigWs
	} else {
		if sigWs != p.currentSigWs {
			return fmt.Errorf("Device not in block, check indention. Line %d\n", lineNum)
		}
	}

	line = bytes.TrimSpace(line)
	p.task.Devices = append(p.task.Devices, string(line))
	return nil
}

func (p *Parser) finishUp() error {
	if p.task.Concurrent <= 0 {
		p.task.Concurrent = 300
	}

	if _, ok := p.task.Commands["main"]; !ok {
		return errors.New("No main command block declared")
	}
	return nil
}

func isStandardMetadata(s string) bool {
	for _, m := range standardMetadata {
		if s == m {
			return true
		}
	}
	return false
}

func normalizeKeyToField(k []byte) string {
	k = bytes.ToLower(k)
	k = bytes.Title(k)
	k = bytes.Replace(k, []byte(" "), []byte(""), -1)
	k = bytes.TrimSpace(k)
	return string(k)
}

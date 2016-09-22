package task

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
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
	task         *Task
	currentLine  int
	currentFile  string
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
	p.task = &Task{}
	p.task.Metadata = make(map[string]string)
	p.currentLine = 0
}

// ParseFile will load the file filename and put it into a Task struct or return an error if something goes wrong
func ParseFile(filename string) (*Task, error) {
	filename, _ = filepath.Abs(filename)
	if stat, err := os.Stat(filename); os.IsNotExist(err) || stat.IsDir() {
		return nil, fmt.Errorf("Task file does not exist: %s\n", filename)
	}

	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	p := NewParser()
	if err := p.parse(file, filename); err != nil {
		return nil, err
	}
	return p.task, nil
}

func ParseString(data string) (*Task, error) {
	p := NewParser()
	if err := p.parse(strings.NewReader(data), ""); err != nil {
		return nil, err
	}
	return p.task, nil
}

func (p *Parser) parseIncludeFile(filename string) error {
	// Build file path relative to parent
	if !filepath.IsAbs(filename) {
		filename = filepath.Join(filepath.Dir(p.currentFile), filename)
	}

	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return fmt.Errorf("Task file does not exist: %s\n", filename)
	}

	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	p.currentLine = 0
	p.currentFile = filename

	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)

	if err := p.scan(scanner); err != nil {
		return err
	}
	return nil
}

func (p *Parser) parse(reader io.Reader, filename string) error {
	p.Clean()

	// Create scanner
	scanner := bufio.NewScanner(reader)
	scanner.Split(bufio.ScanLines)
	p.reflected = false
	p.currentFile = filename

	if err := p.scan(scanner); err != nil {
		return err
	}

	return nil
}

func (p *Parser) scan(scanner *bufio.Scanner) error {
	for scanner.Scan() {
		// Get next line
		lineRaw := scanner.Bytes()
		lineTrimmed := bytes.TrimSpace(lineRaw)
		p.currentLine++

		// Check for blank lines and comments
		if len(lineTrimmed) < 1 || lineTrimmed[0] == '#' {
			continue
		}

		if lineTrimmed[0] == '@' {
			incFilename := string(lineTrimmed[1:])
			// Save current state
			curLine := p.currentLine
			curFile := p.currentFile
			if err := p.parseIncludeFile(incFilename); err != nil {
				return err
			}
			// Restore state
			p.currentLine = curLine
			p.currentFile = curFile
			continue
		}

		if p.runningMode == modeCommand {
			if err := p.parseCommandLine(lineRaw); err != nil {
				return err
			}
		} else if p.runningMode == modeDevices {
			if err := p.parseDeviceLine(lineRaw); err != nil {
				return err
			}
		} else {
			if err := p.parseLine(lineRaw); err != nil {
				return err
			}
		}
	}
	return nil
}

func (p *Parser) parseLine(line []byte) error {
	// Split only on the first colon
	p.runningMode = modeRoot
	parts := bytes.SplitN(line, []byte(":"), 2)

	if len(parts) != 2 {
		return fmt.Errorf("Error on line %d of task file", p.currentLine)
	}
	setting := parts[0]
	settingVal := bytes.TrimSpace(parts[1])

	if bytes.Equal(setting, []byte("commands")) {
		return p.parseCommandBlockStart(setting, settingVal)
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
					return fmt.Errorf("Cannot redeclare setting '%s'. Line %d", setting, p.currentLine)
				}
				f.SetString(string(settingVal))
			} else if f.Kind() == reflect.Int32 {
				if f.Int() > 0 {
					return fmt.Errorf("Cannot redeclare setting '%s'. Line %d", setting, p.currentLine)
				}

				i, err := strconv.Atoi(string(settingVal))
				if err != nil {
					return fmt.Errorf("Expected integer on line %d", p.currentLine)
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
	return fmt.Errorf("Invalid setting \"%s\". Line %d", setting, p.currentLine)
}

func (p *Parser) parseCommandBlockStart(cmd, opts []byte) error {
	if p.task.Commands == nil {
		p.task.Commands = make(map[string]*CommandBlock)
	}

	pieces := bytes.Split(opts, []byte(" "))
	name := ""
	settingsStartIndex := 0

	if !bytes.Contains(pieces[0], []byte("=")) {
		name = string(pieces[0])
		settingsStartIndex = 1
	}

	_, set := p.task.Commands[name]
	if set {
		return fmt.Errorf("%s block with name '%s' already exists. Line %d", cmd, opts, p.currentLine)
	}

	p.task.Commands[name] = &CommandBlock{
		Name: name,
	}

	if len(pieces) > 0 {
		for _, setting := range pieces[settingsStartIndex:] {
			parts := bytes.SplitN(setting, []byte("="), 2)
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
							return fmt.Errorf("Cannot redeclare setting '%s'. Line %d", parts[0], p.currentLine)
						}
						f.SetString(string(parts[1]))
					}
				}
			} else {
				return fmt.Errorf("Invalid block setting \"%s\". Line %d", parts[0], p.currentLine)
			}
		}
	}
	p.task.currentBlock = name
	p.runningMode = modeCommand
	return nil
}

func (p *Parser) parseCommandLine(line []byte) error {
	matches := wsRegex.FindSubmatch(line)
	if len(matches) == 0 {
		return p.parseLine(line)
	}
	sigWs := string(matches[0])
	current := p.task.Commands[p.task.currentBlock]

	if len(current.Commands) == 0 {
		p.currentSigWs = sigWs
	} else {
		if sigWs != p.currentSigWs {
			return fmt.Errorf("Command not in block, check indention. Line %d", p.currentLine)
		}
	}

	line = bytes.TrimSpace(line)
	current.Commands = append(current.Commands, string(line))
	return nil
}

func (p *Parser) parseDeviceLine(line []byte) error {
	matches := wsRegex.FindSubmatch(line)
	if len(matches) == 0 {
		return p.parseLine(line)
	}
	sigWs := string(matches[0])

	if len(p.task.Devices) == 0 {
		p.currentSigWs = sigWs
	} else {
		if sigWs != p.currentSigWs {
			return fmt.Errorf("Device not in block, check indention. Line %d", p.currentLine)
		}
	}

	line = bytes.TrimSpace(line)
	p.task.Devices = append(p.task.Devices, string(line))
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

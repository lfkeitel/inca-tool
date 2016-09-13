package devices

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
)

var (
	groupNameRegex   = regexp.MustCompile(`^\[([\w\- ]+?)\]`)
	lineSettingRegex = regexp.MustCompile(`([\w\-]+?) ?[=:] ?(?:([^"]\S+)|(?:"((?:[^\\"]|\\\\|\\")+)"))`)
)

func ParseFile(filename string) (*DeviceList, error) {
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return nil, fmt.Errorf("Inventory file does not exist: %s\n", filename)
	}

	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	return parse(file)
}

func ParseString(data string) (*DeviceList, error) {
	return parse(strings.NewReader(data))
}

func parse(reader io.Reader) (*DeviceList, error) {
	scanner := bufio.NewScanner(reader)
	scanner.Split(bufio.ScanLines)
	devices := &DeviceList{
		Groups:  make(map[string]*Group),
		Devices: make(map[string]*Device),
	}
	lineNum := 0
	currentGroup := ""

	for scanner.Scan() {
		line := scanner.Bytes()
		line = bytes.TrimSpace(line)
		lineNum++

		// Check for blank lines or comments
		if len(line) < 1 || line[0] == '#' {
			continue
		}

		// Start of group definition
		if line[0] == '[' {
			groupLine := groupNameRegex.FindAllSubmatch(line, -1)
			if len(groupLine) == 0 {
				return nil, fmt.Errorf("Error defining group on line %d\n", lineNum)
			}
			currentGroup = string(groupLine[0][1])
			// Check that group name doesn't conflict
			if _, exists := devices.Devices[currentGroup]; exists {
				return nil, fmt.Errorf("Can't define a group with the same name as a device. Line %d\n", lineNum)
			}
			if _, exists := devices.Groups[currentGroup]; exists {
				// If the group already exists, just set the current group and go on
				continue
			}
			// If the group doesn't exist, create a new group
			devices.Groups[currentGroup] = &Group{
				Name:     currentGroup,
				settings: getLineSettings(line),
				list:     devices,
			}
			continue
		}

		// The "global" group can only have key = value lines, no device definitions
		if currentGroup == "global" {
			settings := getLineSettings(line)
			for key, value := range settings {
				devices.Groups[currentGroup].settings[key] = value
			}
			continue
		}

		// Check for empty group
		if currentGroup == "" {
			return nil, fmt.Errorf("All devices must be inside a group. Line %d\n", lineNum)
		}

		splitLine := bytes.SplitN(line, []byte(" "), 2)
		deviceName := string(splitLine[0])

		// Add device
		if dev, exists := devices.Devices[deviceName]; exists {
			dev.Groups = append(dev.Groups, currentGroup)
			devices.Groups[currentGroup].Devices = append(devices.Groups[currentGroup].Devices, dev)
		} else {
			if _, exists := devices.Groups[deviceName]; exists {
				return nil, fmt.Errorf("Can't define a device with the same name as a group. Line %d\n", lineNum)
			}
			device := &Device{
				Name:     deviceName,
				settings: getLineSettings(line),
				Groups:   []string{currentGroup},
				list:     devices,
			}

			devices.Devices[deviceName] = device
			devices.Groups[currentGroup].Devices = append(devices.Groups[currentGroup].Devices, device)
		}
	}

	return devices, nil
}

func getLineSettings(line []byte) map[string]string {
	regLine := lineSettingRegex.FindAllSubmatch(line, -1)
	sets := make(map[string]string)
	for _, setting := range regLine {
		if len(setting) == 0 {
			continue
		}
		value := setting[2]
		if bytes.Equal(value, []byte("")) {
			value = setting[3]
		}
		sets[string(setting[1])] = string(value)
	}
	return sets
}

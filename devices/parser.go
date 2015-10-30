package devices

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"
)

var (
	currentGroup     = ""
	groupNameRegex   = regexp.MustCompile(`^\[([\w\- ]+?)\]`)
	lineSettingRegex = regexp.MustCompile(`([\w\-]+?) ?= ?(?:([^"]\S+)|(?:"((?:[^\\"]|\\\\|\\")+)"))`)
)

// ParseFile takes a filename and parses it into a DeviceList
func ParseFile(filename string) (*DeviceList, error) {
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return nil, fmt.Errorf("File does not exist: %s\n", filename)
	}

	listFile, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer listFile.Close()

	scanner := bufio.NewScanner(listFile)
	scanner.Split(bufio.ScanLines)
	devices := &DeviceList{
		Groups:  make(map[string]*Group),
		Devices: make(map[string]*Device),
	}
	lineNum := 0

	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line)
		lineNum++

		if len(line) < 1 || line[0] == '#' {
			continue
		}

		// Start of group definition
		if line[0] == '[' {
			groupLine := groupNameRegex.FindAllStringSubmatch(line, -1)
			if len(groupLine) == 0 {
				return nil, fmt.Errorf("Error defining group on line %d\n", lineNum)
			}
			currentGroup = groupLine[0][1]
			if _, exists := devices.Devices[currentGroup]; exists {
				return nil, fmt.Errorf("Can't define a group with the same name as a device. Line %d\n", lineNum)
			}
			devices.Groups[currentGroup] = &Group{
				Name:     currentGroup,
				settings: getLineSettings(line),
				list:     devices,
			}
			continue
		}

		if currentGroup == "global" {
			settings := getLineSettings(line)
			for key, value := range settings {
				devices.Groups[currentGroup].settings[key] = value
			}
			continue
		}

		if currentGroup == "" {
			return nil, fmt.Errorf("All devices must be inside a group. Line %d\n", lineNum)
		}

		splitLine := strings.Split(line, " ")
		deviceName := splitLine[0]

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

func getLineSettings(line string) map[string]string {
	regLine := lineSettingRegex.FindAllStringSubmatch(line, -1)
	sets := make(map[string]string)
	for _, setting := range regLine {
		if len(setting) == 0 {
			continue
		}
		value := setting[2]
		if setting[2] == "" {
			value = setting[3]
		}
		sets[setting[1]] = value
	}
	return sets
}

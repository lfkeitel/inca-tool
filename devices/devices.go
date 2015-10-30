package devices

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// Device represents a device
type Device struct {
	Name         string
	Address      string
	Manufacturer string
	Method       string
}

// The below types are in preparation for a new device list format

// DeviceList is a list of device groups
type DeviceList struct {
	Groups  map[string]*Group
	Devices map[string]*Device
}

// Group is a collection of devices
type Group struct {
	Name    string
	Devices []*Device
}

var (
	currentGroup = ""
)

// LoadDevices takes a filename and parses the file into a slice of Host objects
func LoadDevices(filename string) ([]Device, error) {
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return nil, fmt.Errorf("Devices file does not exist: %s\n", filename)
	}

	listFile, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer listFile.Close()

	scanner := bufio.NewScanner(listFile)
	scanner.Split(bufio.ScanLines)
	var hostList []Device
	lineNum := 0

	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line)
		lineNum++

		if len(line) < 1 || line[0] == '#' {
			continue
		}

		splitLine := strings.Split(line, "::")

		if len(splitLine) != 4 {
			fmt.Printf("Error on line %d in device configuration\n", lineNum)
			continue
		}

		device := Device{
			Name:         splitLine[0],
			Address:      splitLine[1],
			Manufacturer: splitLine[2],
			Method:       splitLine[3],
		}

		hostList = append(hostList, device)
	}

	return hostList, nil
}

// FilterDevices filters a slice of Host objects by their Manufacturer and Method
func FilterDevices(devices []Device, filter string) []Device {
	filters := strings.Split(filter, ":")
	man := filters[0]
	proto := filters[1]
	var hosts []Device

	for _, device := range devices {
		if man != "*" && device.Manufacturer != man {
			continue
		}

		if proto != "*" && device.Method != proto {
			continue
		}

		hosts = append(hosts, device)
	}

	return hosts
}

// LoadAndFilterDevices loads the devices from filename and filters them in one single function
func LoadAndFilterDevices(filename, filter string) ([]Device, error) {
	d, err := LoadDevices(filename)
	if err != nil {
		return nil, err
	}

	return FilterDevices(d, filter), nil
}

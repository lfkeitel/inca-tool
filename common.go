package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	parser "github.com/dragonrider23/inca-tool/taskfileparser"
)

func loadDevices(filename string) ([]host, error) {
	listFile, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer listFile.Close()

	scanner := bufio.NewScanner(listFile)
	scanner.Split(bufio.ScanLines)
	var hostList []host
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

		device := host{
			name:         splitLine[0],
			address:      splitLine[1],
			manufacturer: splitLine[2],
			method:       splitLine[3],
		}

		hostList = append(hostList, device)
	}

	return hostList, nil
}

func filterDevices(devices []host, filter string) []host {
	filters := strings.Split(filter, ":")
	man := filters[0]
	proto := filters[1]
	var hosts []host

	for _, device := range devices {
		if man != "*" && device.manufacturer != man {
			continue
		}

		if proto != "*" && device.method != proto {
			continue
		}

		hosts = append(hosts, device)
	}

	return hosts
}

func getArguments(host host, task *parser.TaskFile, eargs []string) []string {
	argList := make([]string, 6+len(eargs))
	argList[0] = host.method
	argList[1] = host.manufacturer
	argList[2] = host.address
	argList[3] = task.Username
	argList[4] = task.Password
	if task.EnablePassword != "" {
		argList[5] = task.EnablePassword
	} else {
		argList[5] = task.Password
	}

	for i, arg := range eargs {
		argList[i+6] = arg
	}
	return argList
}

package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	us "github.com/dragonrider23/utils/sync"
)

var (
	devicesFile    string
	filterArg      string
	argString      string
	scriptFilename string
	rUser          string
	rPass          string
	rEnablepw      string
	dryRun         bool
)

type host struct {
	name         string
	address      string
	manufacturer string
	method       string
}

type config struct {
	remoteUsername string
	remotePassword string
	enablePassword string
	concurrent     int32
}

func init() {
	flag.StringVar(&devicesFile, "d", "", "Path to an Inca v2 device definitions file.")
	flag.StringVar(&filterArg, "f", "*:*", "Filter in the form manufacturer:protocol")
	flag.StringVar(&argString, "a", "", "List of comma separated arguments to give to script")
	flag.StringVar(&scriptFilename, "s", "", "Script to run for each device")
	flag.StringVar(&rUser, "u", "", "Username to login to device")
	flag.StringVar(&rPass, "p", "", "Password to login to device")
	flag.StringVar(&rEnablepw, "e", "", "Username to enter Enable exec mode")
	flag.BoolVar(&dryRun, "dry", false, "Only list the devices that would be affected, doesn't run any scripts")
}

func main() {
	flag.Parse()

	if !checkArguments() {
		fmt.Println("Inca Tools requires a device file, script file, username, and password to run")
		os.Exit(1)
	}

	if _, err := os.Stat(scriptFilename); os.IsNotExist(err) {
		fmt.Printf("Script file does not exist: %s\n", scriptFilename)
		return
	}

	scriptFilename, _ = filepath.Abs(scriptFilename)

	if argString == "" {
		_, filename := filepath.Split(scriptFilename)
		base := strings.Split(filename, ".")
		pieces := strings.Split(base[0], "--")

		if len(pieces) == 2 {
			argString = pieces[1]
		}
	}

	fmt.Println(argString)

	hosts, err := loadDevices(devicesFile)
	if err != nil {
		fmt.Println("Error loading devices")
		os.Exit(1)
	}

	hosts = filterDevices(hosts, filterArg)

	if !dryRun {
		conf := config{
			remoteUsername: rUser,
			remotePassword: rPass,
			enablePassword: rEnablepw,
			concurrent:     300,
		}

		executeTask(hosts, conf)
	} else {
		fmt.Print("Dry Run\n\n")
		for _, host := range hosts {
			fmt.Print("---------\n")
			fmt.Printf("Name: %s\n", host.name)
			fmt.Printf("Hostname: %s\n", host.address)
			fmt.Printf("Manufacturer: %s\n", host.manufacturer)
			fmt.Printf("Protocol: %s\n", host.method)
		}
	}

	fmt.Printf("\nHosts touched: %d\n", len(hosts))
	fmt.Println("Task completed")
}

func checkArguments() bool {
	if devicesFile == "" ||
		scriptFilename == "" ||
		rUser == "" ||
		rPass == "" {
		return false
	}

	return true
}

func loadDevices(filename string) ([]host, error) {
	listFile, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer listFile.Close()

	scanner := bufio.NewScanner(listFile)
	scanner.Split(bufio.ScanLines)
	fmt.Println(scanner.Text())
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

func executeTask(hosts []host, conf config) error {
	var wg sync.WaitGroup
	lg := us.NewLimitGroup(conf.concurrent) // Used to enforce a maximum number of connections

	for _, host := range hosts {
		host := host
		args := getArguments(argString, host, conf)

		wg.Add(1)
		lg.Add(1)
		go func() {
			defer func() {
				wg.Done()
				lg.Done()
			}()
			scriptExecute(scriptFilename, args)
		}()

		lg.Wait()
	}

	wg.Wait()
	return nil
}

func getArguments(argStr string, host host, conf config) []string {
	args := strings.Split(argStr, ",")
	argList := make([]string, len(args))
	for i, a := range args {
		switch a {
		case "address":
			argList[i] = host.address
			break
		case "username":
			argList[i] = conf.remoteUsername
			break
		case "password":
			argList[i] = conf.remotePassword
			break
		case "enablepw":
			argList[i] = conf.enablePassword
			break
		case "protocol":
			argList[i] = host.method
			break
		case "brand":
			argList[i] = host.manufacturer
		}
	}
	return argList
}

func scriptExecute(sfn string, args []string) error {
	cmd := exec.Command(sfn, args...)
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		fmt.Println(fmt.Sprint(err) + ": " + stderr.String())
		return err
	}
	return nil
}

package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	parser "github.com/dragonrider23/inca-tool/taskfileparser"
)

var (
	dryRun        bool   // flag
	testTaskParse string // flag
	verbose       bool   // flag

	taskFilename string
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
	flag.BoolVar(&dryRun, "d", false, "Only list the devices that would be affected, doesn't run any scripts")
	flag.StringVar(&testTaskParse, "test", "", "Test a task file for validity")
	flag.BoolVar(&verbose, "v", false, "Enable verbose output")
}

func main() {
	start := time.Now()
	flag.Parse()

	if testTaskParse != "" {
		validateTaskFile(testTaskParse)
		os.Exit(0)
	}

	cliArgs := flag.Args()
	cliArgsc := len(cliArgs)

	if cliArgsc > 0 {
		if cliArgs[0] == "run" && cliArgsc >= 2 {
			taskFilename = cliArgs[1]
		} else {
			fmt.Println("Usage: it run [task file]")
			os.Exit(0)
		}
	} else {
		fmt.Println("Usage: it [command] [arguments]")
		os.Exit(0)
	}

	task, err := parser.Parse(taskFilename)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	if !checkArguments(task) {
		fmt.Println("Inca Tools requires a device file, script file, username, and password to run")
		os.Exit(1)
	}

	task.Script, _ = filepath.Abs(task.Script)

	hosts, err := loadDevices(task.DeviceList)
	if err != nil {
		fmt.Println("Error loading devices")
		os.Exit(1)
	}

	hosts = filterDevices(hosts, task.Filter)

	if !dryRun {
		var err error
		switch task.Mode {
		case "script":
			err = execScriptMode(hosts, task)
			break
		case "expect":
			err = execExpectMode(hosts, task)
			break
		case "commands":
			err = execCommandMode(hosts, task)
			break
		}

		if err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}
	} else {
		fmt.Print("Dry Run\n\n")
		for _, host := range hosts {
			fmt.Printf("Name: %s\n", host.name)
			fmt.Printf("Hostname: %s\n", host.address)
			fmt.Printf("Manufacturer: %s\n", host.manufacturer)
			fmt.Printf("Protocol: %s\n", host.method)
			fmt.Print("---------\n")
		}
	}

	fmt.Printf("\nHosts touched: %d\n", len(hosts))
	end := time.Since(start).String()
	fmt.Printf("Task completed in %s\n", end)
}

func checkArguments(task *parser.TaskFile) bool {
	if task.Mode == "" ||
		task.Username == "" ||
		task.Password == "" ||
		task.DeviceList == "" ||
		task.Filter == "" {
		return false
	}

	return true
}

func validateTaskFile(filename string) {
	task, err := parser.Parse(filename)
	if err != nil {
		fmt.Println(err.Error())
	} else {
		if verbose {
			fmt.Printf("%#v\n", task)
		}
		fmt.Println("The task file has no syntax errors.")
	}
}

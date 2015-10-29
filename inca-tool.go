package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	parser "github.com/dragonrider23/inca-tool/taskfileparser"
)

var (
	dryRun        bool   // flag
	testTaskParse string // flag
	verbose       bool   // flag
	debug         bool   // flag

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
	flag.BoolVar(&dryRun, "d", false, "Do everything up to but not including, actually running the script. Also lists affected devices")
	flag.StringVar(&testTaskParse, "test", "", "Test a task file for validity")
	flag.BoolVar(&verbose, "v", false, "Enable verbose output")
	flag.BoolVar(&debug, "debug", false, "Enable debug mode")
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

	hosts, err := loadDevices(task.DeviceList)
	if err != nil {
		fmt.Printf("Error loading devices: %s\n", err.Error())
		os.Exit(1)
	}

	hosts = filterDevices(hosts, task.Filter)

	fmt.Printf("Task Information:\n")
	fmt.Printf("  Name: %s\n", task.Name)
	fmt.Printf("  Description: %s\n", task.Description)
	fmt.Printf("  Author: %s\n", task.Author)
	fmt.Printf("  Last Changed: %s\n", task.Date)
	fmt.Printf("  Version: %s\n", task.Version)
	fmt.Printf("  Filter: %s\n\n", task.Filter)

	if err := execute(hosts, task); err != nil {
		fmt.Printf("Error executing task: %s\n", err.Error())
		os.Exit(1)
	}

	if dryRun {
		fmt.Print("Dry Run\n\n")
		for _, host := range hosts {
			fmt.Printf("Name: %s\n", host.name)
			fmt.Printf("Hostname: %s\n", host.address)
			fmt.Printf("Manufacturer: %s\n", host.manufacturer)
			fmt.Printf("Protocol: %s\n", host.method)
			fmt.Println("---------")
		}
	}

	fmt.Printf("\nHosts touched: %d\n", len(hosts))
	end := time.Since(start).String()
	fmt.Printf("Task completed in %s\n", end)
}

func checkArguments(task *parser.TaskFile) bool {
	if task.Username == "" ||
		task.Password == "" ||
		task.DeviceList == "" ||
		task.Filter == "" ||
		(task.Commands == nil &&
			task.Expects == nil) {
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
			fmt.Printf("Task: %+v\n\n", task)

			fmt.Print("----Task Command Blocks----")
			for _, c := range task.Commands {
				fmt.Printf("\nCommand block Name: %s\n", c.Name)
				fmt.Printf("Command block Type: %s\n", c.Type)
				fmt.Printf("Commands: %v\n", c.Commands)
				fmt.Println("---------------")
			}

			fmt.Print("\n----Task Expect Blocks----")
			for _, e := range task.Expects {
				fmt.Printf("\nExpect Block Name: %s\nLines: %v\n", e.Name, e.String)
				fmt.Println("---------------")
			}
		}
		fmt.Println("The task file has no syntax errors.")
	}
}

func execute(devices []host, task *parser.TaskFile) error {
	if _, ok := task.Commands["main"]; !ok {
		return fmt.Errorf("Main command block not found")
	}
	return executeCommand("main", task, devices, debug)
}

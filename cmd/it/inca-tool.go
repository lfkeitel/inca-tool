package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/lfkeitel/inca-tool/src/device"
	"github.com/lfkeitel/inca-tool/src/manager"
	parser "github.com/lfkeitel/inca-tool/src/task"
)

const (
	incaVersion = "0.3.0"
)

type varSlice map[string]string

func (v varSlice) String() string {
	return ""
}

func (v varSlice) Set(value string) error {
	if strings.TrimSpace(value) == "" {
		return nil
	}

	vars := strings.Split(value, ";")
	for _, val := range vars {
		sVar := strings.Split(val, ":")
		if len(sVar) != 2 {
			return fmt.Errorf("No value given for %s", val)
		}
		v[strings.TrimSpace(sVar[0])] = strings.TrimSpace(sVar[1])
	}
	return nil
}

var (
	dryRun        bool     // flag
	verbose       bool     // flag
	debug         bool     // flag
	inventoryFile string   // flag
	cliVars       varSlice // flag
)

func init() {
	cliVars = (varSlice)(make(map[string]string))
	flag.BoolVar(&dryRun, "r", false, "Do everything up to but not including, actually running the script. Also lists affected devices")
	flag.BoolVar(&verbose, "v", false, "Enable verbose output")
	flag.BoolVar(&debug, "d", false, "Enable debug mode")
	flag.StringVar(&inventoryFile, "i", "hosts", "Inventory file")
	flag.Var(cliVars, "var", "Extra variables")
}

func main() {
	start := time.Now()
	flag.Parse()

	// Set manager package settings
	manager.SetVerbose(verbose)
	manager.SetDebug(debug)
	manager.SetDryRun(dryRun)

	cliArgs := flag.Args()
	cliArgsc := len(cliArgs)

	// Determine what we're doing
	if cliArgsc == 0 {
		printUsage()
		os.Exit(0)
	}

	if err := checkDependencies(); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	command := cliArgs[0]
	if command == "run" && cliArgsc >= 2 { // Run a task file
		runCmd(cliArgs[1:])
	} else if command == "test" && cliArgsc >= 2 { // Test a task file for errors
		for _, file := range cliArgs[1:] {
			manager.ValidateTaskFile(file)
		}
	} else if command == "version" { // Show version info
		os.Exit(0)
	} else if command == "help" { // Show help info
		printUsage()
		os.Exit(0)
	} else if command == "dev" && cliArgsc == 2 { // Dev stuff
		devCmd(cliArgs[1])
		os.Exit(0)
	} else {
		printUsage()
		os.Exit(0)
	}

	end := time.Since(start).String()
	fmt.Printf("\nExecution completed in %s\n", end)
}

func printHeader() {
	fmt.Printf(`Inca Builder, version %s
Copyright (C) 2015 Onesimus Systems
License BSD 3-Clause: <https://opensource.org/licenses/BSD-3-Clause>

This is free software; you are free to change and redistribute it.
There is NO WARRANTY, to the extent permitted by law.
`, incaVersion)
}

func printUsage() {
	fmt.Printf(`Usage: %s [options] [command] [task1 [task2 [task3]...]]

Options:
	-d Enable debug output and functions
	-r Perform a dry run and list the affected hosts
	-v Enable verbose output

Commands:
	run Run the given task files
	test Test task files for errors
	version Show version information
	help Show this usage information
`, os.Args[0])
}

func checkDependencies() error {
	// Check to see if expect is installed
	expect, err := exec.Command("which", "expect").Output()
	if string(expect) == "" || err != nil {
		return fmt.Errorf("Expect doesn't appear to be installed.\n")
	}
	return nil
}

func runCmd(filepaths []string) {
	for _, file := range filepaths {
		// Parse the task file
		task, err := parser.ParseFile(file)
		if err != nil {
			fmt.Println(err.Error())
			continue
		}

		// Inventory from -i flag, overrides task file
		if inventoryFile != "" {
			task.Inventory = inventoryFile
		}
		// Set default if empty
		if task.Inventory == "" {
			task.Inventory = "inventory"
		}

		// Set variables given in the command line into the task
		for k, v := range cliVars {
			task.SetUserData(k, v)
		}
		manager.RunTask(task)
	}
}

func devCmd(filepath string) {
	d, err := device.ParseFile(filepath)
	if err != nil {
		fmt.Println(err.Error())
	} else {
		for _, group := range d.Groups {
			fmt.Printf("Group: %s\n", group.Name)
			fmt.Printf("Group Settings: %#v\n", group.GetSettings())
			for _, dev := range group.Devices {
				fmt.Printf("   DeviceName: %s\n", dev.Name)
				fmt.Printf("   Device Settings: %#v\n", dev.GetSettings())
			}
			fmt.Println("")
		}
	}
}

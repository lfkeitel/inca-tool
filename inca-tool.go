package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/dragonrider23/inca-tool/devices"
	"github.com/dragonrider23/inca-tool/taskmanager"
)

const (
	incaVersion = "0.3.0"
)

var (
	dryRun  bool // flag
	verbose bool // flag
	debug   bool // flag
)

func init() {
	flag.BoolVar(&dryRun, "dry", false, "Do everything up to but not including, actually running the script. Also lists affected devices")
	flag.BoolVar(&verbose, "v", false, "Enable verbose output")
	flag.BoolVar(&debug, "d", false, "Enable debug mode")
}

func main() {
	printHeader()
	start := time.Now()
	flag.Parse()

	// Set taskmanager package settings
	taskmanager.SetVerbose(verbose)
	taskmanager.SetDebug(debug)
	taskmanager.SetDryRun(dryRun)

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
		for _, file := range cliArgs[1:] {
			taskmanager.RunTaskFile(file)
		}
	} else if command == "test" && cliArgsc >= 2 { // Test a task file for errors
		for _, file := range cliArgs[1:] {
			taskmanager.ValidateTaskFile(file)
		}
	} else if command == "version" { // Show version info
		os.Exit(0)
	} else if command == "help" { // Show help info
		printUsage()
		os.Exit(0)
	} else if command == "dev" && cliArgsc == 2 { // Dev stuff
		d, err := devices.ParseFile(cliArgs[1])
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
	-dry Perform a dry run and list the affected hosts
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

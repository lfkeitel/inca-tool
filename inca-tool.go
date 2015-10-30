package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/dragonrider23/inca-tool/common"
	"github.com/dragonrider23/inca-tool/parser"
	"github.com/dragonrider23/inca-tool/scripts"
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

	cliArgs := flag.Args()
	cliArgsc := len(cliArgs)

	// Determine what we're doing
	if cliArgsc > 0 {
		command := cliArgs[0]
		if command == "run" && cliArgsc >= 2 { // Run a task file
			for _, file := range cliArgs[1:] {
				commandRun(file)
			}
		} else if command == "test" && cliArgsc >= 2 { // Test a task file for errors
			for _, file := range cliArgs[1:] {
				validateTaskFile(file)
			}
		} else if command == "version" {
			os.Exit(0)
		} else if command == "help" {
			printUsage()
			os.Exit(0)
		} else {
			printUsage()
			os.Exit(0)
		}
	} else {
		printUsage()
		os.Exit(0)
	}

	end := time.Since(start).String()
	fmt.Printf("\nExecution completed in %s\n", end)
}

func printHeader() {
	fmt.Println(`Inca Builder, version 0.1.0
Copyright (C) 2015 Onesimus Systems
License BSD 3-Clause: <https://opensource.org/licenses/BSD-3-Clause>

This is free software; you are free to change and redistribute it.
There is NO WARRANTY, to the extent permitted by law.
`)
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

func commandRun(taskfile string) {
	// Set scripts package settings
	scripts.SetVerbose(verbose)
	scripts.SetDebug(debug)
	scripts.SetDryRun(dryRun)

	// Parse the task file
	task, err := parser.Parse(taskfile)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	// Load and filter devices
	hosts, err := common.LoadAndFilterDevices(task.DeviceList, task.Filter)
	if err != nil {
		fmt.Printf("Error loading devices: %s\n", err.Error())
		os.Exit(1)
	}

	// Print some task information
	fmt.Printf("Task Information:\n")
	fmt.Printf("  Name: %s\n", task.Name)
	fmt.Printf("  Description: %s\n", task.Description)
	fmt.Printf("  Author: %s\n", task.Author)
	fmt.Printf("  Last Changed: %s\n", task.Date)
	fmt.Printf("  Version: %s\n", task.Version)
	fmt.Printf("  Filter: %s\n\n", task.Filter)

	// If no hosts will be affected, exit
	if len(hosts) == 0 {
		fmt.Println("Due to filtering, no devices would be affected. Exiting.")
		return
	}

	// Compile the script text
	text, err := parser.CompileCommandText("main", task)
	if err != nil {
		if parser.IsScriptRun(err) {
			if err := scripts.ProcessScriptCommand(text, task, hosts); err != nil {
				fmt.Printf("Error executing task: %s\n", err.Error())
			}
			os.Exit(1)
		}
		fmt.Printf("Error compiling script: %s\n", err.Error())
		os.Exit(1)
	}

	// Get the template file
	templateFile := "templates/" + task.Commands["main"].Template + "-inca-template.tmpl"
	if _, err := os.Stat(templateFile); os.IsNotExist(err) {
		fmt.Printf("Template not found: %s\n", task.Commands["main"].Template)
		os.Exit(1)
	}

	// Generate an executable script file
	scriptFilename, err := scripts.GenerateScriptFile(templateFile, text)
	if err != nil {
		fmt.Printf("Error generating script: %s\n", err.Error())
		os.Exit(1)
	}

	if debug {
		fmt.Printf("Script: %s\n", scriptFilename)
	}

	// Execute the script (the dry run setting will stop before actual execution)
	err = scripts.Execute(hosts, task, scriptFilename, nil)
	if !debug {
		os.Remove(scriptFilename)
	}

	if err != nil {
		fmt.Printf("Error executing task: %s\n", err.Error())
		os.Exit(1)
	}

	if dryRun {
		fmt.Print("Dry Run\n\n")
		for _, host := range hosts {
			fmt.Printf("Name: %s\n", host.Name)
			fmt.Printf("Hostname: %s\n", host.Address)
			fmt.Printf("Manufacturer: %s\n", host.Manufacturer)
			fmt.Printf("Protocol: %s\n", host.Method)
			fmt.Println("---------")
		}
	}

	fmt.Printf("\nHosts touched: %d\n", len(hosts))
}

func validateTaskFile(filename string) {
	task, err := parser.Parse(filename)
	if err != nil {
		fmt.Printf("\nErrors found in \"%s\"\n", filename)
		fmt.Printf("   %s\n", err.Error())
		return
	}

	// Compile the script text
	_, err = parser.CompileCommandText("main", task)
	if err != nil {
		if !parser.IsScriptRun(err) {
			fmt.Printf("\nErrors found in \"%s\"\n", filename)
			fmt.Printf("   %s\n", err.Error())
			return
		}
	}

	if verbose {
		fmt.Printf("\nVerbose Information for Task \"%s\"\n", task.Name)
		fmt.Printf("  Name: %s\n", task.Name)
		fmt.Printf("  Description: %s\n", task.Description)
		fmt.Printf("  Author: %s\n", task.Author)
		fmt.Printf("  Date: %s\n", task.Date)
		fmt.Printf("  Version: %s\n\n", task.Version)

		fmt.Printf("  Concurrent Devices: %d\n", task.Concurrent)
		fmt.Printf("  Devices File: %s\n", task.DeviceList)
		fmt.Printf("  Filter: %s\n\n", task.Filter)

		fmt.Printf("  Username: %s\n", task.Username)
		fmt.Printf("  Password: %s\n", task.Password)
		fmt.Printf("  Enable Password: %s\n\n", task.EnablePassword)

		fmt.Print("  ----Task Command Blocks----\n")
		for _, c := range task.Commands {
			fmt.Printf("  Command block Name: %s\n", c.Name)
			fmt.Printf("  Command block Type: %s\n", c.Type)
			fmt.Printf("  Command block Template: %s\n", c.Template)
			fmt.Printf("  Commands:\n")
			for _, cmd := range c.Commands {
				fmt.Printf("     %s\n", cmd)
			}
			fmt.Println("  ---------------")
		}
	}
	fmt.Printf("The task named \"%s\" has no syntax errors.\n", task.Name)
}

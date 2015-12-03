package taskmanager

import (
	"fmt"
	"os"

	"github.com/dragonrider23/inca-tool/devices"
	"github.com/dragonrider23/inca-tool/parser"
	"github.com/dragonrider23/inca-tool/scripts"
)

var (
	verbose = false
	dryRun  = false
	debug   = false
)

// SetVerbose enables or disables verbose output
func SetVerbose(setting bool) {
	verbose = setting
}

// SetDryRun enables or disables actually executing the script
func SetDryRun(setting bool) {
	dryRun = setting
}

// SetDebug enables or disables debug output
func SetDebug(setting bool) {
	debug = setting
}

func RunTaskFile(taskfile string) {
	// Set scripts package settings
	scripts.SetVerbose(verbose)
	scripts.SetDebug(debug)
	scripts.SetDryRun(dryRun)

	// Parse the task file
	fileParser := parser.NewParser()
	task, err := fileParser.ParseFile(taskfile)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	// If no devices were given, print err and exit
	if len(task.Devices) == 0 {
		fmt.Println("No devices were specified in the task file. Exiting.")
		return
	}

	// Load and filter devices
	deviceList, err := devices.ParseFile(task.DeviceList)
	if err != nil {
		fmt.Printf("Error loading devices: %s\n", err.Error())
		os.Exit(1)
	}

	deviceList, err = devices.Filter(deviceList, task.Devices)
	if err != nil {
		fmt.Printf("Error: %s\n", err.Error())
		os.Exit(1)
	}

	// If no devices will be affected, exit
	if len(deviceList.Devices) == 0 {
		fmt.Println("Due to filtering, no devices would be affected. Exiting.")
		return
	}

	// Print some task information
	fmt.Printf("Task Information:\n")
	fmt.Printf("  Name: %s\n", task.Name)
	fmt.Printf("  Description: %s\n", task.Description)
	fmt.Printf("  Author: %s\n", task.Author)
	fmt.Printf("  Last Changed: %s\n", task.Date)
	fmt.Printf("  Version: %s\n", task.Version)

	// Compile the script text
	text, err := parser.CompileCommandText("main", task)
	if err != nil {
		if parser.IsScriptRun(err) {
			// Run straight script file if prompted
			if err := scripts.ProcessScriptCommand(text, task, deviceList); err != nil {
				fmt.Printf("Error executing task: %s\n", err.Error())
			}
			os.Exit(1)
		}
		fmt.Printf("Error compiling script: %s\n", err.Error())
		os.Exit(1)
	}

	// Get the template file
	template := task.Template
	if template == "" {
		template = "expect"
	}
	templateFile := "templates/" + template + "-template.tmpl"
	if _, err := os.Stat(templateFile); os.IsNotExist(err) {
		fmt.Printf("Template not found: %s\n", template)
		os.Exit(1)
	}

	// Generate an executable script file
	scriptFilename, err := scripts.GenerateBaseScriptFile(templateFile, text)
	if err != nil {
		fmt.Printf("Error generating script: %s\n", err.Error())
		os.Exit(1)
	}

	if debug {
		fmt.Printf("Script: %s\n", scriptFilename)
	}

	// Execute the script (the dry run setting will stop before actual execution)
	err = scripts.Execute(deviceList, task, scriptFilename, nil)
	if err != nil {
		fmt.Printf("Error executing task: %s\n", err.Error())
		os.Exit(1)
	}

	if !debug {
		// Remove base script file
		os.Remove(scriptFilename)
	}

	if dryRun {
		fmt.Print("\nDry Run\n\n")
		for _, host := range deviceList.Devices {
			fmt.Printf("Hostname: %s\n", host.Name)
			fmt.Printf("Address: %s\n", host.GetSetting("address"))
			proto := host.GetSetting("protocol")
			if proto == "" {
				proto = "ssh"
			}
			fmt.Printf("Protocol: %s\n", proto)
			fmt.Println("---------")
		}
	}

	fmt.Printf("\nHosts touched: %d\n", len(deviceList.Devices))
}

func ValidateTaskFile(filename string) {
	fileParser := parser.NewParser()
	task, err := fileParser.ParseFile(filename)
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
		fmt.Printf("  Template: %s\n", task.Template)
		fmt.Printf("  Devices File: %s\n\n", task.DeviceList)

		fmt.Print("  ----Task Device Block----\n")
		for _, d := range task.Devices {
			fmt.Printf("  Device(s): %s\n", d)
		}

		fmt.Print("\n  ----Task Command Blocks----\n")
		for _, c := range task.Commands {
			fmt.Printf("  Command block Name: %s\n", c.Name)
			fmt.Printf("  Command block Type: %s\n", c.Type)
			fmt.Printf("  Commands:\n")
			for _, cmd := range c.Commands {
				fmt.Printf("     %s\n", cmd)
			}
			fmt.Println("  ---------------")
		}
	}
	fmt.Printf("The task named \"%s\" has no syntax errors.\n", task.Name)
}
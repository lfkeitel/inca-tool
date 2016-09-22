package manager

import (
	"fmt"
	"os"
	"time"

	"github.com/lfkeitel/inca-tool/src/device"
	"github.com/lfkeitel/inca-tool/src/script"
	"github.com/lfkeitel/inca-tool/src/task"
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

func RunTask(t *task.Task) {
	// Set scripts package settings
	script.SetVerbose(verbose)
	script.SetDebug(debug)
	script.SetDryRun(dryRun)

	// Ensure a temporary directory is available
	os.RemoveAll("tmp")
	os.Mkdir("tmp", 0755)

	// Print a header
	fmt.Printf("Running task %s @ %s\n", t.GetMetadata("name"), time.Now().String())

	// If no devices were given, print err and exit
	if len(t.Devices) == 0 {
		fmt.Println("No devices were given in the task file. Exiting.")
		return
	}

	// Load and filter devices
	if verbose {
		fmt.Printf("Loading inventory from %s\n", t.Inventory)
	}
	deviceList, err := device.ParseFile(t.Inventory)
	if err != nil {
		fmt.Printf("Error loading devices: %s\n", err.Error())
		return
	}

	deviceList, err = device.Filter(deviceList, t.Devices)
	if err != nil {
		fmt.Printf("Error: %s\n", err.Error())
		return
	}

	// If no devices will be affected, exit
	if len(deviceList.Devices) == 0 {
		fmt.Println("No devices match running task. Exiting.")
		return
	}

	// Create a script based on the task
	taskScript, err := script.GenerateScript(t)
	if err != nil {
		fmt.Printf("Error generating script: %s\n", err.Error())
		return
	}

	// Execute the script
	if err := script.Execute(taskScript, deviceList); err != nil {
		fmt.Printf("Error executing task: %s\n", err.Error())
		return
	}

	if !debug {
		taskScript.Clean()
	}

	// Print verbose dry run data
	if dryRun {
		fmt.Print("\nDry Run\n\n")

		if verbose {
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
			fmt.Print("\n")
		}
	}

	fmt.Printf("Hosts touched: %d\n", len(deviceList.Devices))
}

func ValidateTaskFile(filename string) {
	t, err := task.ParseFile(filename)
	if err != nil {
		fmt.Printf("\nErrors found in \"%s\"\n", filename)
		fmt.Printf("   %s\n", err.Error())
		return
	}

	// Compile the script text
	_, err = script.GenerateScript(t)
	if err != nil {
		fmt.Printf("\nErrors found in \"%s\"\n", filename)
		fmt.Printf("   %s\n", err.Error())
		return
	}

	if verbose {
		fmt.Printf("\nInformation for Task \"%s\"\n", t.GetMetadata("name"))
		fmt.Printf("  Description: %s\n", t.GetMetadata("description"))
		fmt.Printf("  Author: %s\n", t.GetMetadata("author"))
		fmt.Printf("  Last Changed: %s\n", t.GetMetadata("date"))
		fmt.Printf("  Version: %s\n", t.GetMetadata("version"))

		fmt.Printf("  Concurrent Devices: %d\n", t.Concurrent)
		fmt.Printf("  Template: %s\n", t.Template)
		fmt.Printf("  Inventory File: %s\n\n", t.Inventory)

		fmt.Print("  ----Custom Data----\n")
		for k, v := range t.Metadata {
			if k[0] != '_' {
				continue
			}
			fmt.Printf("  %s: %s\n", k[1:], v)
		}

		fmt.Print("\n  ----Task Device Block----\n")
		for _, d := range t.Devices {
			fmt.Printf("  Device(s): %s\n", d)
		}

		fmt.Print("\n  ----Task Command Blocks----\n")
		for _, c := range t.Commands {
			fmt.Printf("  Command block Name: %s\n", c.Name)
			fmt.Printf("  Command block Type: %s\n", c.Type)
			fmt.Printf("  Commands:\n")
			for _, cmd := range c.Commands {
				fmt.Printf("     %s\n", cmd)
			}
			fmt.Println("  ---------------")
		}
	}
	fmt.Printf("The task named \"%s\" has no syntax errors.\n", t.GetMetadata("name"))
}

package script

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/lfkeitel/inca-tool/src/device"

	us "github.com/lfkeitel/utils/sync"
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

func Execute(s *Script, hosts *device.DeviceList) error {
	fmt.Println("---------------------------------")

	if s.dontProcess {
		splitCmd := strings.SplitN(s.script, " ", 2)
		fmt.Printf("Running script %s\n", splitCmd[0])
		output := s.task.Output

		if output != "" {
			output = filepath.Join(output, time.Now().Format("20060102-15:04:05"), s.task.GetMetadata("name"), ".out")
		}

		if len(splitCmd) == 1 {
			return executeFile(splitCmd[0], "", output)
		}
		return executeFile(splitCmd[0], splitCmd[1], output)
	}

	// Wait group to enforce maximum concurrent hosts
	lg := us.NewLimitGroup(s.task.Concurrent)

	// For every host
	for _, host := range hosts.Devices {
		// Get variables
		vars := getHostVariables(host)
		fmt.Printf("Configuring host %s (%s)\n", host.Name, vars["hostname"])

		// Generate a host specific script file
		hostScript := fmt.Sprintf("%s-%s.sh", s.script, host.Name)
		err := copyFileContents(s.script, hostScript)
		if err != nil {
			fmt.Printf("Error configuring host %s: %s\n", host.Name, err.Error())
			continue
		}
		if err := insertVariables(hostScript, vars); err != nil {
			return err
		}

		if debug && verbose {
			fmt.Println("Script Variables:")
			for i, v := range vars {
				fmt.Printf("  %s: %s\n", i, v)
			}
		}

		// The magic, set off a goroutine to execute the script
		lg.Add(1)
		go func(script, name, address, output string) {
			defer func() {
				lg.Done()
			}()

			if output != "" {
				output = filepath.Join(output, time.Now().Format("20060102-15:04:05-")+name+".out")
			}

			executeFile(script, "", output)
			fmt.Printf("Finished configuring host %s (%s)\n", name, address)
			if !debug {
				// Remove host specific script file
				os.Remove(script)
			}
		}(hostScript, host.Name, vars["hostname"], s.task.Output)
		// Wait for the next available host execution slot
		lg.Wait()
	}
	// Wait for everybody
	lg.WaitAll()
	return nil
}

func executeFile(sfn, args, outputFile string) error {
	if dryRun {
		return nil
	}

	cmd := exec.Command(sfn, args)
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		fmt.Printf("%s: %s\n", err, stderr.String())
		if debug {
			fmt.Println(out.String())
		}
		return err
	}

	fmt.Println(outputFile)
	if outputFile != "" {
		if err := ioutil.WriteFile(outputFile, out.Bytes(), 0644); err != nil {
			fmt.Println(err.Error())
		}
	}
	return nil
}

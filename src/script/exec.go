package script

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"

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
	if s.dontProcess {
		splitCmd := strings.SplitN(s.script, " ", 2)
		if len(splitCmd) == 1 {
			return executeFile(splitCmd[0], "")
		}
		return executeFile(splitCmd[0], splitCmd[1])
	}

	// Wait group for all hosts
	var wg sync.WaitGroup
	// Wait group to enforce maximum concurrent hosts
	lg := us.NewLimitGroup(s.task.Concurrent)

	// For every host
	for _, host := range hosts.Devices {
		// Get variables
		vars := getHostVariables(host)
		if verbose {
			fmt.Printf("Configuring host %s (%s)\n", host.Name, vars["hostname"])
		}

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
		wg.Add(1)
		lg.Add(1)
		go func(script, name, address string) {
			defer func() {
				wg.Done()
				lg.Done()
			}()
			executeFile(script, "")
			if verbose {
				fmt.Printf("Finished configuring host %s (%s)\n", name, address)
			}
			if !debug {
				// Remove host specific script file
				os.Remove(script)
			}
		}(hostScript, host.Name, vars["hostname"])
		// Wait for the next available host execution slot
		lg.Wait()
	}
	// Wait for everybody
	wg.Wait()
	return nil
}

func executeFile(sfn string, args string) error {
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
	return nil
}

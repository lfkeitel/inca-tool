package scripts

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/dragonrider23/inca-tool/devices"
	"github.com/dragonrider23/inca-tool/parser"

	us "github.com/dragonrider23/utils/sync"
)

var (
	verbose = false
	dryRun  = false
	debug   = false
)

// Execute script on devices based on the task file and extra arguments eargs.
func Execute(devices *devices.DeviceList, task *parser.TaskFile, script string, eargs []string) error {
	// Make sure base script exists
	if _, err := os.Stat(script); os.IsNotExist(err) {
		return fmt.Errorf("Script file does not exist: %s\n", script)
	}
	// Run task
	return runTask(devices, task, script, eargs)
}

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

// ProcessScriptCommand processes an _s special command
func ProcessScriptCommand(cmd string, task *parser.TaskFile, devices *devices.DeviceList) error {
	// Separate the filename from the arguments
	cmdPieces := strings.Split(cmd, "--")
	// Make sure we have enough pieces
	if cmdPieces[0] == "" {
		return fmt.Errorf("'_s' must have a filename")
	}
	// Get the absolute filepath for safety
	script, err := filepath.Abs(strings.TrimSpace(cmdPieces[0]))
	if err != nil {
		return err
	}
	// Build the argument list
	var args []string
	if len(cmdPieces) > 1 {
		args = strings.Split(cmdPieces[1], ";")
		for i := range args {
			args[i] = strings.TrimSpace(args[i])
		}
	}

	return Execute(devices, task, script, args)
}

// GenerateBaseScriptFile generates a script based on the template and data given. It returns the path to the script
func GenerateBaseScriptFile(template string, data string) (string, error) {
	// Generate the base script filename
	tmpFilename := "tmp/builtBaseScript-" + strconv.FormatInt(time.Now().UnixNano(), 10)
	// Copy the template to the base file
	err := copyFileContents(template, tmpFilename)
	if err != nil {
		return "", err
	}
	// Insert the main section
	vars := map[string]string{"main": data}
	if err := insertVariables(tmpFilename, vars); err != nil {
		return "", err
	}
	// Return the filename for the base script
	return tmpFilename, nil
}

func runTask(hosts *devices.DeviceList, task *parser.TaskFile, baseScript string, eargs []string) error {
	// Wait group for all hosts
	var wg sync.WaitGroup
	// Wait group to enforce maximum concurrent hosts
	lg := us.NewLimitGroup(task.Concurrent)

	// For every host
	for _, host := range hosts.Devices {
		// Get variables
		vars := getVariables(host, task)
		if verbose {
			fmt.Printf("Configuring host %s (%s)\n", host.Name, vars["address"])
		}

		// Generate a host specific script file
		hostScript := fmt.Sprintf("%s-%s.sh", baseScript, host.Name)
		err := copyFileContents(baseScript, hostScript)
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
			runScript(script, eargs)
			if verbose {
				fmt.Printf("Finished configuring host %s (%s)\n", name, address)
			}
			if !debug {
				// Remove host specific script file
				os.Remove(script)
			}
		}(hostScript, host.Name, vars["address"])
		// Wait for the next available host execution slot
		lg.Wait()
	}
	// Wait for everybody
	wg.Wait()
	return nil
}

func runScript(sfn string, args []string) error {
	if dryRun {
		return nil
	}

	cmd := exec.Command(sfn, args...)
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		fmt.Println(fmt.Sprint(err) + ": " + stderr.String())
		if debug {
			fmt.Println(out.String())
		}
		return err
	}
	return nil
}

func copyFileContents(src, dst string) error {
	var err error

	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.OpenFile(dst, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0755)
	if err != nil {
		return err
	}

	defer func() {
		err = out.Close()
	}()

	if _, err = io.Copy(out, in); err != nil {
		return err
	}
	return out.Sync()
}

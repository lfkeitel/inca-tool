package scripts

import (
	"bytes"
	"fmt"
	"io/ioutil"
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
	if _, err := os.Stat(script); os.IsNotExist(err) {
		return fmt.Errorf("Script file does not exist: %s\n", script)
	}

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
	cmdPieces := strings.Split(cmd, "--")
	if cmdPieces[0] == "" {
		return fmt.Errorf("'_s' must have a filename")
	}
	script, err := filepath.Abs(strings.TrimSpace(cmdPieces[0]))
	if err != nil {
		return err
	}

	var args []string
	if len(cmdPieces) > 1 {
		args = strings.Split(cmdPieces[1], ";")
		for i := range args {
			args[i] = strings.TrimSpace(args[i])
		}
	}

	return Execute(devices, task, script, args)
}

// GenerateScriptFile generates a script based on the template and data given. It returns the path to the script
func GenerateScriptFile(template string, data string) (string, error) {
	file, err := ioutil.ReadFile(template)
	if err != nil {
		return "", err
	}

	generated := strings.Replace(string(file), "# {{content}}", data, -1)
	tmpFilename := "tmp/builtScript-" + strconv.FormatInt(time.Now().UnixNano(), 10)
	if err := ioutil.WriteFile(tmpFilename, []byte(generated), 0744); err != nil {
		return "", err
	}
	return tmpFilename, nil
}

func runTask(hosts *devices.DeviceList, task *parser.TaskFile, script string, eargs []string) error {
	var wg sync.WaitGroup
	lg := us.NewLimitGroup(task.Concurrent) // Used to enforce a maximum number of connections

	for _, host := range hosts.Devices {
		host := host
		if verbose {
			fmt.Printf("Configuring host %s\n", host.GetSetting("address"))
		}
		args := getArguments(host, task, eargs)

		if debug {
			fmt.Println("Script Arguments:")
			for i, arg := range args {
				fmt.Printf("  %d: %s\n", i, arg)
			}
		}

		wg.Add(1)
		lg.Add(1)
		go func() {
			defer func() {
				wg.Done()
				lg.Done()
			}()
			runScript(script, args)
			if verbose {
				fmt.Printf("Finished configuring host %s\n", host.GetSetting("address"))
			}
		}()

		lg.Wait()
	}

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
		return err
	}
	return nil
}

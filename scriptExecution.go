package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"sync"

	parser "github.com/dragonrider23/inca-tool/taskfileparser"

	us "github.com/dragonrider23/utils/sync"
)

func execScript(devices []host, task *parser.TaskFile, script string, args []string) error {
	if _, err := os.Stat(script); os.IsNotExist(err) {
		return fmt.Errorf("Script file does not exist: %s\n", script)
	}

	return executeTask(devices, task, script, args)
}

func executeTask(hosts []host, task *parser.TaskFile, script string, eargs []string) error {
	var wg sync.WaitGroup
	concurrent := task.Concurrent
	if concurrent <= 0 {
		concurrent = 300
	}
	lg := us.NewLimitGroup(concurrent) // Used to enforce a maximum number of connections

	for _, host := range hosts {
		host := host
		if verbose {
			fmt.Printf("Configuring host %s\n", host.address)
		}
		args := getArguments(host, task, eargs)

		wg.Add(1)
		lg.Add(1)
		go func() {
			defer func() {
				wg.Done()
				lg.Done()
			}()
			scriptExecute(script, args)
			if verbose {
				fmt.Printf("Finished configuring host %s\n", host.address)
			}
		}()

		lg.Wait()
	}

	wg.Wait()
	return nil
}

func scriptExecute(sfn string, args []string) error {
	if debug {
		fmt.Printf("Args: %#v\n", args)
	}

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

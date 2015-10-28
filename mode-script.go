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

func execScriptMode(devices []host, task *parser.TaskFile) error {
	if _, err := os.Stat(task.Script); os.IsNotExist(err) {
		return fmt.Errorf("Script file does not exist: %s\n", task.Script)
	}

	return executeTask(devices, task)
}

func executeTask(hosts []host, task *parser.TaskFile) error {
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
		args := getArguments(host, task)

		wg.Add(1)
		lg.Add(1)
		go func() {
			defer func() {
				wg.Done()
				lg.Done()
			}()
			scriptExecute(task.Script, args)
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

package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	parser "github.com/dragonrider23/inca-tool/taskfileparser"
)

func executeCommand(entry string, task *parser.TaskFile, devices []host, preserveTemp bool) error {
	main := task.Commands[entry]
	cmdStr, err := generateScriptText(entry, task)
	if err != nil {
		if err.Error() == "scriptrun" {
			return processScriptCommand(cmdStr, task, devices)
		}
		return err
	}

	templateFile := "scripts/" + main.Template + "-inca-template.tmpl"
	if _, err := os.Stat(templateFile); os.IsNotExist(err) {
		return fmt.Errorf("Template not found: %s\n", main.Template)
	}

	scriptFilename, err := generateScriptFile(templateFile, cmdStr)

	if err != nil {
		return err
	}
	scriptFilename, _ = filepath.Abs(scriptFilename)

	if debug {
		fmt.Printf("Script: %s\n", scriptFilename)
	}

	err = execScript(devices, task, scriptFilename, nil)
	if !preserveTemp {
		os.Remove(scriptFilename)
	}
	return err
}

func generateScriptText(block string, task *parser.TaskFile) (string, error) {
	main := task.Commands[block]
	var cmdStr string

	for _, cmd := range main.Commands {
		switch cmd[:2] {
		case "_s":
			return cmd[3:], errors.New("scriptrun")
		case "_c":
			commandBlock := cmd[3:]
			if commandBlock == block {
				return "", fmt.Errorf("Cannot include self in command block '%s'\n", commandBlock)
			}

			_, ok := task.Commands[commandBlock]
			if !ok {
				return "", fmt.Errorf("Command block not declared '%s'\n", commandBlock)
			}

			include, err := generateScriptText(commandBlock, task)
			if err != nil {
				return "", err
			}

			cmdStr += include
			break
		case "_e":
			expectName := cmd[3:]
			text, ok := task.Expects[expectName]
			if !ok {
				return "", fmt.Errorf("Expect block not declared '%s'\n", expectName)
			}
			cmdStr += text.String + "\n"
			break
		case "_b":
			builtinName := cmd[3:]
			text := getBuiltinCodeBlock(builtinName)
			if text == "" {
				return "", fmt.Errorf("Builtin block '%s' not found\n", builtinName)
			}
			cmdStr += text
			break
		default:
			cmd = strings.Replace(cmd, "\"", "\\\"", -1)
			cmdStr += fmt.Sprintf("send \"%s\\n\"\n", cmd)
			cmdStr += "expect \"#\"\n"
		}
	}

	return cmdStr, nil
}

func getBuiltinCodeBlock(blockName string) string {
	text, _ := builtinBlocks[blockName]
	return text
}

func generateScriptFile(template string, data string) (string, error) {
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

func processScriptCommand(cmd string, task *parser.TaskFile, devices []host) error {
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

	return execScript(devices, task, script, args)
}

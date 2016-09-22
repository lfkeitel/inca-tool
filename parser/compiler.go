package parser

import (
	"errors"
	"fmt"
	"strings"
)

var (
	errScriptRunException = errors.New("scriptrun")
)

// CompileCommandText generates a script based on the parameters of the task.
func CompileCommandText(entry string, task *TaskFile) (string, error) {
	// Check command block exists
	_, ok := task.Commands[entry]
	if !ok {
		return "", fmt.Errorf("Command block \"%s\" not declared\n", entry)
	}

	cmdStr, err := generateScriptText(entry, task)
	if err != nil {
		switch err.Error() {
		case "scriptrun":
			return cmdStr, errScriptRunException
		}
	}
	return cmdStr, err
}

// IsScriptRun returns if the error means a script should be ran
func IsScriptRun(err error) bool {
	return err == errScriptRunException
}

func generateScriptText(block string, task *TaskFile) (string, error) {
	main := task.Commands[block]
	cmdStr := ""
	prompt := task.Prompt
	if prompt == "" {
		prompt = "#"
	}

	for _, cmd := range main.Commands {
		switch cmd[:3] {
		case "_s ": // Include and run a script file
			return cmd[3:], errors.New("scriptrun")
		case "_c ": // Include another command block
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
		case "_b ": // Include a builtin command block
			builtinName := cmd[3:]
			if builtinName == "nil" {
				return cmdStr, nil
			}
			text := getBuiltinCodeBlock(builtinName)
			if text == "" {
				return "", fmt.Errorf("Builtin block '%s' not found\n", builtinName)
			}
			cmdStr += text
			break
		default:
			if cmd[0] == '_' {
				return "", fmt.Errorf("Command line cannot start with \"_\": %s\n", cmd)
			}
			switch main.Type {
			case "raw":
				cmdStr += cmd + "\n"
				break
			case "expect":
			default: // Wrap command lines with expect's send command and prompt
				cmd = strings.Replace(cmd, "\"", "\\\"", -1)
				cmdStr += fmt.Sprintf("send \"%s\\n\"\n", cmd)
				cmdStr += fmt.Sprintf("expect \"%s\"\n", prompt)
			}
		}
	}

	return cmdStr, nil
}

func getBuiltinCodeBlock(blockName string) string {
	text, _ := builtinBlocks[blockName]
	return text
}

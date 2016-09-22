package script

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/lfkeitel/inca-tool/src/task"
)

var templatesDir = "templates/"

type Script struct {
	script      string
	dontProcess bool
	task        *task.Task
}

func (s *Script) Clean() error {
	return os.Remove(s.script)
}

func GenerateScript(t *task.Task) (*Script, error) {
	cmds, err := compileCommands(t, t.DefaultCommandBlock)
	if err != nil {
		return nil, err
	}

	if cmds.isScript {
		return &Script{script: cmds.text, dontProcess: true, task: t}, nil
	}

	baseTemplate, err := getTemplateFile(t)
	if err != nil {
		return nil, err
	}

	template, err := generateTemplate(baseTemplate, cmds.text, t.GetAllMetadata())
	if err != nil {
		return nil, err
	}
	return &Script{script: template, task: t}, nil
}

type commands struct {
	text     string
	isScript bool
}

func compileCommands(t *task.Task, entry string) (*commands, error) {
	// Check command block exists
	cb, ok := t.Commands[entry]
	if !ok {
		return nil, fmt.Errorf("Command block \"%s\" not declared\n", entry)
	}

	cmdStr, err := generateScriptText(t, cb)
	if err != nil {
		switch err.Error() {
		case "scriptrun":
			return &commands{text: cmdStr, isScript: true}, nil
		}
	}
	return &commands{text: cmdStr}, err
}

func generateScriptText(t *task.Task, main *task.CommandBlock) (string, error) {
	cmdStr := ""
	prompt := t.Prompt
	if prompt == "" {
		prompt = "#"
	}

	for _, cmd := range main.Commands {
		switch cmd[:3] {
		case "_s ": // Include and run a script file
			return cmd[3:], errors.New("scriptrun")
		case "_c ": // Include another command block
			commandBlock := cmd[3:]
			if commandBlock == main.Name {
				return "", fmt.Errorf("Cannot include self in command block '%s'\n", commandBlock)
			}

			cb, ok := t.Commands[commandBlock]
			if !ok {
				return "", fmt.Errorf("Command block not declared '%s'\n", commandBlock)
			}

			include, err := generateScriptText(t, cb)
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

func getTemplateFile(t *task.Task) (string, error) {
	template := t.Template

	if template == "" {
		template = "expect"
	}

	templateFile := templatesDir + template + "-template.tmpl"
	if _, err := os.Stat(templateFile); os.IsNotExist(err) {
		return "", fmt.Errorf("Template not found: %s\n", template)
	}
	return templateFile, nil
}

func generateTemplate(template, mainSection string, taskVars map[string]string) (string, error) {
	// Generate the base script filename
	tmpFilename := "tmp/builtBaseScript-" + strconv.FormatInt(time.Now().UnixNano(), 10)

	// Copy the template to the base file
	err := copyFileContents(template, tmpFilename)
	if err != nil {
		return "", err
	}

	// Insert the main section, maps are not guaranteed to be in order, so these steps are separate
	if err := insertVariables(tmpFilename, map[string]string{"main": mainSection}); err != nil {
		return "", err
	}

	// Process custom variable data
	if err := insertVariables(tmpFilename, taskVars); err != nil {
		return "", err
	}

	if debug && verbose {
		fmt.Println("Base Variables:")
		for i, v := range taskVars {
			fmt.Printf("  %s: %s\n", i, v)
		}
	}

	// Return the filename for the base script
	return tmpFilename, nil
}

func insertVariables(filename string, vars map[string]string) error {
	file, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}

	for n, v := range vars {
		if n[0] == '_' {
			n = n[1:]
		}
		file = bytes.Replace(file, []byte("{{"+n+"}}"), []byte(v), -1)
	}

	if err := ioutil.WriteFile(filename, file, 0744); err != nil {
		return err
	}
	return nil
}

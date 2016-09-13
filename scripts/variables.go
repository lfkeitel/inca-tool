package scripts

import (
	"io/ioutil"
	"strings"

	"github.com/lfkeitel/inca-tool/devices"
	"github.com/lfkeitel/inca-tool/parser"
)

func insertVariables(filename string, vars map[string]string) error {
	file, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}

	generated := string(file)
	for n, v := range vars {
		generated = strings.Replace(generated, "{{"+n+"}}", v, -1)
	}

	if err := ioutil.WriteFile(filename, []byte(generated), 0744); err != nil {
		return err
	}
	return nil
}

func getVariables(host *devices.Device, task *parser.TaskFile) map[string]string {
	argList := make(map[string]string)
	argList["protocol"] = host.GetSetting("protocol")
	if argList["protocol"] == "" {
		argList["protocol"] = "ssh"
	}

	argList["hostname"] = host.GetSetting("address")
	if argList["hostname"] == "" {
		argList["hostname"] = host.Name
	}

	argList["remote_user"] = host.GetSetting("remote_user")
	if argList["remote_user"] == "" {
		argList["remote_user"] = "root"
	}

	argList["remote_password"] = host.GetSetting("remote_password")

	argList["cisco_enable"] = host.GetSetting("cisco_enable")
	if argList["cisco_enable"] == "" {
		argList["cisco_enable"] = host.GetSetting("remote_password")
	}

	return argList
}

package scripts

import (
	"bytes"
	"io/ioutil"

	"github.com/lfkeitel/inca-tool/devices"
)

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

func getHostVariables(host *devices.Device) map[string]string {
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

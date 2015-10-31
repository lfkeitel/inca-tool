package scripts

import (
	"github.com/dragonrider23/inca-tool/devices"
	"github.com/dragonrider23/inca-tool/parser"
)

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

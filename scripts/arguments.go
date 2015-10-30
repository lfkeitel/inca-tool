package scripts

import (
	"github.com/dragonrider23/inca-tool/devices"
	"github.com/dragonrider23/inca-tool/parser"
)

func getArguments(host *devices.Device, task *parser.TaskFile, eargs []string) []string {
	argList := make([]string, 5+len(eargs))
	argList[0] = host.GetSetting("protocol")
	if argList[0] == "" {
		argList[0] = "ssh"
	}
	argList[1] = host.GetSetting("address")
	argList[2] = host.GetSetting("remote_user")
	argList[3] = host.GetSetting("remote_password")
	if host.GetSetting("cisco_enable") != "" {
		argList[4] = host.GetSetting("cisco_enable")
	} else {
		argList[4] = host.GetSetting("remote_password")
	}

	for i, arg := range eargs {
		argList[i+5] = arg
	}
	return argList
}

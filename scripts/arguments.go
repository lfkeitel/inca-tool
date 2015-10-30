package scripts

import (
	"github.com/dragonrider23/inca-tool/common"
	"github.com/dragonrider23/inca-tool/parser"
)

func getArguments(host common.Host, task *parser.TaskFile, eargs []string) []string {
	argList := make([]string, 6+len(eargs))
	argList[0] = host.Method
	argList[1] = host.Manufacturer
	argList[2] = host.Address
	argList[3] = task.Username
	argList[4] = task.Password
	if task.EnablePassword != "" {
		argList[5] = task.EnablePassword
	} else {
		argList[5] = task.Password
	}

	for i, arg := range eargs {
		argList[i+6] = arg
	}
	return argList
}

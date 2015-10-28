package main

import (
	parser "github.com/dragonrider23/inca-tool/taskfileparser"
)

func execCommandMode(devices []host, task *parser.TaskFile) error {
	// Build an expect script from commands given the device type
	// Substitute special functions beginning with _
	// Then set task.Expect to the compiled string and execute that
	return nil
}

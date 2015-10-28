package main

import (
	parser "github.com/dragonrider23/inca-tool/taskfileparser"
)

func execExpectMode(devices []host, task *parser.TaskFile) error {
	// Take the raw expect string, put it in a temp file?
	// Or, send it as an argument?
	// Probably safer to just write a temporary file
	// Send temporary file to script mode function for execution
	return nil
}

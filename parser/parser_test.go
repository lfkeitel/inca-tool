package parser

import (
	"fmt"
	"reflect"
	"testing"
)

// Constant file header
var testFileHeader = `
# Task - Doesn't really matter, for information purposes
name: Testing
description: Test Description
author: Lee Keitel
date: 10/27/2015
version: 1.0.0`

// Concurrent setting parts
var testFileConcurrent = []string{
	"concurrent: 10",
	"concurrent: 0\ntemplate: bash\nprompt: $",
}

// Device list parts
var testFileDeviceParts = []string{
	`device list: inventory.conf
devices:
    local`,

	`devices:
    local
    juniper`,
}

// Command block parts
var testFileCommandBlocks = []string{
	// test no settings
	`commands: main
    _b juniper-configure
    set system hostname Keitel1
    _b juniper-commit-rollback-failed`,

	// Test settings
	`commands: main type=raw
    _b juniper-configure
    set system hostname Keitel1
    _b juniper-commit-rollback-failed`,

	// Test no main block
	`commands:
    set thing
    this should fail`,
}

// File parts to test as one
var testFileParses = [][]int{
	[]int{0, 0, 0},
	[]int{1, 1, 1},
	[]int{1, 1, 2},
}

// Should the above tests pass or fail
var testFileParsesShouldParse = []bool{
	true,
	true,
	false,
}

// Comparision of passing tests
var testCasesStructs = []*TaskFile{
	&TaskFile{
		Metadata: map[string]string{
			"name":        "Testing",
			"description": "Test Description",
			"author":      "Lee Keitel",
			"date":        "10/27/2015",
			"version":     "1.0.0",
		},
		Concurrent: 10,
		Template:   "",
		Prompt:     "",

		DeviceList: "inventory.conf",
		Devices: []string{
			"local",
		},

		currentBlock: "main",
		Commands: map[string]*CommandBlock{
			"main": &CommandBlock{
				Name: "main",
				Type: "",
				Commands: []string{
					"_b juniper-configure",
					"set system hostname Keitel1",
					"_b juniper-commit-rollback-failed",
				},
			},
		},
	},
	&TaskFile{
		Metadata: map[string]string{
			"name":        "Testing",
			"description": "Test Description",
			"author":      "Lee Keitel",
			"date":        "10/27/2015",
			"version":     "1.0.0",
		},
		Concurrent: 300,
		Template:   "bash",
		Prompt:     "$",

		DeviceList: "devices.conf",
		Devices: []string{
			"local",
			"juniper",
		},

		currentBlock: "main",
		Commands: map[string]*CommandBlock{
			"main": &CommandBlock{
				Name: "main",
				Type: "raw",
				Commands: []string{
					"_b juniper-configure",
					"set system hostname Keitel1",
					"_b juniper-commit-rollback-failed",
				},
			},
		},
	},
}

func TestGeneralParse(t *testing.T) {
	for i, testCase := range testFileParses {
		file := testFileHeader + "\n" +
			testFileConcurrent[testCase[0]] + "\n" +
			testFileDeviceParts[testCase[1]] + "\n" +
			testFileCommandBlocks[testCase[2]]

		parsed, err := ParseString(file)
		if err == nil && !testFileParsesShouldParse[i] {
			t.Errorf("Parse succeeded but should have failed: %s\n", file)
		}
		if err != nil && testFileParsesShouldParse[i] {
			t.Errorf("Parse failed but should have succeeded: %s\n", err.Error())
		}

		if err == nil && testFileParsesShouldParse[i] {
			if err := compareTasks(parsed, i); err != nil {
				t.Error(err.Error())
			}
		}
	}
}

func compareTasks(t *TaskFile, i int) error {
	base := testCasesStructs[i]
	if !reflect.DeepEqual(base, t) {
		if !reflect.DeepEqual(base.Commands["main"], t.Commands["main"]) {
			return fmt.Errorf("Commands not equal:\n%#v\n\n%#v\n\n", base.Commands["main"], t.Commands["main"])
		}
		return fmt.Errorf("Structs not equal:\n%#v\n\n%#v\n\n", base, t)
	}
	return nil
}

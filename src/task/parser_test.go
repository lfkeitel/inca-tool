package task

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
	`inventory: inventory.conf
devices:
    local`,

	`devices:
    local
    juniper`,
}

// Command block parts
var testFileCommandBlocks = []string{
	// test no settings or name
	`commands:
    _b juniper-configure
    set system hostname Keitel1
    _b juniper-commit-rollback-failed`,

	// Test settings no name
	`commands: type=raw
    _b juniper-configure
    set system hostname Keitel1
    _b juniper-commit-rollback-failed`,

	// Test name no settings
	`commands: main
    set thing
    this should pass now`,
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
	true,
}

// Comparision of passing tests
var testCasesStructs = []*Task{
	&Task{
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

		Inventory: "inventory.conf",
		Devices: []string{
			"local",
		},

		currentBlock:        "",
		DefaultCommandBlock: "",
		Commands: map[string]*CommandBlock{
			"": &CommandBlock{
				Name: "",
				Type: "",
				Commands: []string{
					"_b juniper-configure",
					"set system hostname Keitel1",
					"_b juniper-commit-rollback-failed",
				},
			},
		},
	},
	&Task{
		Metadata: map[string]string{
			"name":        "Testing",
			"description": "Test Description",
			"author":      "Lee Keitel",
			"date":        "10/27/2015",
			"version":     "1.0.0",
		},
		Concurrent: 0,
		Template:   "bash",
		Prompt:     "$",

		Inventory: "",
		Devices: []string{
			"local",
			"juniper",
		},

		currentBlock:        "",
		DefaultCommandBlock: "",
		Commands: map[string]*CommandBlock{
			"": &CommandBlock{
				Name: "",
				Type: "raw",
				Commands: []string{
					"_b juniper-configure",
					"set system hostname Keitel1",
					"_b juniper-commit-rollback-failed",
				},
			},
		},
	},
	&Task{
		Metadata: map[string]string{
			"name":        "Testing",
			"description": "Test Description",
			"author":      "Lee Keitel",
			"date":        "10/27/2015",
			"version":     "1.0.0",
		},
		Concurrent: 0,
		Template:   "bash",
		Prompt:     "$",

		Inventory: "",
		Devices: []string{
			"local",
			"juniper",
		},

		currentBlock:        "main",
		DefaultCommandBlock: "",
		Commands: map[string]*CommandBlock{
			"main": &CommandBlock{
				Name: "main",
				Type: "",
				Commands: []string{
					"set thing",
					"this should pass now",
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
			t.Errorf("Parse #%d succeeded but should have failed: %s\n", i, file)
		}
		if err != nil && testFileParsesShouldParse[i] {
			t.Errorf("Parse #%d failed but should have succeeded: %s\n", i, err.Error())
		}

		if err == nil && testFileParsesShouldParse[i] {
			if err := compareTasks(parsed, i); err != nil {
				t.Errorf("Case #%d: %s", i, err.Error())
			}
		}
	}
}

func compareTasks(t *Task, i int) error {
	base := testCasesStructs[i]
	if !reflect.DeepEqual(base, t) {
		if !reflect.DeepEqual(base.Commands[base.currentBlock], t.Commands[base.currentBlock]) {
			return fmt.Errorf(
				"Commands not equal:\n%#v\n\n%#v\n\n",
				base.Commands[base.currentBlock],
				t.Commands[base.currentBlock],
			)
		}
		return fmt.Errorf("Structs not equal:\n%#v\n\n%#v\n\n", base, t)
	}
	return nil
}

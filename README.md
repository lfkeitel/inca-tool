Inca Tool
===========

Inca Tool is a CLI utility to manage infrastructure and deploy configurations using device definitions from an existing Inca v2 installation.

Usage
-----

it [options] [command] [task1 [task2 [task3]...]]

Options:

- `-d` - Enable debug output and functions
- `-dry` - Perform a dry run and list the affected hosts
- `-v` - Enable verbose output

Commands:

- `run` - Run the given task files
- `test` - Test task files for errors
- `version` - Show version information
- `help` - Show this usage information

Task files
----------

Inca Tool uses task files to manage jobs. Task files are simple text file that specify the settings and commands for the job.

```
# Metadata - Doesn't really matter, for information purposes
name: Cisco logging
description: Add logging to 10.254.68.230
author: Lee Keitel
date: 10/27/2015
version: 1.0.0

# How many devices to run at the same time
concurrent: 5

# device list defaults to "devices.conf"
device list: devices.conf

# list of groups or individual devices this task applies to
# devices are defined in the device list file
devices:
    group1
    device2

# List of commands to execute - special commands are prefixed with _
# Comments in command blocks must have no indention or they will be parsed
# as their own command line
# The line is structured as "commands: [name] [key=value]"
commands: main template=bash type=raw
    set hostname AwesomeDevice1
```

Command Block Settings
----------------------

- template
    - Default: expect
    - Values: expect, bash
    - Specifies the template script file to use.
- type
    - Default: expect
    - Values: expect, raw
    - Determines any extra processing needed for the block. Expect will surround the commands with the necessary items to work with Expect such as encapsulation in a send command and issuing an expect command.
- prompt
    - Default: #
    - Values: Any string with no space
    - The prompt used in expect to determine when the next command can be send.

Special Commands
----------------

- `_c foobar` - Inline a command block named foobar
- `_s foobar.baz -- arg1; arg2` - Immediately execute the script named foobar.baz. This stops command processing and uses the script file for the job. The script will be executed with the provided arguments arg1, arg2, etc. Arguments are separated by a semicolon
- `_b foo` - Inline a builtin command block. Inca Tool has a few builtin command blocks for common functions on Juniper and Cisco devices. See below.

Builtin Command Blocks
----------------------

- `juniper-configure` - Enter Juniper's configure mode.
- `juniper-exit-nocommit` - Exits from the Juniper configure mode and if requested will exit without commiting changes. This can be useful to get information from the switch and ensuring no actual configuration change takes place.
- `juniper-commit-rollback-failed` - Attempt to commit changes on a Juniper device and rollback if commit fails. The script as a whole will fail for that device and an error will be show to the console.
- `cisco-enable-mode` - Enter Cisco's Enable exec mode.
- `cisco-end-wrmem` - Exit a Cisco's configure terminal mode and save the running configuration.

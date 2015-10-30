Inca Builder
===========

Inca Builder is a CLI utility to manage infrastructure and deploy configurations using device definitions from an existing Inca v2 installation.

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

Inca Builder uses task files to manage jobs. Task files are simple text file that specify the settings and commands for the job.

```
# Metadata - Doesn't really matter, for information purposes
name: Cisco logging
description: Add logging to 10.254.68.230
author: Lee Keitel
date: 10/27/2015
version: 1.0.0

# How many devices to run at the same time
concurrent: 5

device list: devices.conf
# manufacturer as defined in the devices file
manufacturer: cisco
# as defined in the devices file
group: ssh
# short hand of above, can not be used in combination with the above two settings
filter: cisco:ssh

username: user
password: pass
enablepassword: enable

# List of commands to execute - special commands are prefixed with _
# Comments in command blocks must have no indention or they will be parsed
# as their own command line
# The line is structured as "commands: [name] [key=value]"
commands: main template=bash type=raw
    touch keitel1
```

Special Commands
----------------

- `_c foobar` - Inline a command block named foobar
- `_s foobar.baz -- arg1; arg2` - Immediately execute the script named foobar.baz. This stops command processing and uses the script file for the job. The script will be executed with the normal arguments (see below) plus the provided arguments arg1, arg2, etc. Arguments are separated by a semicolon
- `_b foo` - Inline a builtin command block. Inca Builder has a few builtin command blocks for common functions on Juniper and Cisco devices. See below.

Script Files
------------

Script files are called with the following arguments:

1. protocol (group)
2. manufacturer (as defined in device list)
3. hostname
4. username
5. password
6. enable password

Any additional arguments will be passed as argument 7 onward.

Builtin Command Blocks
----------------------

- `juniper-exit-nocommit` - Exits from the Juniper configure mode and if requested will exit without commiting changes. This can be useful to get information from the switch and ensuring no actual configuration change takes place.
- `juniper-commit-rollback-failed` - Attempt to commit changes on a Juniper device and rollback if commit fails. The script as a whole will fail for that device and an error will be show to the console.
- `cisco-end-wrmem` - Exit a Cisco's configure terminal mode and save the running configuration.

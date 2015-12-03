Inca Tool
===========

Inca Tool is a CLI utility to manage infrastructure and deploy configurations across multiple devices at the same time.

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

Inca Tool uses task files to manage jobs. Task files are simple text files that specify the settings and commands for a job.

```
# Metadata - Doesn't really matter, for information purposes
name: Cisco logging
description: Add logging to 10.254.68.230
author: Lee Keitel
date: 10/27/2015
version: 1.0.0

# How many devices to run at the same time, Defaults to 300
concurrent: 5

# script template to use, Defaults to expect, Values can be "expect", "bash"
template: expect

# device list file, Defaults to "devices.conf"
device list: devices.conf

# unique part of prompt to wait for when using Expect
prompt: #

# list of groups or individual devices this task applies to
# devices are defined in the device list file
devices:
    group1
    device2

# List of commands to execute - special commands are prefixed with _
# Comments in command blocks must have no indention or they will be parsed
# as their own command line
# The line is structured as "commands: [name] [key=value]"
commands: main type=raw
    set hostname AwesomeDevice1
```

Command Block Settings
----------------------

- type
    - Default: expect
    - Values: expect, raw
    - Determines any extra processing needed for the block. Expect will surround the commands with the necessary items to work with Expect such as encapsulation in a send command and issuing an expect command.

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

Device List Syntax
------------------

- All devices must be in a group.
- The global group can only contain settings.
- Device names cannot contain a space.
- Group names may contain numbers, letters, underscores, hyphens and spaces.
- If multiple groups are declared with the same name, the devices will be appended to a single group.
    - Example: The following will result in a single group named "group1" with devices device1, device2, device3, and device4.
```
[group1]
device1
device2

[group1]
device3
device4
```

- Devices may be in multiple groups. Any device settings must be declared on the first declaration.
- Settings are "key=value" pairs separated by a space on the same line as the device name. If a setting value contains a space, it must be enclosed in double quotes.
- Both devices and groups may have settings
- Order of setting precedence is Global -> Task -> Group -> Device
- Available settings:
    - remote_user - Defaults to "root"
    - remote_password - Defaults to ""
    - cisco_enable - Defaults to remote_password
    - protocol - Defaults to "ssh"
    - address - Defaults to device name

Example:

```
[global]
remote_user = user
remote_password = pass

[server room]
Server_Switch_1 address=10.0.0.1
Switch2.example.com

[building 1] remote_user="jarvis"
Building1_1 address=10.0.0.2 protocol=telnet
Builsing1_2 address=10.0.0.3 remote_password="chicken feet"

[all switches]
Server_Switch_1
Switch2.example.com
Building1_1
Building1_2
```

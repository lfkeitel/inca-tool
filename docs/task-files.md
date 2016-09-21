#Task Files

##What are they?

Task files are make up the core of Inca Tool. A task file is used to manage a particular job. A task file is meant to perform a single task. However, multiple tasks files may be ran at once if needed. Task files use a custom syntax that provides metadata for the task, various parameters, and the commands to execute for the task.

##Syntax

The syntax of task files can be broken up into three types: key: value, simple list, and extended list.

###Key Value
Most settings are simple key value pairs which take the form:

```
key: value
```

The space before the value is not required but is there for readability.

###Simple List
A list is defined as:

```
key:
    value1
    value2
```

Note that the values are indented and have the same indention. For both types of lists, indention matters. If two values have a different indention, or if one uses tabs and the other spaces, the task file will return an error when it's parsed. Be sure to use consistent indention.

###Extended List
And extended list is a simple list but can take a main value as well as key value pairs as settings for the list:

```
key: value setting=val setting2=val2
    list value1
    list value2
```

Although it may seem strange at first, this is can be very powerful for creating complex task files.

##Task File Structure
A task file is made up of several parts. Some optional, some required.

###Metadata
Metadata does not affect how the task is ran, but can be used to ensure the correct task was run and can serve as documentation. All metadata are simple key value pairs. Here's a list of all available metadata settings:

- name
- description
- author
- date
- version

###Job Settings
There are a couple of settings that affect how templates are generated and ran.

- concurrent
    - Type: key-value integer
    - Default: 300
    - Valid values: Any integer
    - Description:
        - The number of devices that can be configured concurrently. The higher this number, the more file descriptors are needed to run the job. Setting this to 0 means no limit.
- template
    - Type: key-value string
    - Default: expect
    - Valid values: expect, bash
    - Description:
        - The template to use when generating a job file.
- prompt
    - Type key-value string
    - Default: #
    - Valid values: Any string
    - Description:
        - The unique part of a prompt to wait for when using Expect.

###Inventory
For a task to run, Inca Tool needs to know which devices should be configured. This is where the inventory file and device filter come in.

- inventory
    - Type: key-value string
    - Default: devices.conf
    - Valid values: File system path
    - Description:
        - The path to the inventory file. It may be absolute or relative. The path will be relative to the current working directory. NOTE: It is recommended to provide the inventory file via the -i cli flag. If this setting is used in a task file, it will override the file given on the command line.
- devices
    - Type: simple list
    - Default: Empty
    - Value values: group or device names
    - Description:
        - This list contains the group or devices names that will configured with the task. If a group or name doesn't exist in the provided inventory file, an error will be given.

###Command Blocks
Command blocks are where the set of commands are defined that will be ran on the client device. Multiple command blocks may be created so long as they have different names. One command block must be named `main`. This is the block that will be executed. Other block can be included using the `_c` syntax described below. If a `main` block doesn't exist, the task will not run and an error will be produced.

Block Syntax:

```
commands: name setting=value
    command 1
    command 2
    _c other-command-block
    _b builtin-command-block
    _s /path/to/script
```

####Command Block Settings

- type
    - Default: expect
    - Values: expect, raw
    - Determines any extra processing needed for the block. Expect will surround the commands with the necessary items to work with Expect such as encapsulation in a send command and issuing an expect command.

####Special Commands

- `_c foobar` - Inline a command block named foobar
- `_s foobar.baz -- arg1; arg2` - Immediately execute the script named foobar.baz. This stops command processing and uses the script file for the job. The script will be executed with the provided arguments arg1, arg2, etc. Arguments are separated by a semicolon
- `_b foo` - Inline a builtin command block. Inca Tool has a few builtin command blocks for common functions on Juniper and Cisco devices. See below.

####Builtin Command Blocks

- `juniper-configure` - Enter Juniper's configure mode.
- `juniper-exit-nocommit` - Exits from the Juniper configure mode and if requested will exit without commiting changes. This can be useful to get information from the switch and ensuring no actual configuration change takes place.
- `juniper-commit-rollback-failed` - Attempt to commit changes on a Juniper device and rollback if commit fails. The script as a whole will fail for that device and an error will be show to the console.
- `cisco-enable-mode` - Enter Cisco's Enable exec mode.
- `cisco-end-wrmem` - Exit a Cisco's configure terminal mode and save the running configuration.

##Full Task File Example

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

# device list file, defaults to "devices.conf", recommended to use the -i flag instead
inventory: devices.conf

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

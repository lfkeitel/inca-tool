Inca Tool
=========

Inca tool is a CLI utility to manage infrastructure and deploy configurations using device definitions from an existing Inca v2 installation.

Usage
-----

it [options]

Options
-------

-a Arguments given to script in the correct order, no dollar signs, comma separated
-d Device definitions file
-dry List devices that would be affected by the task, doesn't actually run the script
-e Enable password for Cisco devices
-f Filter to apply to devices
-p Password for above username
-s Script to run for each device
-u Username to login to device

Inca tool will take the definitions from the devices file, filter them with the given -f argument, and iterate over them calling the script -s with the arguments -a in order.

The arguments for a file may also be specified in the filename itself:

cisco-new-logging--username,password,enable.exp

The arguments should be at the end of the file before the extension and separated by a double hyphen "--". Unlike the normal device types file, the -a flag nor the filename arguments should have a dollar sign. Passing the argument list with -a is recommended and will override whatever the filename says.

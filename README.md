#Inca Tool

Inca Tool is a CLI utility to manage infrastructure and deploy configurations across multiple devices at the same time. It's different from other automation systems in that Inca Tool is more "raw". Meaning, it can be very versatile and work with varying devices and systems.

##Usage

it [options] [command] [task1 [task2 [task3]...]]

Options:

- `-d` - Enable debug output and functions
- `-r` - Perform a dry run and list the affected hosts
- `-v` - Enable verbose output
- `-i` - Specify an inventory file to use, if a task file specifies a file, this setting will override it

Commands:

- `run` - Run the given task files
- `test` - Test task files for errors
- `version` - Show version information
- `help` - Show this usage information

##Where's the documentation?

Documentation is available on [Read the Docs](http://inca-tool.readthedocs.io/en/latest/).

##What is Inca Tool?

Inca Tool is a system automation application designed to be client-less and system independent. It was created to work with older network infrastructure that doesn't support newer configuration methods such as netconf.

##Why another automation tool?

Inca Tool was created because our environment is extremely mixed. We have multiple manufactures for network infrastructure and multiple OSes for server systems. Although tools such as Ansible can do what we need for server systems, it's very lacking in support for out network infrastructure. Ansible has recently come out with tools to configure IOS and JunOS systems, but unfortunately the modules require much newer software versions than we have in production. Although it's never good to have outdated software, those systems still need to be managed. Out of that need, Inca Tool was born.

##Inca Tool appears to be just a script generator.

Yes. Essentially that's exactly what it is. Inca Tool takes a template and the user commands to execute, produces a script, and then runs it on the local machine. While not very sophisticated, it provides the most flexibility for different systems. The heart of Inca Tool is actually a completely different program called [Expect](http://expect.sourceforge.net/). Using Expect, Inca Tool can interact with a device as if the user was doing it themselves. Other templates can be utilized to use other programs. The hope is that the current template system will be expanded to provide more templates and to more easily create templates.

##License

Inca Tool is release under the BSD 3 Clause License. License text can be found in the LICENSE.md file.

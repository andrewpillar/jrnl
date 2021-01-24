package main

import (
	"fmt"
	"os"
	"sort"
)

type Command struct {
	Argv0 string // Argv0 is the name of the process running the command.
	Usage string // Usage is the usage line of the command.
	Short string // Short is a short description of what the command does.
	Long  string // Long is a longer more in depth description of the command.

	// Run is the function that will actually execute the command. This will
	// be passed a pointer to the Command itself, along with the arguments
	// given to it. The first item in the arguments list will be the command
	// name.
	Run func(*Command, []string)

	// Commands is the set of sub-commands the command could have.
	Commands *CommandSet
}

type CommandSet struct {
	longest int
	names   []string
	cmds    map[string]*Command

	Argv0 string
	Long  string
	Usage func()
}

type ErrCommandNotFound string

func helpCmd(cmd *Command, args []string) {
	if len(args) < 2 {
		cmd.Commands.usage()
		return
	}

	name := args[1]

	cmd1, ok := cmd.Commands.cmds[name]

	if !ok {
		fmt.Fprintf(os.Stderr, "%s %s: no such command. Run '%s %s'.\n", cmd.Argv0, args[0], cmd.Argv0, args[0])
		os.Exit(1)
	}

	if cmd1.Long == "" && cmd1.Commands != nil {
		fmt.Printf("usage: %s %s\n", cmd1.Argv0, cmd1.Usage)
		cmd1.Commands.usage()
		return
	}

	fmt.Printf("usage: %s %s\n", cmd1.Argv0, cmd1.Usage)

	if cmd1.Long != "" {
		fmt.Println()
		fmt.Println(cmd1.Long)
	}
}

func HelpCmd(cmds *CommandSet) *Command {
	return &Command{
		Usage:    cmds.Argv0 + " help [command]",
		Short:    "display usage and help information about a given command",
		Long:     "",
		Run:      helpCmd,
		Commands: cmds,
	}
}

func (c *CommandSet) defaultUsage() {
	sort.Strings(c.names)

	fmt.Println(c.Long)

	if len(c.names) > 0 {
		fmt.Println("The commands are:")
		fmt.Println()
	}

	printHelp := false

	for _, name := range c.names {
		if name == "help" {
			printHelp = true
			continue
		}

		fmt.Printf("    %s%*s%s\n", name, c.longest-len(name)+4, " ", c.cmds[name].Short)
	}

	if printHelp {
		fmt.Printf("\nUse '%s help [command]' for more information about that command.\n", c.Argv0)
	}
}

func (c *CommandSet) usage() {
	if c.Usage == nil {
		c.defaultUsage()
		return
	}
	c.Usage()
}

func (c *CommandSet) Add(name string, cmd *Command) {
	if c.cmds == nil {
		c.cmds = make(map[string]*Command)
	}

	if _, ok := c.cmds[name]; !ok {
		if l := len(name); l > c.longest {
			c.longest = l
		}

		cmd.Argv0 = c.Argv0

		c.names = append(c.names, name)
		c.cmds[name] = cmd
	}
}

func (c *CommandSet) Parse(args []string) error {
	if len(args) < 1 {
		c.usage()
		return nil
	}

	name := args[0]

	cmd, ok := c.cmds[name]

	if !ok {
		return ErrCommandNotFound(name)
	}
	cmd.Run(cmd, args)
	return nil
}

func (e ErrCommandNotFound) Error() string { return "command not found " + string(e) }

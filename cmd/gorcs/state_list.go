package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/arran4/golang-rcs/internal/cli"
)

var _ Cmd = (*StateList)(nil)

type StateList struct {
	*State
	Flags       *flag.FlagSet
	files       []string
	SubCommands map[string]Cmd
}

type UsageDataStateList struct {
	*StateList
	Recursive bool
}

func (c *StateList) Usage() {
	err := executeUsage(os.Stderr, "state_list_usage.txt", UsageDataStateList{c, false})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error generating usage: %s\n", err)
	}
}

func (c *StateList) UsageRecursive() {
	err := executeUsage(os.Stderr, "state_list_usage.txt", UsageDataStateList{c, true})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error generating usage: %s\n", err)
	}
}

func (c *StateList) Execute(args []string) error {
	if len(args) > 0 {
		if cmd, ok := c.SubCommands[args[0]]; ok {
			return cmd.Execute(args[1:])
		}
	}
	var remainingArgs []string
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if arg == "--" {
			remainingArgs = append(remainingArgs, args[i+1:]...)
			break
		}
		if strings.HasPrefix(arg, "-") {
			name := arg
			trimmedName := strings.TrimLeft(name, "-")
			switch trimmedName {
			case "help", "h":
				c.Usage()
				return nil
			default:
				return fmt.Errorf("unknown flag: %s", name)
			}
		} else {
			remainingArgs = append(remainingArgs, arg)
		}
	}
	// Handle vararg files
	{
		varArgStart := 0
		if varArgStart < len(remainingArgs) {
			varArgs := remainingArgs[varArgStart:]
			c.files = varArgs
		}
	}

	if err := cli.StateList(c.files...); err != nil {
		return fmt.Errorf("state list failed: %w", err)
	}

	return nil
}

func (c *State) NewStateList() *StateList {
	set := flag.NewFlagSet("list", flag.ContinueOnError)
	v := &StateList{
		State:       c,
		Flags:       set,
		SubCommands: make(map[string]Cmd),
	}
	set.Usage = v.Usage

	v.SubCommands["help"] = &InternalCommand{
		Exec: func(args []string) error {
			for _, arg := range args {
				if arg == "-deep" {
					v.UsageRecursive()
					return nil
				}
			}
			v.Usage()
			return nil
		},
		UsageFunc: v.Usage,
	}
	v.SubCommands["usage"] = &InternalCommand{
		Exec: func(args []string) error {
			for _, arg := range args {
				if arg == "-deep" {
					v.UsageRecursive()
					return nil
				}
			}
			v.Usage()
			return nil
		},
		UsageFunc: v.Usage,
	}
	return v
}

package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/arran4/golang-rcs/internal/cli"
)

var _ Cmd = (*StateSet)(nil)

type StateSet struct {
	*State
	Flags       *flag.FlagSet
	state       string
	rev         string
	files       []string
	SubCommands map[string]Cmd
}

type UsageDataStateSet struct {
	*StateSet
	Recursive bool
}

func (c *StateSet) Usage() {
	err := executeUsage(os.Stderr, "state_set_usage.txt", UsageDataStateSet{c, false})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error generating usage: %s\n", err)
	}
}

func (c *StateSet) UsageRecursive() {
	err := executeUsage(os.Stderr, "state_set_usage.txt", UsageDataStateSet{c, true})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error generating usage: %s\n", err)
	}
}

func (c *StateSet) Execute(args []string) error {
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
			value := ""
			hasValue := false
			if strings.Contains(arg, "=") {
				parts := strings.SplitN(arg, "=", 2)
				name = parts[0]
				value = parts[1]
				hasValue = true
			}
			trimmedName := strings.TrimLeft(name, "-")
			switch trimmedName {
			case "state":
				if !hasValue {
					if i+1 < len(args) {
						value = args[i+1]
						i++
					} else {
						return fmt.Errorf("flag %s requires a value", name)
					}
				}
				c.state = value
			case "rev":
				if !hasValue {
					if i+1 < len(args) {
						value = args[i+1]
						i++
					} else {
						return fmt.Errorf("flag %s requires a value", name)
					}
				}
				c.rev = value
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

	if err := cli.StateSet(c.rev, c.state, c.files...); err != nil {
		return fmt.Errorf("state set failed: %w", err)
	}

	return nil
}

func (c *State) NewStateSet() *StateSet {
	set := flag.NewFlagSet("set", flag.ContinueOnError)
	v := &StateSet{
		State:       c,
		Flags:       set,
		SubCommands: make(map[string]Cmd),
	}
	set.Usage = v.Usage

	set.StringVar(&v.state, "state", "", "State to set")
	set.StringVar(&v.rev, "rev", "", "Revision to modify")

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

package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
)

var _ Cmd = (*AccessList)(nil)

type AccessList struct {
	*RootCmd
	Flags       *flag.FlagSet
	SubCommands map[string]Cmd
}

type UsageDataAccessList struct {
	*AccessList
	Recursive bool
}

func (c *AccessList) Usage() {
	err := executeUsage(os.Stderr, "access_list_usage.txt", UsageDataAccessList{c, false})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error generating usage: %s\n", err)
	}
}

func (c *AccessList) UsageRecursive() {
	err := executeUsage(os.Stderr, "access_list_usage.txt", UsageDataAccessList{c, true})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error generating usage: %s\n", err)
	}
}

func (c *AccessList) Execute(args []string) error {
	if len(args) > 0 {
		if cmd, ok := c.SubCommands[args[0]]; ok {
			return cmd.Execute(args[1:])
		}
	}
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if arg == "--" {
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
		}
	}

	c.Usage()

	return nil
}

func (c *RootCmd) NewAccessList() *AccessList {
	set := flag.NewFlagSet("access-list", flag.ContinueOnError)
	v := &AccessList{
		RootCmd:     c,
		Flags:       set,
		SubCommands: make(map[string]Cmd),
	}
	set.Usage = v.Usage

	v.SubCommands["copy"] = v.NewCopy()

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

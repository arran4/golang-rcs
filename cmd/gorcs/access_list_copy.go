package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/arran4/golang-rcs/internal/cli"
)

var _ Cmd = (*AccessListCopy)(nil)

type AccessListCopy struct {
	*AccessList
	Flags       *flag.FlagSet
	from        string
	files       []string
	SubCommands map[string]Cmd
}

type UsageDataAccessListCopy struct {
	*AccessListCopy
	Recursive bool
}

func (c *AccessListCopy) Usage() {
	err := executeUsage(os.Stderr, "access_list_copy_usage.txt", UsageDataAccessListCopy{c, false})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error generating usage: %s\n", err)
	}
}

func (c *AccessListCopy) UsageRecursive() {
	err := executeUsage(os.Stderr, "access_list_copy_usage.txt", UsageDataAccessListCopy{c, true})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error generating usage: %s\n", err)
	}
}

func (c *AccessListCopy) Execute(args []string) error {
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
			case "from":
				if !hasValue {
					if i+1 < len(args) {
						value = args[i+1]
						i++
					} else {
						return fmt.Errorf("flag %s requires a value", name)
					}
				}
				c.from = value
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

	c.files = remainingArgs

	if c.from == "" {
		return fmt.Errorf("flag -from is required")
	}
	if len(c.files) == 0 {
		return fmt.Errorf("at least one target file is required")
	}

	if err := cli.AccessListCopy(c.from, c.files...); err != nil {
		return fmt.Errorf("access-list copy failed: %w", err)
	}

	return nil
}

func (c *AccessList) NewCopy() *AccessListCopy {
	set := flag.NewFlagSet("copy", flag.ContinueOnError)
	v := &AccessListCopy{
		AccessList:  c,
		Flags:       set,
		SubCommands: make(map[string]Cmd),
	}
	set.StringVar(&v.from, "from", "", "Source RCS file to copy access list from")
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

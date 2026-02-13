package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/arran4/golang-rcs/internal/cli"
)

var _ Cmd = (*Validate)(nil)

type Validate struct {
	*RootCmd
	Flags       *flag.FlagSet
	output      string
	force       bool
	files       []string
	SubCommands map[string]Cmd
}

type UsageDataValidate struct {
	*Validate
	Recursive bool
}

func (c *Validate) Usage() {
	err := executeUsage(os.Stderr, "validate_usage.txt", UsageDataValidate{c, false})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error generating usage: %s\n", err)
	}
}

func (c *Validate) UsageRecursive() {
	err := executeUsage(os.Stderr, "validate_usage.txt", UsageDataValidate{c, true})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error generating usage: %s\n", err)
	}
}

func (c *Validate) Execute(args []string) error {
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
		if strings.HasPrefix(arg, "-") && arg != "-" {
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
			case "output", "o":
				if !hasValue {
					if i+1 < len(args) {
						value = args[i+1]
						i++
					} else {
						return fmt.Errorf("flag %s requires a value", name)
					}
				}
				c.output = value
			case "force", "f":
				if hasValue {
					b, err := strconv.ParseBool(value)
					if err != nil {
						return fmt.Errorf("invalid boolean value for flag %s: %s", name, value)
					}
					c.force = b
				} else {
					c.force = true
				}
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
		if varArgStart > len(remainingArgs) {
			varArgStart = len(remainingArgs)
		}
		varArgs := remainingArgs[varArgStart:]
		c.files = varArgs
	}

	if err := cli.Validate(c.output, c.force, c.files...); err != nil {
		return err
	}

	return nil
}

func (c *RootCmd) NewValidate() *Validate {
	set := flag.NewFlagSet("validate", flag.ContinueOnError)
	v := &Validate{
		RootCmd:     c,
		Flags:       set,
		SubCommands: make(map[string]Cmd),
	}

	set.StringVar(&v.output, "o", "", "Output file path")
	set.StringVar(&v.output, "output", "", "Output file path")

	set.BoolVar(&v.force, "f", false, "Force overwrite output")
	set.BoolVar(&v.force, "force", false, "Force overwrite output")

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

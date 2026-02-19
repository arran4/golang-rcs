package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"errors"
	"github.com/arran4/golang-rcs/cmd"
	"github.com/arran4/golang-rcs/internal/cli"
)

var _ Cmd = (*Clean)(nil)

type Clean struct {
	*Locks
	Flags         *flag.FlagSet
	revision      string
	unlock        bool
	user          string
	quiet         bool
	files         []string
	SubCommands   map[string]Cmd
	CommandAction func(c *Clean) error
}

type UsageDataClean struct {
	*Clean
	Recursive bool
}

func (c *Clean) Usage() {
	err := executeUsage(os.Stderr, "clean_usage.txt", UsageDataClean{c, false})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error generating usage: %s\n", err)
	}
}

func (c *Clean) UsageRecursive() {
	err := executeUsage(os.Stderr, "clean_usage.txt", UsageDataClean{c, true})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error generating usage: %s\n", err)
	}
}

func (c *Clean) Execute(args []string) error {
	if len(args) > 0 {
		if cmd, ok := c.SubCommands[args[0]]; ok {
			return cmd.Execute(args[1:])
		}
	}
	remainingArgs := make([]string, 0, len(args))
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

			// Handle flags
			if trimmedName == "help" || trimmedName == "h" {
				c.Usage()
				return nil
			}
			if trimmedName == "q" {
				c.quiet = true
				continue
			}

			if strings.HasPrefix(trimmedName, "r") || strings.HasPrefix(trimmedName, "R") {
				prefix := "r"
				if strings.HasPrefix(trimmedName, "R") {
					prefix = "R"
				}
				if trimmedName == prefix {
					if !hasValue {
						if i+1 < len(args) {
							value = args[i+1]
							i++
						} else {
							return fmt.Errorf("flag -%s requires a value", prefix)
						}
					}
					c.revision = value
				} else {
					c.revision = strings.TrimPrefix(trimmedName, prefix)
				}
				continue
			}

			if strings.HasPrefix(trimmedName, "u") {
				c.unlock = true
				if trimmedName == "u" {
					if hasValue {
						c.revision = value
					}
					// If -u without value, it unlocks default/locked rev.
				} else {
					// -uRev
					c.revision = strings.TrimPrefix(trimmedName, "u")
				}
				continue
			}

			if strings.HasPrefix(trimmedName, "w") {
				if trimmedName == "w" {
					if !hasValue {
						if i+1 < len(args) {
							value = args[i+1]
							i++
						} else {
							return fmt.Errorf("flag -w requires a value")
						}
					}
					c.user = value
				} else {
					c.user = strings.TrimPrefix(trimmedName, "w")
				}
				continue
			}

			return fmt.Errorf("unknown flag: %s", name)

		} else {
			remainingArgs = append(remainingArgs, arg)
		}
	}
	c.files = remainingArgs

	if c.CommandAction != nil {
		if err := c.CommandAction(c); err != nil {
			return fmt.Errorf("clean failed: %w", err)
		}
	} else {
		c.Usage()
	}

	return nil
}

func (c *Locks) NewClean() *Clean {
	set := flag.NewFlagSet("clean", flag.ContinueOnError)
	v := &Clean{
		Locks:       c,
		Flags:       set,
		SubCommands: make(map[string]Cmd),
	}
	set.Usage = v.Usage

	v.CommandAction = func(c *Clean) error {
		dirty, err := cli.LocksClean(c.revision, c.unlock, c.user, c.quiet, c.files...)
		if err != nil {
			if errors.Is(err, cmd.ErrPrintHelp) {
				c.Usage()
				return nil
			}
			if errors.Is(err, cmd.ErrHelp) {
				fmt.Fprintf(os.Stderr, "Use '%s help' for more information.\n", os.Args[0])
				return nil
			}
			return err
		}
		if dirty {
			return &cmd.ErrExitCode{Code: 1}
		}
		return nil
	}

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

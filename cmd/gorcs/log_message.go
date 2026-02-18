package main

import (
	"flag"
	"fmt"
	"os"
)

var _ Cmd = (*LogMessage)(nil)

type LogMessage struct {
	*Log
	Flags       *flag.FlagSet
	SubCommands map[string]Cmd
}

type UsageDataLogMessage struct {
	*LogMessage
	Recursive bool
}

func (c *LogMessage) Usage() {
	fmt.Fprintf(os.Stderr, "Usage of %s log message:\n", os.Args[0])
	c.Flags.PrintDefaults()
	fmt.Fprintln(os.Stderr, "  Commands:")
	for name := range c.SubCommands {
		fmt.Fprintf(os.Stderr, "    %s\n", name)
	}
}

func (c *LogMessage) UsageRecursive() {
	c.Usage()
	for _, cmd := range c.SubCommands {
		if r, ok := cmd.(interface{ UsageRecursive() }); ok {
			r.UsageRecursive()
		} else {
			cmd.Usage()
		}
	}
}

func (c *LogMessage) Execute(args []string) error {
	if len(args) > 0 {
		if cmd, ok := c.SubCommands[args[0]]; ok {
			return cmd.Execute(args[1:])
		}
	}
	c.Usage()
	return fmt.Errorf("unknown subcommand")
}

func (c *Log) NewLogMessage() *LogMessage {
	set := flag.NewFlagSet("message", flag.ContinueOnError)
	v := &LogMessage{
		Log:         c,
		Flags:       set,
		SubCommands: make(map[string]Cmd),
	}
	set.Usage = v.Usage

	v.SubCommands["change"] = v.NewLogMessageChange()
	v.SubCommands["print"] = v.NewLogMessagePrint()
	v.SubCommands["list"] = v.NewLogMessageList()

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
	return v
}

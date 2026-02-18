package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/arran4/golang-rcs/internal/cli"
)

var _ Cmd = (*Log)(nil)

type Log struct {
	*RootCmd
	Flags       *flag.FlagSet
	SubCommands map[string]Cmd
}

type UsageDataLog struct {
	*Log
	Recursive bool
}

func (c *Log) Usage() {
	err := executeUsage(os.Stderr, "log_usage.txt", UsageDataLog{c, false})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error generating usage: %s\n", err)
	}
}

func (c *Log) UsageRecursive() {
	err := executeUsage(os.Stderr, "log_usage.txt", UsageDataLog{c, true})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error generating usage: %s\n", err)
	}
}

func (c *Log) Execute(args []string) error {
	if len(args) > 0 {
		if cmd, ok := c.SubCommands[args[0]]; ok {
			return cmd.Execute(args[1:])
		}
	}
	c.Usage()
	return nil
}

func (c *RootCmd) NewLog() *Log {
	set := flag.NewFlagSet("log", flag.ContinueOnError)
	v := &Log{
		RootCmd:     c,
		Flags:       set,
		SubCommands: make(map[string]Cmd),
	}
	set.Usage = v.Usage

	v.SubCommands["message"] = v.NewLogMessage()

	v.SubCommands["help"] = &InternalCommand{
		Exec: func(args []string) error {
			v.Usage()
			return nil
		},
		UsageFunc: v.Usage,
	}
	return v
}

// LogMessage command

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
	err := executeUsage(os.Stderr, "log_message_usage.txt", UsageDataLogMessage{c, false})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error generating usage: %s\n", err)
	}
}

func (c *LogMessage) UsageRecursive() {
	err := executeUsage(os.Stderr, "log_message_usage.txt", UsageDataLogMessage{c, true})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error generating usage: %s\n", err)
	}
}

func (c *LogMessage) Execute(args []string) error {
	if len(args) > 0 {
		if cmd, ok := c.SubCommands[args[0]]; ok {
			return cmd.Execute(args[1:])
		}
	}
	c.Usage()
	return nil
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
			v.Usage()
			return nil
		},
		UsageFunc: v.Usage,
	}
	return v
}

// LogMessageChange command

type LogMessageChange struct {
	*LogMessage
	Flags    *flag.FlagSet
	Revision string
	Message  string
	Files    []string
}

func (c *LogMessageChange) Usage() {
	err := executeUsage(os.Stderr, "log_message_change_usage.txt", c)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error generating usage: %s\n", err)
	}
}

func (c *LogMessageChange) Execute(args []string) error {
	c.Flags.Parse(args)
	c.Files = c.Flags.Args()

	if c.Revision == "" {
		return fmt.Errorf("flag -rev is required")
	}
	if c.Message == "" {
		return fmt.Errorf("flag -m is required")
	}
	if len(c.Files) == 0 {
		return fmt.Errorf("at least one file is required")
	}

	return cli.LogMessageChange(c.Revision, c.Message, c.Files...)
}

func (c *LogMessage) NewLogMessageChange() *LogMessageChange {
	set := flag.NewFlagSet("change", flag.ContinueOnError)
	v := &LogMessageChange{
		LogMessage: c,
		Flags:      set,
	}
	set.StringVar(&v.Revision, "rev", "", "Revision to change")
	set.StringVar(&v.Message, "m", "", "New log message")
	set.Usage = v.Usage
	return v
}

// LogMessagePrint command

type LogMessagePrint struct {
	*LogMessage
	Flags    *flag.FlagSet
	Revision string
	Files    []string
}

func (c *LogMessagePrint) Usage() {
	err := executeUsage(os.Stderr, "log_message_print_usage.txt", c)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error generating usage: %s\n", err)
	}
}

func (c *LogMessagePrint) Execute(args []string) error {
	c.Flags.Parse(args)
	c.Files = c.Flags.Args()

	if c.Revision == "" {
		return fmt.Errorf("flag -rev is required")
	}
	if len(c.Files) == 0 {
		return fmt.Errorf("at least one file is required")
	}

	return cli.LogMessagePrint(c.Revision, c.Files...)
}

func (c *LogMessage) NewLogMessagePrint() *LogMessagePrint {
	set := flag.NewFlagSet("print", flag.ContinueOnError)
	v := &LogMessagePrint{
		LogMessage: c,
		Flags:      set,
	}
	set.StringVar(&v.Revision, "rev", "", "Revision to print")
	set.Usage = v.Usage
	return v
}

// LogMessageList command

type LogMessageList struct {
	*LogMessage
	Flags *flag.FlagSet
	Files []string
}

func (c *LogMessageList) Usage() {
	err := executeUsage(os.Stderr, "log_message_list_usage.txt", c)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error generating usage: %s\n", err)
	}
}

func (c *LogMessageList) Execute(args []string) error {
	c.Flags.Parse(args)
	c.Files = c.Flags.Args()

	if len(c.Files) == 0 {
		return fmt.Errorf("at least one file is required")
	}

	return cli.LogMessageList(c.Files...)
}

func (c *LogMessage) NewLogMessageList() *LogMessageList {
	set := flag.NewFlagSet("list", flag.ContinueOnError)
	v := &LogMessageList{
		LogMessage: c,
		Flags:      set,
	}
	set.Usage = v.Usage
	return v
}

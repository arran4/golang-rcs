package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/arran4/golang-rcs/internal/cli"
)

var _ Cmd = (*LogMessagePrint)(nil)

type LogMessagePrint struct {
	*LogMessage
	Flags    *flag.FlagSet
	revision string
	files    []string
}

type UsageDataLogMessagePrint struct {
	*LogMessagePrint
	Recursive bool
}

func (c *LogMessagePrint) Usage() {
	// TODO: Create usage template
	fmt.Fprintf(os.Stderr, "Usage of %s log message print:\n", os.Args[0])
	c.Flags.PrintDefaults()
}

func (c *LogMessagePrint) Execute(args []string) error {
	if err := c.Flags.Parse(args); err != nil {
		return err
	}
	c.files = c.Flags.Args()
	if c.revision == "" {
		return fmt.Errorf("revision required")
	}
	return cli.LogMessagePrint(c.revision, c.files...)
}

func (c *LogMessage) NewLogMessagePrint() *LogMessagePrint {
	set := flag.NewFlagSet("print", flag.ContinueOnError)
	v := &LogMessagePrint{
		LogMessage: c,
		Flags:      set,
	}
	set.StringVar(&v.revision, "rev", "", "Revision to print log message for")
	set.Usage = v.Usage
	return v
}

package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/arran4/golang-rcs/internal/cli"
)

var _ Cmd = (*Locks)(nil)

type Locks struct {
	*RootCmd
	Flags *flag.FlagSet
}

func (c *Locks) Execute(args []string) error {
	if len(args) == 0 {
		c.Usage()
		return nil
	}
	subCmd := args[0]

	switch subCmd {
	case "help", "-h", "-help":
		c.Usage()
		return nil
	}

	// Subcommand specific flag parsing
	fs := flag.NewFlagSet(subCmd, flag.ContinueOnError)
	var revision string
	fs.StringVar(&revision, "revision", "", "revision to operate on")
	fs.StringVar(&revision, "rev", "", "revision to operate on (alias)")

	if err := fs.Parse(args[1:]); err != nil {
		return err
	}

	files := fs.Args()
	if len(files) == 0 {
		return fmt.Errorf("no files provided")
	}

	return cli.Locks(subCmd, revision, files...)
}

func (c *Locks) Usage() {
	fmt.Fprintf(os.Stderr, "Usage: %s locks <subcommand> [flags] [files...]\n", os.Args[0])
	fmt.Fprintln(os.Stderr, "Subcommands: lock, unlock, clean, clear, strict, nonstrict")
	fmt.Fprintln(os.Stderr, "Flags:")
	fmt.Fprintln(os.Stderr, "  -revision string")
}

func (c *RootCmd) NewLocks() *Locks {
	return &Locks{
		RootCmd: c,
		Flags:   flag.NewFlagSet("locks", flag.ContinueOnError),
	}
}

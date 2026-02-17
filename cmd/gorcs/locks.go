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
	Flags         *flag.FlagSet
	SubCommands   map[string]Cmd
	CommandAction func(c *Locks) error
}

func (c *Locks) Usage() {
	fmt.Fprintf(os.Stderr, "Usage of %s locks [command] [flags] [files...]:\n", os.Args[0])
	c.Flags.PrintDefaults()
	fmt.Fprintln(os.Stderr, "  Commands:")
	for name := range c.SubCommands {
		fmt.Fprintf(os.Stderr, "    %s\n", name)
	}
}

func (c *Locks) Execute(args []string) error {
	if len(args) > 0 {
		if cmd, ok := c.SubCommands[args[0]]; ok {
			return cmd.Execute(args[1:])
		}
	}
	if err := c.Flags.Parse(args); err != nil {
		return err
	}
	c.Usage()
	return nil
}

func (c *RootCmd) NewLocks() *Locks {
	set := flag.NewFlagSet("locks", flag.ContinueOnError)
	v := &Locks{RootCmd: c, Flags: set, SubCommands: make(map[string]Cmd)}
	set.Usage = v.Usage

	v.SubCommands["lock"] = c.NewLocksLock()
	v.SubCommands["unlock"] = c.NewLocksUnlock()
	v.SubCommands["strict"] = c.NewLocksStrict()
	v.SubCommands["nonstrict"] = c.NewLocksNonStrict()
	v.SubCommands["clean"] = c.NewLocksClean()
	v.SubCommands["clear"] = c.NewLocksClean()

	v.SubCommands["help"] = &InternalCommand{Exec: func(args []string) error { v.Usage(); return nil }, UsageFunc: v.Usage}
	return v
}

// Lock

type LocksLock struct {
	*RootCmd
	Flags    *flag.FlagSet
	revision string
	user     string
}

func (c *RootCmd) NewLocksLock() *LocksLock {
	set := flag.NewFlagSet("lock", flag.ContinueOnError)
	v := &LocksLock{RootCmd: c, Flags: set}
	set.StringVar(&v.revision, "revision", "", "Revision to lock")
	set.StringVar(&v.user, "w", "", "User locking the revision")
	set.Usage = v.Usage
	return v
}

func (c *LocksLock) Usage() {
	fmt.Fprintf(os.Stderr, "Usage of %s locks lock [flags] [files...]:\n", os.Args[0])
	c.Flags.PrintDefaults()
}

func (c *LocksLock) Execute(args []string) error {
	if err := c.Flags.Parse(args); err != nil {
		return err
	}
	files := c.Flags.Args()
	if len(files) == 0 {
		return fmt.Errorf("no files provided")
	}
	return cli.Lock(c.revision, c.user, files...)
}

// Unlock

type LocksUnlock struct {
	*RootCmd
	Flags    *flag.FlagSet
	revision string
	user     string
}

func (c *RootCmd) NewLocksUnlock() *LocksUnlock {
	set := flag.NewFlagSet("unlock", flag.ContinueOnError)
	v := &LocksUnlock{RootCmd: c, Flags: set}
	set.StringVar(&v.revision, "revision", "", "Revision to unlock")
	set.StringVar(&v.user, "w", "", "User unlocking the revision")
	set.Usage = v.Usage
	return v
}

func (c *LocksUnlock) Usage() {
	fmt.Fprintf(os.Stderr, "Usage of %s locks unlock [flags] [files...]:\n", os.Args[0])
	c.Flags.PrintDefaults()
}

func (c *LocksUnlock) Execute(args []string) error {
	if err := c.Flags.Parse(args); err != nil {
		return err
	}
	files := c.Flags.Args()
	if len(files) == 0 {
		return fmt.Errorf("no files provided")
	}
	return cli.Unlock(c.revision, c.user, files...)
}

// Strict

type LocksStrict struct {
	*RootCmd
	Flags *flag.FlagSet
}

func (c *RootCmd) NewLocksStrict() *LocksStrict {
	set := flag.NewFlagSet("strict", flag.ContinueOnError)
	v := &LocksStrict{RootCmd: c, Flags: set}
	set.Usage = v.Usage
	return v
}

func (c *LocksStrict) Usage() {
	fmt.Fprintf(os.Stderr, "Usage of %s locks strict [files...]:\n", os.Args[0])
	c.Flags.PrintDefaults()
}

func (c *LocksStrict) Execute(args []string) error {
	if err := c.Flags.Parse(args); err != nil {
		return err
	}
	files := c.Flags.Args()
	if len(files) == 0 {
		return fmt.Errorf("no files provided")
	}
	return cli.SetStrict(true, files...)
}

// NonStrict

type LocksNonStrict struct {
	*RootCmd
	Flags *flag.FlagSet
}

func (c *RootCmd) NewLocksNonStrict() *LocksNonStrict {
	set := flag.NewFlagSet("nonstrict", flag.ContinueOnError)
	v := &LocksNonStrict{RootCmd: c, Flags: set}
	set.Usage = v.Usage
	return v
}

func (c *LocksNonStrict) Usage() {
	fmt.Fprintf(os.Stderr, "Usage of %s locks nonstrict [files...]:\n", os.Args[0])
	c.Flags.PrintDefaults()
}

func (c *LocksNonStrict) Execute(args []string) error {
	if err := c.Flags.Parse(args); err != nil {
		return err
	}
	files := c.Flags.Args()
	if len(files) == 0 {
		return fmt.Errorf("no files provided")
	}
	return cli.SetStrict(false, files...)
}

// Clean

type LocksClean struct {
	*RootCmd
	Flags    *flag.FlagSet
	revision string
	user     string
}

func (c *RootCmd) NewLocksClean() *LocksClean {
	set := flag.NewFlagSet("clean", flag.ContinueOnError)
	v := &LocksClean{RootCmd: c, Flags: set}
	set.StringVar(&v.revision, "revision", "", "Revision to unlock/clean")
	set.StringVar(&v.user, "w", "", "User owning the lock")
	set.Usage = v.Usage
	return v
}

func (c *LocksClean) Usage() {
	fmt.Fprintf(os.Stderr, "Usage of %s locks clean [flags] [files...]:\n", os.Args[0])
	c.Flags.PrintDefaults()
}

func (c *LocksClean) Execute(args []string) error {
	if err := c.Flags.Parse(args); err != nil {
		return err
	}
	files := c.Flags.Args()
	if len(files) == 0 {
		return fmt.Errorf("no files provided")
	}
	return cli.Clean(c.revision, c.user, files...)
}

package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/arran4/golang-rcs/cmd"
	"github.com/arran4/golang-rcs/internal/cli"
)

var _ Cmd = (*Comment)(nil)

type Comment struct {
	*RootCmd
	Flags         *flag.FlagSet
	SubCommands   map[string]Cmd
	CommandAction func(c *Comment) error
}

type UsageDataComment struct {
	*Comment
	Recursive bool
}

func (c *Comment) Usage() {
	err := executeUsage(os.Stderr, "comment_usage.txt", UsageDataComment{c, false})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error generating usage: %s\n", err)
	}
}

func (c *Comment) UsageRecursive() {
	err := executeUsage(os.Stderr, "comment_usage.txt", UsageDataComment{c, true})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error generating usage: %s\n", err)
	}
}

func (c *Comment) Execute(args []string) error {
	if len(args) > 0 {
		if cmd, ok := c.SubCommands[args[0]]; ok {
			return cmd.Execute(args[1:])
		}
	}
	c.Usage()
	return nil
}

func (c *RootCmd) NewComment() *Comment {
	set := flag.NewFlagSet("comment", flag.ContinueOnError)
	v := &Comment{
		RootCmd:     c,
		Flags:       set,
		SubCommands: make(map[string]Cmd),
	}
	set.Usage = v.Usage

	v.SubCommands["leader"] = v.NewLeader()

	v.SubCommands["help"] = &InternalCommand{
		Exec: func(args []string) error {
			v.Usage()
			return nil
		},
		UsageFunc: v.Usage,
	}
	return v
}

// Leader

var _ Cmd = (*CommentLeader)(nil)

type CommentLeader struct {
	*Comment
	Flags         *flag.FlagSet
	Files         []string
	SubCommands   map[string]Cmd
	CommandAction func(c *CommentLeader) error
}

type UsageDataCommentLeader struct {
	*CommentLeader
	Recursive bool
}

func (c *CommentLeader) Usage() {
	err := executeUsage(os.Stderr, "leader_usage.txt", UsageDataCommentLeader{c, false})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error generating usage: %s\n", err)
	}
}

func (c *CommentLeader) UsageRecursive() {
	err := executeUsage(os.Stderr, "leader_usage.txt", UsageDataCommentLeader{c, true})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error generating usage: %s\n", err)
	}
}

func (c *CommentLeader) Execute(args []string) error {
	if len(args) > 0 {
		if cmd, ok := c.SubCommands[args[0]]; ok {
			return cmd.Execute(args[1:])
		}
	}
	// Parse args as files if no subcommand matched
	remainingArgs := make([]string, 0, len(args))
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if arg == "--" {
			remainingArgs = append(remainingArgs, args[i+1:]...)
			break
		}
		if strings.HasPrefix(arg, "-") {
			// flags? For now assume no flags for leader command itself except help
			if arg == "-help" || arg == "--help" || arg == "-h" {
				c.Usage()
				return nil
			}
			return fmt.Errorf("unknown flag: %s", arg)
		}
		remainingArgs = append(remainingArgs, arg)
	}
	c.Files = remainingArgs

	if c.CommandAction != nil {
		return c.CommandAction(c)
	}

	c.Usage()
	return nil
}

func (c *Comment) NewLeader() *CommentLeader {
	set := flag.NewFlagSet("leader", flag.ContinueOnError)
	v := &CommentLeader{
		Comment:     c,
		Flags:       set,
		SubCommands: make(map[string]Cmd),
	}
	set.Usage = v.Usage

	v.SubCommands["leader-set"] = v.NewLeaderSet()
	v.SubCommands["list"] = v.NewList()

	v.CommandAction = func(c *CommentLeader) error {
		err := cli.CommentLeader(c.Files)
		if err != nil {
			if errors.Is(err, cmd.ErrPrintHelp) || err.Error() == "missing files" {
				c.Usage()
				return nil
			}
			if errors.Is(err, cmd.ErrHelp) {
				fmt.Fprintf(os.Stderr, "Use '%s help' for more information.\n", os.Args[0])
				return nil
			}
			if e, ok := err.(*cmd.ErrExitCode); ok {
				return e
			}
			return fmt.Errorf("leader failed: %w", err)
		}
		return nil
	}

	return v
}

// LeaderSet

var _ Cmd = (*CommentLeaderLeaderSet)(nil)

type CommentLeaderLeaderSet struct {
	*CommentLeader
	Flags         *flag.FlagSet
	LeaderArg     string
	Files         []string
	CommandAction func(c *CommentLeaderLeaderSet) error
}

type UsageDataCommentLeaderLeaderSet struct {
	*CommentLeaderLeaderSet
	Recursive bool
}

func (c *CommentLeaderLeaderSet) Usage() {
	err := executeUsage(os.Stderr, "leader-set_usage.txt", UsageDataCommentLeaderLeaderSet{c, false})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error generating usage: %s\n", err)
	}
}

func (c *CommentLeaderLeaderSet) Execute(args []string) error {
	if len(args) == 0 {
		c.Usage()
		return nil
	}
	// First arg is leader unless it starts with -

	// Check for help flag
	for _, arg := range args {
		if arg == "-help" || arg == "--help" || arg == "-h" {
			c.Usage()
			return nil
		}
	}

	c.LeaderArg = args[0]
	c.Files = args[1:]

	if c.CommandAction != nil {
		return c.CommandAction(c)
	}
	return nil
}

func (c *CommentLeader) NewLeaderSet() *CommentLeaderLeaderSet {
	set := flag.NewFlagSet("leader-set", flag.ContinueOnError)
	v := &CommentLeaderLeaderSet{
		CommentLeader: c,
		Flags:         set,
	}
	set.Usage = v.Usage

	v.CommandAction = func(c *CommentLeaderLeaderSet) error {
		if c.LeaderArg == "" {
			return fmt.Errorf("missing leader argument")
		}
		err := cli.CommentLeaderLeaderSet(c.LeaderArg, c.Files)
		if err != nil {
			if errors.Is(err, cmd.ErrPrintHelp) {
				c.Usage()
				return nil
			}
			if errors.Is(err, cmd.ErrHelp) {
				fmt.Fprintf(os.Stderr, "Use '%s help' for more information.\n", os.Args[0])
				return nil
			}
			if e, ok := err.(*cmd.ErrExitCode); ok {
				return e
			}
			return fmt.Errorf("leader leader-set failed: %w", err)
		}
		return nil
	}
	return v
}

// List

var _ Cmd = (*CommentLeaderList)(nil)

type CommentLeaderList struct {
	*CommentLeader
	Flags         *flag.FlagSet
	CommandAction func(c *CommentLeaderList) error
}

type UsageDataCommentLeaderList struct {
	*CommentLeaderList
	Recursive bool
}

func (c *CommentLeaderList) Usage() {
	err := executeUsage(os.Stderr, "list_usage.txt", UsageDataCommentLeaderList{c, false})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error generating usage: %s\n", err)
	}
}

func (c *CommentLeaderList) Execute(args []string) error {
	// check help
	for _, arg := range args {
		if arg == "-help" || arg == "--help" || arg == "-h" {
			c.Usage()
			return nil
		}
	}
	if c.CommandAction != nil {
		return c.CommandAction(c)
	}
	return nil
}

func (c *CommentLeader) NewList() *CommentLeaderList {
	set := flag.NewFlagSet("list", flag.ContinueOnError)
	v := &CommentLeaderList{
		CommentLeader: c,
		Flags:         set,
	}
	set.Usage = v.Usage

	v.CommandAction = func(c *CommentLeaderList) error {
		err := cli.CommentLeaderList()
		if err != nil {
			if errors.Is(err, cmd.ErrPrintHelp) {
				c.Usage()
				return nil
			}
			if errors.Is(err, cmd.ErrHelp) {
				fmt.Fprintf(os.Stderr, "Use '%s help' for more information.\n", os.Args[0])
				return nil
			}
			if e, ok := err.(*cmd.ErrExitCode); ok {
				return e
			}
			return fmt.Errorf("leader list failed: %w", err)
		}
		return nil
	}
	return v
}

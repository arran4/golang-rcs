package main

import (
	"flag"
	"testing"
)

func TestCi_Execute(t *testing.T) {
	parent := &RootCmd{FlagSet: flag.NewFlagSet("root", flag.ContinueOnError), Commands: map[string]Cmd{}}
	cmd := parent.NewCi()

	called := false
	cmd.CommandAction = func(c *Ci) error {
		called = true
		if !c.quiet {
			t.Fatal("quiet flag not parsed")
		}
		if c.message != "test msg" {
			t.Fatalf("message = %q, want %q", c.message, "test msg")
		}
		if !c.lockSet {
			t.Fatal("lock flag not parsed")
		}
		if c.user != "alice" {
			t.Fatalf("user = %q, want %q", c.user, "alice")
		}
		if len(c.files) != 1 || c.files[0] != "input.txt" {
			t.Fatalf("files = %v", c.files)
		}
		return nil
	}

	if err := cmd.Execute([]string{"-q", "-l", "-mtest msg", "-walice", "input.txt"}); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if !called {
		t.Fatal("command action not called")
	}
}

func TestCi_CompactFlags(t *testing.T) {
	parent := &RootCmd{FlagSet: flag.NewFlagSet("root", flag.ContinueOnError), Commands: map[string]Cmd{}}
	cmd := parent.NewCi()

	called := false
	cmd.CommandAction = func(c *Ci) error {
		called = true
		if !c.forceSet {
			t.Fatal("force flag not parsed")
		}
		if c.forceRev != "1.5" {
			t.Fatalf("forceRev = %q, want %q", c.forceRev, "1.5")
		}
		if c.state != "Rel" {
			t.Fatalf("state = %q, want %q", c.state, "Rel")
		}
		if c.revision != "1.3" {
			t.Fatalf("revision = %q, want %q", c.revision, "1.3")
		}
		return nil
	}

	if err := cmd.Execute([]string{"-f1.5", "-sRel", "-r1.3", "f.txt"}); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if !called {
		t.Fatal("command action not called")
	}
}

func TestCi_SplitFlags(t *testing.T) {
	parent := &RootCmd{FlagSet: flag.NewFlagSet("root", flag.ContinueOnError), Commands: map[string]Cmd{}}
	cmd := parent.NewCi()

	called := false
	cmd.CommandAction = func(c *Ci) error {
		called = true
		if c.message != "hello world" {
			t.Fatalf("message = %q, want %q", c.message, "hello world")
		}
		if c.date != "2020-01-01" {
			t.Fatalf("date = %q, want %q", c.date, "2020-01-01")
		}
		return nil
	}

	if err := cmd.Execute([]string{"-m", "hello world", "-d", "2020-01-01", "f.txt"}); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if !called {
		t.Fatal("command action not called")
	}
}

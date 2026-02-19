package main

import (
	"flag"
	"testing"
)

func TestCo_Execute_Legacy(t *testing.T) {
	parent := &RootCmd{FlagSet: flag.NewFlagSet("root", flag.ContinueOnError), Commands: map[string]Cmd{}}
	cmd := parent.NewCo()

	called := false
	cmd.CommandAction = func(c *Co) error {
		called = true
		if !c.legacyDates {
			t.Fatalf("legacyDates flag not parsed")
		}
		if c.date != "Thu Jan 11 20:00:00 PST 1990" {
			t.Fatalf("date = %q", c.date)
		}
		return nil
	}

	if err := cmd.Execute([]string{"-legacy-zones", "-d", "Thu Jan 11 20:00:00 PST 1990", "input.txt"}); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if !called {
		t.Fatal("command action not called")
	}
}

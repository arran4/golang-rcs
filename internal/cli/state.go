package cli

import (
	"fmt"
	rcs "github.com/arran4/golang-rcs"
	"os"
	"strings"
)

// StateSet is a subcommand `gorcs state set`
//
// Flags:
//
//	rev: Revision to modify
//	state: State to set
//	files: ... List of RCS files to process
func StateSet(rev string, state string, files ...string) error {
	if len(files) == 0 {
		return fmt.Errorf("no files provided")
	}
	for _, file := range files {
		if err := stateSetFile(file, rev, state); err != nil {
			return fmt.Errorf("file %s: %w", file, err)
		}
	}
	return nil
}

func stateSetFile(filename string, rev string, state string) error {
	rcsFile := filename
	if !strings.HasSuffix(rcsFile, ",v") {
		rcsFile += ",v"
	}
	b, err := os.ReadFile(rcsFile)
	if err != nil {
		return fmt.Errorf("read: %w", err)
	}
	parsed, err := rcs.ParseFile(strings.NewReader(string(b)))
	if err != nil {
		return fmt.Errorf("parse: %w", err)
	}

	targetRev := rev
	if targetRev == "" {
		targetRev = parsed.Head
	}
	if targetRev == "" {
		return fmt.Errorf("no revision specified and no head revision found")
	}

	newState := state
	if newState == "" {
		newState = "Exp"
	}

	found := false
	for _, rh := range parsed.RevisionHeads {
		if rh.Revision.String() == targetRev {
			rh.State = rcs.ID(newState)
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("revision %s not found", targetRev)
	}

	if err := os.WriteFile(rcsFile, []byte(parsed.String()), 0644); err != nil {
		return fmt.Errorf("write: %w", err)
	}
	return nil
}

// StateGet is a subcommand `gorcs state get`
//
// Flags:
//
//	rev: Revision to get state of
//	files: ... List of RCS files to process
func StateGet(rev string, files ...string) error {
	if len(files) == 0 {
		return fmt.Errorf("no files provided")
	}
	for _, file := range files {
		if err := stateGetFile(file, rev); err != nil {
			return fmt.Errorf("file %s: %w", file, err)
		}
	}
	return nil
}

func stateGetFile(filename string, rev string) error {
	rcsFile := filename
	if !strings.HasSuffix(rcsFile, ",v") {
		rcsFile += ",v"
	}
	b, err := os.ReadFile(rcsFile)
	if err != nil {
		return fmt.Errorf("read: %w", err)
	}
	parsed, err := rcs.ParseFile(strings.NewReader(string(b)))
	if err != nil {
		return fmt.Errorf("parse: %w", err)
	}

	targetRev := rev
	if targetRev == "" {
		targetRev = parsed.Head
	}
	if targetRev == "" {
		return fmt.Errorf("no revision specified and no head revision found")
	}

	found := false
	for _, rh := range parsed.RevisionHeads {
		if rh.Revision.String() == targetRev {
			fmt.Println(rh.State)
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("revision %s not found", targetRev)
	}
	return nil
}

// StateList is a subcommand `gorcs state list`
//
// Flags:
//
//	files: ... List of RCS files to process
func StateList(files ...string) error {
	if len(files) == 0 {
		return fmt.Errorf("no files provided")
	}
	for _, file := range files {
		if err := stateListFile(file); err != nil {
			return fmt.Errorf("file %s: %w", file, err)
		}
	}
	return nil
}

func stateListFile(filename string) error {
	rcsFile := filename
	if !strings.HasSuffix(rcsFile, ",v") {
		rcsFile += ",v"
	}
	b, err := os.ReadFile(rcsFile)
	if err != nil {
		return fmt.Errorf("read: %w", err)
	}
	parsed, err := rcs.ParseFile(strings.NewReader(string(b)))
	if err != nil {
		return fmt.Errorf("parse: %w", err)
	}

	for _, rh := range parsed.RevisionHeads {
		fmt.Printf("%s %s\n", rh.Revision, rh.State)
	}
	return nil
}

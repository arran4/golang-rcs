package cli

import (
	"fmt"
	"os"
	"strings"

	rcs "github.com/arran4/golang-rcs"
)

// StateAlter is a subcommand `gorcs state alter`
//
// Flags:
//
//	state: -state new state (defaults to Exp if not provided and resetting)
//	revision: -rev revision to change
//	files: ... List of working files to process
func StateAlter(state string, revision string, files ...string) error {
	for _, file := range files {
		rcsFile := file
		if !strings.HasSuffix(rcsFile, ",v") {
			rcsFile += ",v"
		}

		b, err := os.ReadFile(rcsFile)
		if err != nil {
			return fmt.Errorf("read %s: %w", rcsFile, err)
		}

		parsedFile, err := rcs.ParseFile(strings.NewReader(string(b)))
		if err != nil {
			return fmt.Errorf("parse %s: %w", rcsFile, err)
		}

		rev := revision
		if rev == "" {
			rev = parsedFile.Head
		}

		st := state
		if st == "" {
			st = "Exp"
		}

		if err := parsedFile.SetState(rev, st); err != nil {
			return fmt.Errorf("set state in %s: %w", rcsFile, err)
		}

		// Write back the file
		if err := os.WriteFile(rcsFile, []byte(parsedFile.String()), 0644); err != nil {
			return fmt.Errorf("write %s: %w", rcsFile, err)
		}
	}
	return nil
}

// StateGet is a subcommand `gorcs state get`
//
// Flags:
//
//	revision: -rev revision to get
//	files: ... List of working files to process
func StateGet(revision string, files ...string) error {
	for _, file := range files {
		rcsFile := file
		if !strings.HasSuffix(rcsFile, ",v") {
			rcsFile += ",v"
		}

		f, err := os.Open(rcsFile)
		if err != nil {
			return fmt.Errorf("open %s: %w", rcsFile, err)
		}

		parsedFile, err := rcs.ParseFile(f)
		if err := f.Close(); err != nil {
			return fmt.Errorf("close %s: %w", rcsFile, err)
		}
		if err != nil {
			return fmt.Errorf("parse %s: %w", rcsFile, err)
		}

		rev := revision
		if rev == "" {
			rev = parsedFile.Head
		}

		st, err := parsedFile.GetState(rev)
		if err != nil {
			return fmt.Errorf("get state in %s: %w", rcsFile, err)
		}

		fmt.Println(st)
	}
	return nil
}

// StateLs is a subcommand `gorcs state ls`
//
// Flags:
//
//	files: ... List of working files to process
func StateLs(files ...string) error {
	for _, file := range files {
		rcsFile := file
		if !strings.HasSuffix(rcsFile, ",v") {
			rcsFile += ",v"
		}

		f, err := os.Open(rcsFile)
		if err != nil {
			return fmt.Errorf("open %s: %w", rcsFile, err)
		}

		parsedFile, err := rcs.ParseFile(f)
		if err := f.Close(); err != nil {
			return fmt.Errorf("close %s: %w", rcsFile, err)
		}
		if err != nil {
			return fmt.Errorf("parse %s: %w", rcsFile, err)
		}

		if len(files) > 1 {
			fmt.Printf("File: %s\n", file)
		}
		states := parsedFile.ListStates()
		for _, s := range states {
			fmt.Printf("%s %s\n", s.Revision, s.State)
		}
		if len(files) > 1 {
			fmt.Println()
		}
	}
	return nil
}

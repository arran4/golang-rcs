package cli

import (
	"fmt"
	"os"
	"strings"

	rcs "github.com/arran4/golang-rcs"
)

// LocksLockSet is a subcommand `gorcs locks lock-set`
// TODO: change to set when go-subcommand v0.0.21 is released
//
// Flags:
//
//	revision: -rev revision to lock
//	user: -u user to lock for
//	files: ... List of working files to process
func LocksLockSet(revision, user string, files ...string) error {
	if revision == "" {
		return fmt.Errorf("revision required")
	}
	if user == "" {
		return fmt.Errorf("user required")
	}
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

		parsedFile.SetLock(user, revision)

		if err := os.WriteFile(rcsFile, []byte(parsedFile.String()), 0644); err != nil {
			return fmt.Errorf("write %s: %w", rcsFile, err)
		}
	}
	return nil
}

// LocksRemove is a subcommand `gorcs locks remove`
//
// Flags:
//
//	revision: -rev revision to unlock
//	user: -u user to unlock for
//	files: ... List of working files to process
func LocksRemove(revision, user string, files ...string) error {
	if revision == "" {
		return fmt.Errorf("revision required")
	}
	if user == "" {
		return fmt.Errorf("user required")
	}
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

		parsedFile.ClearLock(user, revision)

		if err := os.WriteFile(rcsFile, []byte(parsedFile.String()), 0644); err != nil {
			return fmt.Errorf("write %s: %w", rcsFile, err)
		}
	}
	return nil
}

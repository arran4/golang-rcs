package cli

import (
	"fmt"
	"os"
	"strings"

	rcs "github.com/arran4/golang-rcs"
)

// LogMessageChange is a subcommand `gorcs log message change`
//
// Changes the log message for a specific revision in the given files.
//
// Flags:
//
//	rev: -rev revision to change log message for
//	m: -m new log message
//	files: ... List of files to process
func LogMessageChange(rev, m string, files ...string) error {
	if rev == "" {
		return fmt.Errorf("revision required")
	}
	for _, file := range files {
		rcsFile := file
		if !strings.HasSuffix(rcsFile, ",v") {
			rcsFile += ",v"
		}
		b, err := os.ReadFile(rcsFile)
		if err != nil {
			return fmt.Errorf("read %s: %w", rcsFile, err)
		}
		parsed, err := rcs.ParseFile(strings.NewReader(string(b)))
		if err != nil {
			return fmt.Errorf("parse %s: %w", rcsFile, err)
		}

		if err := parsed.SetLog(rev, m); err != nil {
			return fmt.Errorf("set log for %s: %w", rcsFile, err)
		}

		if err := os.WriteFile(rcsFile, []byte(parsed.String()), 0644); err != nil {
			return fmt.Errorf("write %s: %w", rcsFile, err)
		}
		fmt.Printf("Updated log message for %s revision %s\n", rcsFile, rev)
	}
	return nil
}

// LogMessagePrint is a subcommand `gorcs log message print`
//
// Prints the log message for a specific revision in the given files.
//
// Flags:
//
//	rev: -rev revision to print log message for
//	files: ... List of files to process
func LogMessagePrint(rev string, files ...string) error {
	if rev == "" {
		return fmt.Errorf("revision required")
	}
	for _, file := range files {
		rcsFile := file
		if !strings.HasSuffix(rcsFile, ",v") {
			rcsFile += ",v"
		}
		b, err := os.ReadFile(rcsFile)
		if err != nil {
			return fmt.Errorf("read %s: %w", rcsFile, err)
		}
		parsed, err := rcs.ParseFile(strings.NewReader(string(b)))
		if err != nil {
			return fmt.Errorf("parse %s: %w", rcsFile, err)
		}

		log, err := parsed.GetLog(rev)
		if err != nil {
			return fmt.Errorf("get log for %s: %w", rcsFile, err)
		}

		fmt.Printf("File: %s\nRevision: %s\nMessage:\n%s\n", rcsFile, rev, log)
	}
	return nil
}

// LogMessageList is a subcommand `gorcs log message list`
//
// Lists all log messages in the given files.
//
// Flags:
//
//	files: ... List of files to process
func LogMessageList(files ...string) error {
	for _, file := range files {
		rcsFile := file
		if !strings.HasSuffix(rcsFile, ",v") {
			rcsFile += ",v"
		}
		b, err := os.ReadFile(rcsFile)
		if err != nil {
			return fmt.Errorf("read %s: %w", rcsFile, err)
		}
		parsed, err := rcs.ParseFile(strings.NewReader(string(b)))
		if err != nil {
			return fmt.Errorf("parse %s: %w", rcsFile, err)
		}

		fmt.Printf("File: %s\n", rcsFile)
		for _, rc := range parsed.RevisionContents {
			fmt.Printf("Revision: %s\nMessage: %s\n", rc.Revision, rc.Log)
		}
	}
	return nil
}

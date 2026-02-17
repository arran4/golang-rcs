package cli

import (
	"fmt"
	"os"
	"strings"

	rcs "github.com/arran4/golang-rcs"
)

// LogMessageChange changes the log message for a specific revision in the given files.
func LogMessageChange(revision, message string, files ...string) error {
	if revision == "" {
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

		if err := parsed.SetLog(revision, message); err != nil {
			return fmt.Errorf("set log for %s: %w", rcsFile, err)
		}

		if err := os.WriteFile(rcsFile, []byte(parsed.String()), 0644); err != nil {
			return fmt.Errorf("write %s: %w", rcsFile, err)
		}
		fmt.Printf("Updated log message for %s revision %s\n", rcsFile, revision)
	}
	return nil
}

// LogMessagePrint prints the log message for a specific revision in the given files.
func LogMessagePrint(revision string, files ...string) error {
	if revision == "" {
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

		log, err := parsed.GetLog(revision)
		if err != nil {
			return fmt.Errorf("get log for %s: %w", rcsFile, err)
		}

		fmt.Printf("File: %s\nRevision: %s\nMessage:\n%s\n", rcsFile, revision, log)
	}
	return nil
}

// LogMessageList lists all log messages in the given files.
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

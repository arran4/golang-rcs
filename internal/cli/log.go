package cli

import (
	"fmt"
	"os"
	"strings"

	rcs "github.com/arran4/golang-rcs"
)

// LogMessageChange changes the log message for a specific revision in the given RCS files.
func LogMessageChange(revision, message string, files ...string) error {
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

		if err := parsedFile.ChangeLogMessage(revision, message); err != nil {
			return fmt.Errorf("change log message in %s: %w", rcsFile, err)
		}

		// Write back the file
		if err := os.WriteFile(rcsFile, []byte(parsedFile.String()), 0644); err != nil {
			return fmt.Errorf("write %s: %w", rcsFile, err)
		}
	}
	return nil
}

// LogMessagePrint prints the log message for a specific revision in the given RCS files.
func LogMessagePrint(revision string, files ...string) error {
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

		msg, err := parsedFile.GetLogMessage(revision)
		if err != nil {
			return fmt.Errorf("get log message in %s: %w", rcsFile, err)
		}

		fmt.Printf("File: %s Revision: %s\n%s\n", file, revision, msg)
	}
	return nil
}

// LogMessageList lists all log messages in the given RCS files.
func LogMessageList(files ...string) error {
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

		fmt.Printf("File: %s\n", file)
		logs := parsedFile.ListLogMessages()
		for _, log := range logs {
			fmt.Printf("Revision: %s\n%s\n", log.Revision, log.Log)
		}
		fmt.Println()
	}
	return nil
}

package cli

import (
	"fmt"
	"os"
	"strings"

	rcs "github.com/arran4/golang-rcs"
)

func Locks(subCommand, revision string, files ...string) error {
	for _, file := range files {
		if err := locksFile(subCommand, revision, file); err != nil {
			return err
		}
	}
	return nil
}

func locksFile(subCommand, revision string, workingFile string) error {
	rcsFile := workingFile
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

	changed := false
	user := currentLoggedInUser()

	switch subCommand {
	case "lock":
		if revision == "" {
			return fmt.Errorf("lock requires revision")
		}
		if parsed.SetLock(user, revision) {
			changed = true
		}
	case "unlock":
		if revision == "" {
			return fmt.Errorf("unlock requires revision")
		}
		if parsed.ClearLock(user, revision) {
			changed = true
		}
	case "strict":
		if !parsed.Strict {
			parsed.Strict = true
			changed = true
		}
	case "nonstrict":
		if parsed.Strict {
			parsed.Strict = false
			changed = true
		}
	case "clean", "clear":
		wb, err := os.ReadFile(workingFile)
		if err != nil {
			return fmt.Errorf("read working file %s: %w", workingFile, err)
		}

		targetRev := revision
		if targetRev == "" {
			targetRev = parsed.Head
		}

		verdict, err := parsed.Checkout(user, rcs.WithRevision(targetRev))
		if err != nil {
			return fmt.Errorf("checkout for clean check: %w", err)
		}

		if string(wb) == verdict.Content {
			if parsed.ClearLock(user, targetRev) {
				changed = true
			}
		} else {
			return fmt.Errorf("working file %s is modified", workingFile)
		}

	default:
		return fmt.Errorf("unknown subcommand: %s", subCommand)
	}

	if changed {
		if err := os.WriteFile(rcsFile, []byte(parsed.String()), 0644); err != nil {
			return fmt.Errorf("write %s: %w", rcsFile, err)
		}
	}
	return nil
}

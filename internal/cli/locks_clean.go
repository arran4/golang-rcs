package cli

import (
	"fmt"
	"os"
	"strings"

	rcs "github.com/arran4/golang-rcs"
)

// LocksClean compares the working file with the revision and optionally unlocks it.
// If the working file matches the revision, it is removed.
// Returns true if any file was found to be dirty (not matching the revision).
//
// Flags:
//
//	revision: -r revision to check against
//	unlock: -u unlock revision
//	user: -w user for lock operations
//	quiet: -q suppress output
//	files: ... List of working files to process
func LocksClean(revision string, unlock bool, user string, quiet bool, files ...string) (bool, error) {
	if user == "" {
		user = currentLoggedInUser()
	}
	if len(files) == 0 {
		return false, fmt.Errorf("no files provided")
	}

	anyDirty := false

	for _, workingFile := range files {
		rcsFile := workingFile
		if !strings.HasSuffix(rcsFile, ",v") {
			rcsFile += ",v"
		}

		// Read working file
		workingContent, err := os.ReadFile(workingFile)
		if err != nil {
			// If working file missing, maybe ignore or error?
			// rcsclean behavior implies working file exists.
			return false, fmt.Errorf("read %s: %w", workingFile, err)
		}

		// Read RCS file
		b, err := os.ReadFile(rcsFile)
		if err != nil {
			return false, fmt.Errorf("read %s: %w", rcsFile, err)
		}
		parsed, err := rcs.ParseFile(strings.NewReader(string(b)))
		if err != nil {
			return false, fmt.Errorf("parse %s: %w", rcsFile, err)
		}

		ops := make([]any, 0)
		if revision != "" {
			ops = append(ops, rcs.WithRevision(revision))
		}
		if unlock {
			ops = append(ops, rcs.WithClearLock)
		}

		verdict, err := parsed.Clean(user, workingContent, ops...)
		if err != nil {
			return false, fmt.Errorf("clean %s: %w", workingFile, err)
		}

		if verdict.Clean {
			if err := os.Remove(workingFile); err != nil {
				return false, fmt.Errorf("remove %s: %w", workingFile, err)
			}
			if !quiet {
				fmt.Printf("removed %s\n", workingFile)
			}
		} else {
			anyDirty = true
			// rcsclean usually doesn't say "dirty", it just doesn't remove it.
			// But maybe we want to know?
			// "rcsclean: input.txt is not clean" ??
			// Standard rcsclean is silent about dirty files unless -n is used?
			// I'll keep it silent matching standard, but exit code will reflect it.
		}

		if verdict.Unlocked {
			// Save RCS file
			if err := os.WriteFile(rcsFile, []byte(parsed.String()), 0644); err != nil {
				return false, fmt.Errorf("write %s: %w", rcsFile, err)
			}
			if !quiet {
				fmt.Printf("unlocked %s\n", rcsFile)
			}
		}
	}
	return anyDirty, nil
}

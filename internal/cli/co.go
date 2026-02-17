package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	rcs "github.com/arran4/golang-rcs"
)

type COVerdict struct {
	File          string
	RCSFile       string
	Revision      string
	FileModified  bool
	LockSet       bool
	LockCleared   bool
	CheckedOutLen int
}

// Co performs checkout operations over one or more working files.
//
// Flags:
//
//	revision: -r revision to check out
//	lock: -l lock checked-out revision
//	unlock: -u unlock checked-out revision
//	force: -f force overwrite of working file
//	user: -w user for lock operations
//	quiet: -q suppress status output
//	files: ... List of working files to process
func Co(revision string, lock, unlock, force bool, user string, quiet bool, files ...string) error {
	if lock && unlock {
		return fmt.Errorf("cannot combine -l and -u")
	}
	if user == "" {
		user = currentLoggedInUser()
	}
	if len(files) == 0 {
		return fmt.Errorf("no files provided")
	}
	for _, file := range files {
		result, err := coFile(revision, lock, unlock, force, user, quiet, file)
		if err != nil {
			return err
		}
		if !quiet {
			fmt.Printf("co %s: rev=%s bytes=%d lock-set=%t lock-cleared=%t rcs-updated=%t\n",
				filepath.Base(result.File),
				result.Revision,
				result.CheckedOutLen,
				result.LockSet,
				result.LockCleared,
				result.FileModified,
			)
		}
	}
	return nil
}

func coFile(revision string, lock, unlock, force bool, user string, quiet bool, workingFile string) (COVerdict, error) {
	rcsFile := workingFile
	if !strings.HasSuffix(rcsFile, ",v") {
		rcsFile += ",v"
	}
	b, err := os.ReadFile(rcsFile)
	if err != nil {
		return COVerdict{}, fmt.Errorf("read %s: %w", rcsFile, err)
	}
	parsed, err := rcs.ParseFile(strings.NewReader(string(b)))
	if err != nil {
		return COVerdict{}, fmt.Errorf("parse %s: %w", rcsFile, err)
	}

	ops := make([]any, 0, 2)
	if revision != "" {
		ops = append(ops, rcs.WithRevision(revision))
	}
	if lock {
		ops = append(ops, rcs.WithSetLock)
	} else if unlock {
		ops = append(ops, rcs.WithClearLock)
	}

	verdict, err := parsed.Checkout(user, ops...)
	if err != nil {
		return COVerdict{}, fmt.Errorf("co %s: %w", rcsFile, err)
	}

	if !force {
		if info, err := os.Stat(workingFile); err == nil {
			if info.Mode().Perm()&0200 != 0 {
				return COVerdict{}, fmt.Errorf("writable %s exists; remove it or use -f", workingFile)
			}
		}
	}
	_ = os.Remove(workingFile)

	if err := os.WriteFile(workingFile, []byte(verdict.Content), 0644); err != nil {
		return COVerdict{}, fmt.Errorf("write %s: %w", workingFile, err)
	}
	if verdict.FileModified {
		if err := os.WriteFile(rcsFile, []byte(parsed.String()), 0644); err != nil {
			return COVerdict{}, fmt.Errorf("write %s: %w", rcsFile, err)
		}
	}
	return COVerdict{
		File:          workingFile,
		RCSFile:       rcsFile,
		Revision:      verdict.Revision,
		FileModified:  verdict.FileModified,
		LockSet:       verdict.LockSet,
		LockCleared:   verdict.LockCleared,
		CheckedOutLen: len(verdict.Content),
	}, nil
}

func currentLoggedInUser() string {
	if u := strings.TrimSpace(os.Getenv("USER")); u != "" {
		return u
	}
	if u := strings.TrimSpace(os.Getenv("USERNAME")); u != "" {
		return u
	}
	return "unknown"
}

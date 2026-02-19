package cli

import (
	"fmt"
	rcs "github.com/arran4/golang-rcs"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// BranchesDefaultBranchDefaultSet is a subcommand `gorcs branches default branch-default-set`
// TODO: change to set when go-subcommand v0.0.21 is released
//
// Flags:
//
//	name: default branch name/revision to set
//	files: ... List of working files to process
func BranchesDefaultBranchDefaultSet(name string, files ...string) error {
	if strings.TrimSpace(name) == "" {
		return fmt.Errorf("default branch name is required")
	}
	defaultBranch, err := normalizeDefaultBranch(name)
	if err != nil {
		return err
	}
	if len(files) == 0 {
		return fmt.Errorf("no files provided")
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
		parsed.Branch = defaultBranch
		if err := os.WriteFile(rcsFile, []byte(parsed.String()), 0644); err != nil {
			return fmt.Errorf("write %s: %w", rcsFile, err)
		}
		fmt.Printf("set default branch for %s to %s\n", filepath.Base(file), defaultBranch)
	}
	return nil
}

func normalizeDefaultBranch(name string) (string, error) {
	parts := strings.Split(name, ".")
	if len(parts) < 3 {
		return "", fmt.Errorf("invalid default branch name: %q", name)
	}
	for _, p := range parts {
		if p == "" {
			return "", fmt.Errorf("invalid default branch name: %q", name)
		}
		if _, err := strconv.Atoi(p); err != nil {
			return "", fmt.Errorf("invalid default branch name: %q", name)
		}
	}
	if len(parts)%2 == 0 {
		parts = parts[:len(parts)-1]
	}
	return strings.Join(parts, "."), nil
}

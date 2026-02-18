package cli

import (
	"fmt"
	rcs "github.com/arran4/golang-rcs"
	"os"
)

// AccessListCopy is a subcommand `gorcs access-list copy`
//
// Flags:
//
//	from: -from Source RCS file to copy access list from
//	files: ... List of working files/RCS files to update
func AccessListCopy(fromFile string, toFiles ...string) error {
	fromF, err := OpenFile(fromFile, false)
	if err != nil {
		return fmt.Errorf("error opening from file %s: %w", fromFile, err)
	}
	defer func() {
		_ = fromF.Close()
	}()

	fromParsed, err := rcs.ParseFile(fromF)
	if err != nil {
		return fmt.Errorf("error parsing from file %s: %w", fromFile, err)
	}

	for _, toFile := range toFiles {
		if err := copyAccessListToFile(fromParsed, toFile); err != nil {
			return fmt.Errorf("error copying access list to %s: %w", toFile, err)
		}
	}
	return nil
}

func copyAccessListToFile(fromParsed *rcs.File, toFile string) error {
	toF, err := OpenFile(toFile, false)
	if err != nil {
		return fmt.Errorf("error opening to file %s: %w", toFile, err)
	}
	// We need to read the whole file, modify it, and write it back.
	// Since OpenFile returns a ReadCloser, we can parse it.
	// But we need to close it before writing back?
	// Or write to a temp file and rename?
	// For now, let's just read, close, and then write.

	toParsed, err := rcs.ParseFile(toF)
	if closeErr := toF.Close(); closeErr != nil {
		return fmt.Errorf("error closing to file %s: %w", toFile, closeErr)
	}
	if err != nil {
		return fmt.Errorf("error parsing to file %s: %w", toFile, err)
	}

	toParsed.CopyAccessList(fromParsed)

	// Write back
	content := toParsed.String()
	if err := os.WriteFile(toFile, []byte(content), 0644); err != nil {
		return fmt.Errorf("error writing to file %s: %w", toFile, err)
	}

	fmt.Printf("Updated access list for %s\n", toFile)
	return nil
}

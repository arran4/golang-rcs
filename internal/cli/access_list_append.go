package cli

import (
	"fmt"
	rcs "github.com/arran4/golang-rcs"
	"os"
)

// AccessListAppend is a subcommand `gorcs access-list append`
//
// Flags:
//
//	from: -from Source RCS file to append access list from
//	files: ... List of working files/RCS files to update
func AccessListAppend(from string, toFiles ...string) error {
	fromF, err := OpenFile(from, false)
	if err != nil {
		return fmt.Errorf("error opening from file %s: %w", from, err)
	}
	defer func() {
		_ = fromF.Close()
	}()

	fromParsed, err := rcs.ParseFile(fromF)
	if err != nil {
		return fmt.Errorf("error parsing from file %s: %w", from, err)
	}

	for _, toFile := range toFiles {
		if err := appendAccessListToFile(fromParsed, toFile); err != nil {
			return fmt.Errorf("error appending access list to %s: %w", toFile, err)
		}
	}
	return nil
}

func appendAccessListToFile(fromParsed *rcs.File, toFile string) error {
	toF, err := OpenFile(toFile, false)
	if err != nil {
		return fmt.Errorf("error opening to file %s: %w", toFile, err)
	}

	toParsed, err := rcs.ParseFile(toF)
	if closeErr := toF.Close(); closeErr != nil {
		return fmt.Errorf("error closing to file %s: %w", toFile, closeErr)
	}
	if err != nil {
		return fmt.Errorf("error parsing to file %s: %w", toFile, err)
	}

	toParsed.AppendAccessList(fromParsed)

	// Write back
	content := toParsed.String()
	if err := os.WriteFile(toFile, []byte(content), 0644); err != nil {
		return fmt.Errorf("error writing to file %s: %w", toFile, err)
	}

	fmt.Printf("Updated access list for %s\n", toFile)
	return nil
}

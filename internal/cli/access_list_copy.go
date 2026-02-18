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
//	fromFile: -from Source RCS file to copy access list from
//	toFiles: ... List of target files
func AccessListCopy(fromFile string, toFiles ...string) error {
	fromF, err := OpenFile(fromFile, false)
	if err != nil {
		return fmt.Errorf("failed to open from file %s: %w", fromFile, err)
	}
	defer func() {
		_ = fromF.Close()
	}()

	fromRCS, err := rcs.ParseFile(fromF)
	if err != nil {
		return fmt.Errorf("failed to parse from file %s: %w", fromFile, err)
	}

	for _, toFile := range toFiles {
		if err := copyAccessListTo(fromRCS, toFile); err != nil {
			return fmt.Errorf("failed to copy to %s: %w", toFile, err)
		}
	}

	return nil
}

func copyAccessListTo(fromRCS *rcs.File, toFile string) error {
	f, err := os.Open(toFile)
	if err != nil {
		return fmt.Errorf("failed to open file %s: %w", toFile, err)
	}

	toRCS, err := rcs.ParseFile(f)
	if closeErr := f.Close(); closeErr != nil {
		return fmt.Errorf("failed to close file %s: %w", toFile, closeErr)
	}
	if err != nil {
		return fmt.Errorf("failed to parse file %s: %w", toFile, err)
	}

	toRCS.CopyAccessList(fromRCS)

	content := toRCS.String()

	if err := writeOutput(toFile, []byte(content), true); err != nil {
		return fmt.Errorf("failed to write output to %s: %w", toFile, err)
	}

	return nil
}

// AccessListAppend is a subcommand `gorcs access-list append`
//
// Flags:
//
//	fromFile: -from Source RCS file to append access list from
//	toFiles: ... List of target files
func AccessListAppend(fromFile string, toFiles ...string) error {
	fromF, err := OpenFile(fromFile, false)
	if err != nil {
		return fmt.Errorf("failed to open from file %s: %w", fromFile, err)
	}
	defer func() {
		_ = fromF.Close()
	}()

	fromRCS, err := rcs.ParseFile(fromF)
	if err != nil {
		return fmt.Errorf("failed to parse from file %s: %w", fromFile, err)
	}

	for _, toFile := range toFiles {
		if err := appendAccessListTo(fromRCS, toFile); err != nil {
			return fmt.Errorf("failed to append to %s: %w", toFile, err)
		}
	}

	return nil
}

func appendAccessListTo(fromRCS *rcs.File, toFile string) error {
	f, err := os.Open(toFile)
	if err != nil {
		return fmt.Errorf("failed to open file %s: %w", toFile, err)
	}

	toRCS, err := rcs.ParseFile(f)
	if closeErr := f.Close(); closeErr != nil {
		return fmt.Errorf("failed to close file %s: %w", toFile, closeErr)
	}
	if err != nil {
		return fmt.Errorf("failed to parse file %s: %w", toFile, err)
	}

	toRCS.AppendAccessList(fromRCS)

	content := toRCS.String()

	if err := writeOutput(toFile, []byte(content), true); err != nil {
		return fmt.Errorf("failed to write output to %s: %w", toFile, err)
	}

	return nil
}

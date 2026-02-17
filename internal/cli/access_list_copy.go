package cli

import (
	"fmt"
	rcs "github.com/arran4/golang-rcs"
	"os"
)

// AccessListCopy copies the access list from one RCS file to others.
func AccessListCopy(fromFile string, toFiles ...string) error {
	fromF, err := OpenFile(fromFile, false)
	if err != nil {
		return fmt.Errorf("failed to open from file %s: %w", fromFile, err)
	}
	defer fromF.Close()

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
	f.Close()
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

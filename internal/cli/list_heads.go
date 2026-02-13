package cli

import (
	"fmt"
	rcs "github.com/arran4/golang-rcs"
	"os"
	"time"
)

// ListHeads is a subcommand `gorcs list-heads`
//
// Flags:
//
//	files: ... List of files to process
func ListHeads(files ...string) error {
	for _, f := range files {
		if err := listHeadsFile(f); err != nil {
			return err
		}
	}
	return nil
}

func listHeadsFile(fn string) error {
	f, err := os.Open(fn)
	if err != nil {
		return fmt.Errorf("error with file: %w", err)
	}
	defer func() {
		_ = f.Close()
	}()
	fmt.Println("Parsing: ", fn)
	r, err := rcs.ParseFile(f)
	if err != nil {
		return fmt.Errorf("error parsing %s: %w", fn, err)
	}
	for _, rh := range r.RevisionHeads {
		dt, _ := rh.Date.DateTime()
		fmt.Printf("%s on %s by %s\n", rh.Revision, dt.In(time.Local), rh.Author)
	}
	return nil
}

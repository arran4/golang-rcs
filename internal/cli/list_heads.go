package cli

import (
	"fmt"
	rcs "github.com/arran4/golang-rcs"
	"time"
)

// ListHeads is a subcommand `gorcs list-heads`
//
// Flags:
//
//	mmap: -m --mmap Use mmap to read file
//	files: ... List of files to process
func ListHeads(useMmap bool, files ...string) error {
	var err error
	if files, err = ensureFiles(files); err != nil {
		return err
	}
	for _, f := range files {
		if err := listHeadsFile(f, useMmap); err != nil {
			return err
		}
	}
	return nil
}

func listHeadsFile(fn string, useMmap bool) error {
	f, closeFunc, err := OpenFile(fn, useMmap)
	if err != nil {
		return fmt.Errorf("error with file: %w", err)
	}
	defer func() {
		_ = closeFunc()
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

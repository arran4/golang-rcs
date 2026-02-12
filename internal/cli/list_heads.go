package cli

import (
	"fmt"
	rcs "github.com/arran4/golang-rcs"
	"log"
	"os"
	"time"
)

// ListHeads is a subcommand `gorcs list-heads`
//
// Flags:
//
//	files: ... List of files to process
func ListHeads(files ...string) {
	for _, f := range files {
		listHeadsFile(f)
	}
}

func listHeadsFile(fn string) {
	f, err := os.Open(fn)
	if err != nil {
		log.Panicf("Error with file: %s", err)
	}
	defer func() {
		if err = f.Close(); err != nil {
			log.Panicf("Error closing file; %s: %s", fn, err)
		}
	}()
	fmt.Println("Parsing: ", fn)
	r, err := rcs.ParseFile(f)
	if err != nil {
		log.Panicf("Error parsing %s: %s", fn, err)
	}
	for _, rh := range r.RevisionHeads {
		fmt.Printf("%s on %s by %s\n", rh.Revision, rh.Date.In(time.Local), rh.Author)
	}
}

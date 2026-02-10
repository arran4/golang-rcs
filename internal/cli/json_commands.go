package cli

import (
	"encoding/json"
	"fmt"
	rcs "github.com/arran4/golang-rcs"
	"io"
	"log"
	"os"
)

// ToJson is a subcommand `gorcs to-json`
//
// Flags:
//   files: ... List of files to process, or - for stdin
func ToJson(files ...string) {
	for _, fn := range files {
		var f io.Reader
		var err error
		if fn == "-" {
			f = os.Stdin
		} else {
			file, err := os.Open(fn)
			if err != nil {
				log.Panicf("Error with file %s: %s", fn, err)
			}
			defer func() {
				if err = file.Close(); err != nil {
					log.Panicf("Error closing file; %s: %s", fn, err)
				}
			}()
			f = file
		}
		r, err := rcs.ParseFile(f)
		if err != nil {
			log.Panicf("Error parsing %s: %s", fn, err)
		}
		b, err := json.Marshal(r)
		if err != nil {
			log.Panicf("Error serializing %s: %s", fn, err)
		}
		fmt.Printf("%s", b)
	}
}

// FromJson is a subcommand `gorcs from-json`
//
// Flags:
//   files: ... List of files to process, or - for stdin
func FromJson(files ...string) {
	for _, fn := range files {
		var f io.Reader
		if fn == "-" {
			f = os.Stdin
		} else {
			file, err := os.Open(fn)
			if err != nil {
				log.Panicf("Error with file %s: %s", fn, err)
			}
			defer func() {
				if err = file.Close(); err != nil {
					log.Panicf("Error closing file; %s: %s", fn, err)
				}
			}()
			f = file
		}
		var r rcs.File
		if err := json.NewDecoder(f).Decode(&r); err != nil {
			log.Panicf("Error parsing %s: %s", fn, err)
		}
		fmt.Print(r.String())
	}
}

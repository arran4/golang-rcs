package cli

import (
	"encoding/json"
	"fmt"
	rcs "github.com/arran4/golang-rcs"
	"io"
	"log"
	"os"
	"strings"
)

// ToJson is a subcommand `gorcs to-json`
//
// Flags:
//
//	output: -o --output Output file path
//	force: -f --force Force overwrite output
//	files: ... List of files to process, or - for stdin
func ToJson(output string, force bool, files ...string) {
	if output != "" && len(files) > 1 {
		log.Panicf("Cannot specify output file with multiple input files")
	}
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
		r, err := rcs.ParseFile(f)
		if err != nil {
			log.Panicf("Error parsing %s: %s", fn, err)
		}
		b, err := json.Marshal(r)
		if err != nil {
			log.Panicf("Error serializing %s: %s", fn, err)
		}

		if output == "-" {
			fmt.Printf("%s", b)
		} else if output != "" {
			writeOutput(output, b, force)
		} else if fn == "-" {
			fmt.Printf("%s", b)
		} else {
			// Default output: filename + .json
			outPath := fn + ".json"
			writeOutput(outPath, b, force)
		}
	}
}

// FromJson is a subcommand `gorcs from-json`
//
// Flags:
//
//	output: -o --output Output file path
//	force: -f --force Force overwrite output
//	files: ... List of files to process, or - for stdin
func FromJson(output string, force bool, files ...string) {
	if output != "" && len(files) > 1 {
		log.Panicf("Cannot specify output file with multiple input files")
	}
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

		outBytes := []byte(r.String())

		if output == "-" {
			fmt.Print(string(outBytes))
		} else if output != "" {
			writeOutput(output, outBytes, force)
		} else if fn == "-" {
			fmt.Print(string(outBytes))
		} else {
			// Default output: remove .json suffix, append ,v if not present
			outPath := fn
			if strings.HasSuffix(fn, ".json") {
				outPath = strings.TrimSuffix(fn, ".json")
			}
			if !strings.HasSuffix(outPath, ",v") {
				outPath += ",v"
			}
			writeOutput(outPath, outBytes, force)
		}
	}
}

func writeOutput(path string, data []byte, force bool) {
	if !force {
		if _, err := os.Stat(path); err == nil {
			log.Panicf("Output file %s already exists, use -f to force overwrite", path)
		}
	}
	if err := os.WriteFile(path, data, 0644); err != nil {
		log.Panicf("Error writing output to %s: %s", path, err)
	}
}

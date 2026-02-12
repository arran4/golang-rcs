package cli

import (
	"fmt"
	rcs "github.com/arran4/golang-rcs"
	"io"
	"log"
	"os"
)

// Format is a subcommand `gorcs format`
//
// Flags:
//
//	output: -o --output Output file path
//	force: -f --force Force overwrite output
//	overwrite: -w --overwrite Overwrite input file
//	stdout: -s --stdout Force output to stdout
//   ignore-truncation: --ignore-truncation Ignore year truncation
//	files: ... List of files to process, or - for stdin
func Format(output string, force, overwrite, stdout, ignoreTruncation bool, files ...string) {
	runFormat(output, force, overwrite, stdout, ignoreTruncation, files...)
}

func runFormat(output string, force, overwrite, stdout, ignoreTruncation bool, files ...string) {
	if output != "" && len(files) > 1 {
		log.Panicf("Cannot specify output file with multiple input files")
	}
	if overwrite && output != "" {
		log.Panicf("Cannot specify both overwrite and output file")
	}
	if overwrite && stdout {
		log.Panicf("Cannot specify both overwrite and stdout")
	}
	if output != "" && stdout {
		log.Panicf("Cannot specify both output and stdout")
	}

	targetStdout := stdout || (!overwrite && output == "")

	if targetStdout && len(files) > 1 {
		// Txtar format
		for _, fn := range files {
			r := parseFileOrStdin(fn)
			content := r.String()
			fmt.Printf("-- %s --\n", fn)
			fmt.Print(content)
		}
		return
	}

	for _, fn := range files {
		r := parseFileOrStdin(fn)
		if ignoreTruncation {
			r.DateYearPrefixTruncated = false
			for _, h := range r.RevisionHeads {
				h.YearTruncated = false
			}
		}
		content := r.String()

		if overwrite {
			if fn == "-" {
				log.Panicf("Cannot overwrite stdin")
			}
			if err := os.WriteFile(fn, []byte(content), 0644); err != nil {
				log.Panicf("Error writing file %s: %s", fn, err)
			}
		} else if output != "" {
			writeOutput(output, []byte(content), force)
		} else {
			// Stdout
			fmt.Print(content)
		}
	}
}

func parseFileOrStdin(fn string) *rcs.File {
	var f io.Reader
	var file *os.File
	var err error
	if fn == "-" {
		f = os.Stdin
	} else {
		file, err = os.Open(fn)
		if err != nil {
			log.Panicf("Error opening file %s: %s", fn, err)
		}
		f = file
	}
	r, err := rcs.ParseFile(f)
	if file != nil {
		file.Close()
	}
	if err != nil {
		log.Panicf("Error parsing %s: %s", fn, err)
	}
	return r
}

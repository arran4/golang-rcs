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
//			output: -o --output Output file path
//			force: -f --force Force overwrite output
//			overwrite: -w --overwrite Overwrite input file
//			stdout: -s --stdout Force output to stdout
//	    keep-truncated-years: --keep-truncated-years Keep truncated years (do not expand to 4 digits)
//			files: ... List of files to process, or - for stdin
func Format(stdin io.Reader, stdout io.Writer, output string, force, overwrite, stdoutFlag, keepTruncatedYears bool, files ...string) {
	runFormat(stdin, stdout, output, force, overwrite, stdoutFlag, keepTruncatedYears, files...)
}

func runFormat(stdin io.Reader, stdout io.Writer, output string, force, overwrite, stdoutFlag, keepTruncatedYears bool, files ...string) {
	if output != "" && len(files) > 1 {
		log.Panicf("Cannot specify output file with multiple input files")
	}
	if overwrite && output != "" {
		log.Panicf("Cannot specify both overwrite and output file")
	}
	if overwrite && stdoutFlag {
		log.Panicf("Cannot specify both overwrite and stdout")
	}
	if output != "" && stdoutFlag {
		log.Panicf("Cannot specify both output and stdout")
	}

	targetStdout := stdoutFlag || (!overwrite && output == "")

	if targetStdout && len(files) > 1 {
		// Txtar format
		for _, fn := range files {
			r := parseFileOrStdin(stdin, fn)
			content := r.String()
			fmt.Fprintf(stdout, "-- %s --\n", fn)
			fmt.Fprint(stdout, content)
		}
		return
	}

	for _, fn := range files {
		r := parseFileOrStdin(stdin, fn)
		if !keepTruncatedYears {
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
			fmt.Fprint(stdout, content)
		}
	}
}

func parseFileOrStdin(stdin io.Reader, fn string) *rcs.File {
	var f io.Reader
	var file *os.File
	var err error
	if fn == "-" {
		f = stdin
	} else {
		file, err = os.Open(fn)
		if err != nil {
			log.Panicf("Error opening file %s: %s", fn, err)
		}
		f = file
	}
	r, err := rcs.ParseFile(f)
	if file != nil {
		if err := file.Close(); err != nil {
			log.Panicf("Error closing file %s: %s", fn, err)
		}
	}
	if err != nil {
		log.Panicf("Error parsing %s: %s", fn, err)
	}
	return r
}

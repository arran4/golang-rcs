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
func Format(output string, force, overwrite, stdoutFlag, keepTruncatedYears bool, files ...string) {
	runFormat(os.Stdin, os.Stdout, output, force, overwrite, stdoutFlag, keepTruncatedYears, files...)
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

	for _, fn := range files {
		var content string
		var err error

		if fn == "-" {
			content, err = processReader(stdin, keepTruncatedYears)
			if err != nil {
				log.Panicf("Error parsing stdin: %s", err)
			}
		} else {
			// Using closure to ensure Close is called immediately after use
			func() {
				f, openErr := os.Open(fn)
				if openErr != nil {
					log.Panicf("Error opening file %s: %s", fn, openErr)
				}
				defer func() {
					if closeErr := f.Close(); closeErr != nil {
						log.Panicf("Error closing file %s: %s", fn, closeErr)
					}
				}()

				content, err = processReader(f, keepTruncatedYears)
			}()
			if err != nil {
				log.Panicf("Error parsing file %s: %s", fn, err)
			}
		}

		if overwrite {
			if fn == "-" {
				log.Panicf("Cannot overwrite stdin")
			}
			if err := os.WriteFile(fn, []byte(content), 0644); err != nil {
				log.Panicf("Error writing file %s: %s", fn, err)
			}
		} else if output != "" && output != "-" {
			writeOutput(output, []byte(content), force)
		} else {
			// Stdout
			_, _ = fmt.Fprint(stdout, content)
		}
	}
}

func processReader(r io.Reader, keepTruncatedYears bool) (string, error) {
	parsedFile, err := rcs.ParseFile(r)
	if err != nil {
		return "", err
	}
	if !keepTruncatedYears {
		parsedFile.DateYearPrefixTruncated = false
		for _, h := range parsedFile.RevisionHeads {
			if h.YearTruncated {
				if t, err := h.Date.DateTime(); err == nil {
					h.Date = rcs.DateTime(t.Format(rcs.DateFormat))
					h.YearTruncated = false
				}
			}
		}
	}
	return parsedFile.String(), nil
}

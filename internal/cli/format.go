package cli

import (
	"fmt"
	rcs "github.com/arran4/golang-rcs"
	"io"
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
func Format(output string, force, overwrite, stdoutFlag, keepTruncatedYears bool, files ...string) error {
	return runFormat(os.Stdin, os.Stdout, output, force, overwrite, stdoutFlag, keepTruncatedYears, files...)
}

func runFormat(stdin io.Reader, stdout io.Writer, output string, force, overwrite, stdoutFlag, keepTruncatedYears bool, files ...string) error {
	if output != "" && len(files) > 1 {
		return fmt.Errorf("cannot specify output file with multiple input files")
	}
	if overwrite && output != "" {
		return fmt.Errorf("cannot specify both overwrite and output file")
	}
	if overwrite && stdoutFlag {
		return fmt.Errorf("cannot specify both overwrite and stdout")
	}
	if output != "" && stdoutFlag {
		return fmt.Errorf("cannot specify both output and stdout")
	}

	for _, fn := range files {
		var content string
		var err error

		if fn == "-" {
			content, err = processReader(stdin, keepTruncatedYears)
			if err != nil {
				return fmt.Errorf("error parsing stdin: %w", err)
			}
		} else {
			// Using closure to ensure Close is called immediately after use
			err = func() error {
				f, openErr := os.Open(fn)
				if openErr != nil {
					return fmt.Errorf("error opening file %s: %w", fn, openErr)
				}
				defer func() {
					_ = f.Close()
				}()

				content, err = processReader(f, keepTruncatedYears)
				return err
			}()
			if err != nil {
				return fmt.Errorf("error parsing file %s: %w", fn, err)
			}
		}

		if overwrite {
			if fn == "-" {
				return fmt.Errorf("cannot overwrite stdin")
			}
			if err := os.WriteFile(fn, []byte(content), 0644); err != nil {
				return fmt.Errorf("error writing file %s: %w", fn, err)
			}
		} else if output != "" {
			if err := writeOutput(output, []byte(content), force); err != nil {
				return err
			}
		} else {
			// Stdout
			_, _ = fmt.Fprint(stdout, content)
		}
	}
	return nil
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

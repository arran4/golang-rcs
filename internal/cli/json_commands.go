package cli

import (
	"encoding/json"
	"fmt"
	rcs "github.com/arran4/golang-rcs"
	"io"
	"os"
	"strings"
)

// ToJson is a subcommand `gorcs to-json`
//
// Flags:
//
//	output: -o --output Output file path
//	force: -f --force Force overwrite output
//	indent: -I --indent Indent JSON output
//	files: ... List of files to process, or - for stdin
func ToJson(output string, force, indent bool, files ...string) error {
	if output != "" && output != "-" && len(files) > 1 {
		return fmt.Errorf("cannot specify output file with multiple input files")
	}
	for _, fn := range files {
		var f io.Reader
		if fn == "-" {
			f = os.Stdin
		} else {
			file, err := os.Open(fn)
			if err != nil {
				return fmt.Errorf("error with file %s: %w", fn, err)
			}
			defer func() {
				_ = file.Close()
			}()
			f = file
		}
		r, err := rcs.ParseFile(f)
		if err != nil {
			return fmt.Errorf("error parsing %s: %w", fn, err)
		}
		var b []byte
		if indent {
			b, err = json.MarshalIndent(r, "", "  ")
		} else {
			b, err = json.Marshal(r)
		}
		if err != nil {
			return fmt.Errorf("error serializing %s: %w", fn, err)
		}

		if output == "-" {
			fmt.Printf("%s", b)
		} else if output != "" {
			if err := writeOutput(output, b, force); err != nil {
				return err
			}
		} else if fn == "-" {
			fmt.Printf("%s", b)
		} else {
			// Default output: filename + .json
			outPath := fn + ".json"
			if err := writeOutput(outPath, b, force); err != nil {
				return err
			}
		}
	}
	return nil
}

// FromJson is a subcommand `gorcs from-json`
//
// Flags:
//
//	output: -o --output Output file path
//	force: -f --force Force overwrite output
//	files: ... List of files to process, or - for stdin
func FromJson(output string, force bool, files ...string) error {
	if output != "" && output != "-" && len(files) > 1 {
		return fmt.Errorf("cannot specify output file with multiple input files")
	}
	for _, fn := range files {
		var f io.Reader
		if fn == "-" {
			f = os.Stdin
		} else {
			file, err := os.Open(fn)
			if err != nil {
				return fmt.Errorf("error with file %s: %w", fn, err)
			}
			defer func() {
				_ = file.Close()
			}()
			f = file
		}
		var r rcs.File
		if err := json.NewDecoder(f).Decode(&r); err != nil {
			return fmt.Errorf("error parsing %s: %w", fn, err)
		}

		outBytes := []byte(r.String())

		if output == "-" {
			fmt.Print(string(outBytes))
		} else if output != "" {
			if err := writeOutput(output, outBytes, force); err != nil {
				return err
			}
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
			if err := writeOutput(outPath, outBytes, force); err != nil {
				return err
			}
		}
	}
	return nil
}

func writeOutput(path string, data []byte, force bool) error {
	if !force {
		if _, err := os.Stat(path); err == nil {
			return fmt.Errorf("output file %s already exists, use -f to force overwrite", path)
		}
	}
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("error writing output to %s: %w", path, err)
	}
	fmt.Printf("Wrote: %s\n", path)
	return nil
}

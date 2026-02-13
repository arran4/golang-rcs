package cli

import "os"

// Validate is a subcommand `gorcs validate`
//
// Flags:
//
//	output: -o --output Output file path
//	force: -f --force Force overwrite output
//	overwrite: -w --overwrite Overwrite input file
//	files: ... List of files to process, or - for stdin
func Validate(output string, force, overwrite bool, files ...string) {
	// Validate is currently functionally identical to Format (parse and re-serialize).
	// If validation rules diverge in future, logic can be separated here.
	runFormat(os.Stdin, os.Stdout, output, force, overwrite, false, files...)
}

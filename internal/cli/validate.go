package cli

import "os"

// Validate is a subcommand `gorcs validate`
//
// Flags:
//
//	output: -o --output Output file path
//	force: -f --force Force overwrite output
//	files: ... List of files to process, or - for stdin
func Validate(output string, force bool, files ...string) error {
	var err error
	if files, err = ensureFiles(files); err != nil {
		return err
	}
	// Validate is currently functionally identical to Format (parse and re-serialize).
	// If validation rules diverge in future, logic can be separated here.
	return runFormat(os.Stdin, os.Stdout, output, force, false, false, false, files...)
}

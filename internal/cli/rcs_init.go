package cli

import (
	"fmt"
	"os"
	"strings"

	rcs "github.com/arran4/golang-rcs"
)

// Init is a subcommand `gorcs init`
//
// Flags:
//
//	description: -t description of the file
//	files: ... List of working files to process
func Init(description string, files ...string) error {
	if len(files) == 0 {
		return fmt.Errorf("no files provided")
	}
	for _, file := range files {
		if err := initFile(description, file); err != nil {
			return err
		}
	}
	return nil
}

func initFile(description, workingFile string) error {
	rcsFile := workingFile
	if !strings.HasSuffix(rcsFile, ",v") {
		rcsFile += ",v"
	}

	if _, err := os.Stat(rcsFile); err == nil {
		return fmt.Errorf("file %s already exists", rcsFile)
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("stat %s: %w", rcsFile, err)
	}

	f := rcs.NewFile()
	if description != "" {
		f.Description = description
	}

	if err := os.WriteFile(rcsFile, []byte(f.String()), 0600); err != nil {
		return fmt.Errorf("write %s: %w", rcsFile, err)
	}

	return nil
}

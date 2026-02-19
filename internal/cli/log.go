package cli

import (
	"fmt"
	"os"
	"strings"

	rcs "github.com/arran4/golang-rcs"
)

// LogMessageChange is a subcommand `gorcs log message change`
//
// Flags:
//
//	revision: -rev revision to change
//	message: -m new log message
//	files: ... List of working files to process
func LogMessageChange(revision, message string, files ...string) error {
	for _, file := range files {
		rcsFile := file
		if !strings.HasSuffix(rcsFile, ",v") {
			rcsFile += ",v"
		}

		f, err := os.Open(rcsFile)
		if err != nil {
			return fmt.Errorf("open %s: %w", rcsFile, err)
		}

		parsedFile, err := rcs.ParseFile(f)
		if err := f.Close(); err != nil {
			return fmt.Errorf("close %s: %w", rcsFile, err)
		}
		if err != nil {
			return fmt.Errorf("parse %s: %w", rcsFile, err)
		}

		if err := parsedFile.ChangeLogMessage(revision, message); err != nil {
			return fmt.Errorf("change log message in %s: %w", rcsFile, err)
		}

		// Write back the file
		if err := os.WriteFile(rcsFile, []byte(parsedFile.String()), 0644); err != nil {
			return fmt.Errorf("write %s: %w", rcsFile, err)
		}
	}
	return nil
}

// LogMessagePrint is a subcommand `gorcs log message print`
//
// Flags:
//
//	revision: -rev revision to print
//	files: ... List of working files to process
func LogMessagePrint(revision string, files ...string) error {
	for _, file := range files {
		rcsFile := file
		if !strings.HasSuffix(rcsFile, ",v") {
			rcsFile += ",v"
		}

		f, err := os.Open(rcsFile)
		if err != nil {
			return fmt.Errorf("open %s: %w", rcsFile, err)
		}

		parsedFile, err := rcs.ParseFile(f)
		if err := f.Close(); err != nil {
			return fmt.Errorf("close %s: %w", rcsFile, err)
		}
		if err != nil {
			return fmt.Errorf("parse %s: %w", rcsFile, err)
		}

		msg, err := parsedFile.GetLogMessage(revision)
		if err != nil {
			return fmt.Errorf("get log message in %s: %w", rcsFile, err)
		}

		fmt.Printf("File: %s Revision: %s\n%s\n", file, revision, msg)
	}
	return nil
}

// LogMessageList is a subcommand `gorcs log message list`
//
// Flags:
//
//	files: ... List of working files to process
func LogMessageList(files ...string) error {
	for _, file := range files {
		rcsFile := file
		if !strings.HasSuffix(rcsFile, ",v") {
			rcsFile += ",v"
		}

		f, err := os.Open(rcsFile)
		if err != nil {
			return fmt.Errorf("open %s: %w", rcsFile, err)
		}

		parsedFile, err := rcs.ParseFile(f)
		if err := f.Close(); err != nil {
			return fmt.Errorf("close %s: %w", rcsFile, err)
		}
		if err != nil {
			return fmt.Errorf("parse %s: %w", rcsFile, err)
		}

		fmt.Printf("File: %s\n", file)
		logs := parsedFile.ListLogMessages()
		for _, log := range logs {
			fmt.Printf("Revision: %s\n%s\n", log.Revision, log.Log)
		}
		fmt.Println()
	}
	return nil
}

// Log is a subcommand `gorcs log`
//
// Flags:
//
//	filterStr: -F or --filter filter string
//	stateFilter: -s state filters
//	files: ... List of working files to process
func Log(filterStr string, stateFilter string, files ...string) error {
	if len(files) > 0 && files[0] == "filter-reference" {
		printFilterReference()
		return nil
	}

	var filters []rcs.Filter

	if filterStr != "" {
		f, err := rcs.ParseFilter(filterStr)
		if err != nil {
			return fmt.Errorf("parse filter %q: %w", filterStr, err)
		}
		filters = append(filters, f)
	}

	var allowedStates []string
	if stateFilter != "" {
		states := strings.Split(stateFilter, ",")
		for _, state := range states {
			state = strings.TrimSpace(state)
			if state != "" {
				allowedStates = append(allowedStates, state)
			}
		}
	}

	if len(allowedStates) > 0 {
		var sFilters []rcs.Filter
		for _, state := range allowedStates {
			sFilters = append(sFilters, &rcs.StateFilter{State: state})
		}
		filters = append(filters, &rcs.OrFilter{Filters: sFilters})
	}

	var combinedFilter rcs.Filter
	if len(filters) > 0 {
		combinedFilter = &rcs.AndFilter{Filters: filters}
	}

	for _, file := range files {
		rcsFile := file
		if !strings.HasSuffix(rcsFile, ",v") {
			rcsFile += ",v"
		}

		f, err := os.Open(rcsFile)
		if err != nil {
			return fmt.Errorf("open %s: %w", rcsFile, err)
		}

		parsedFile, err := rcs.ParseFile(f)
		if err := f.Close(); err != nil {
			return fmt.Errorf("close %s: %w", rcsFile, err)
		}
		if err != nil {
			return fmt.Errorf("parse %s: %w", rcsFile, err)
		}

		if err := rcs.PrintRLog(os.Stdout, parsedFile, rcsFile, file, combinedFilter); err != nil {
			return fmt.Errorf("print rlog %s: %w", rcsFile, err)
		}
	}
	return nil
}

func printFilterReference() {
	fmt.Println("Filter Reference:")
	fmt.Println()
	fmt.Println("Filtering allows selecting specific revisions based on criteria.")
	fmt.Println("Syntax:")
	fmt.Println("  state=<value>       Select revisions with the given state.")
	fmt.Println("  s=<value>           Alias for state.")
	fmt.Println("  state in (val ...)  Select revisions where state matches one of the values.")
	fmt.Println("  <expr> OR <expr>    Logical OR.")
	fmt.Println("  <expr> || <expr>    Logical OR.")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  state=Rel")
	fmt.Println("  s=Exp || s=Prod")
	fmt.Println("  state in (Rel Prod)")
}

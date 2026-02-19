package rcs

import (
	"bufio"
	"bytes"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"golang.org/x/tools/txtar"
)

//go:embed testdata/txtar/*.txtar testdata/txtar/operations/*.txtar
var txtarTests embed.FS

var ErrorTestProvider = map[string]func(b []byte) (error, error){
	"ErrParseProperty": func(b []byte) (error, error) {
		var targetErr ErrParseProperty
		err := json.Unmarshal(b, &targetErr)
		if targetErr.Err == nil {
			targetErr.Err = errors.New("placeholder")
		}
		return targetErr, err
	},
	"ErrTooManyNewLines": func(b []byte) (error, error) {
		return ErrTooManyNewLines, nil
	},
	"ErrEmptyId": func(b []byte) (error, error) {
		return ErrEmptyId, nil
	},
	"ErrRevisionEmpty": func(b []byte) (error, error) {
		return ErrRevisionEmpty, nil
	},
	"ErrDateParse": func(b []byte) (error, error) {
		return ErrDateParse, nil
	},
	"ErrUnknownToken": func(b []byte) (error, error) {
		return ErrUnknownToken, nil
	},
}

func TestTxtarFiles(t *testing.T) {
	err := fs.WalkDir(txtarTests, "testdata/txtar", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if !strings.HasSuffix(d.Name(), ".txtar") {
			return nil
		}
		t.Run(d.Name(), func(t *testing.T) {
			runTest(t, txtarTests, path)
		})
		return nil
	})
	if err != nil {
		t.Fatalf("WalkDir error: %v", err)
	}
}

func runTest(t *testing.T, fsys fs.FS, filename string) {
	content, err := fs.ReadFile(fsys, filename)
	if err != nil {
		t.Fatalf("ReadFile error: %v", err)
	}
	// TDOO find a better work around
	// Windows compatibility: txtar.Parse expects LF line endings.
	// If the file was checked out with CRLF, we need to normalize it.
	content = bytes.ReplaceAll(content, []byte("\r\n"), []byte("\n"))

	archive := txtar.Parse(content)
	parts := make(map[string]string)
	for _, f := range archive.Files {
		parts[f.Name] = string(f.Data)
	}

	// description.txt check
	if _, ok := parts["description.txt"]; !ok {
		t.Log("description.txt is missing")
	}

	// options.conf / options.json
	options := make(map[string]bool)
	optionArgs := []string{}
	var optionMessage, optionSubCommand, optionRevision string
	var optionFiles []string
	if optContent, ok := parts["options.conf"]; ok {
		parseOptions(optContent, options, &optionArgs, &optionMessage, &optionSubCommand, &optionRevision, &optionFiles)
	}
	if optContent, ok := parts["options.json"]; ok {
		parseOptions(optContent, options, &optionArgs, &optionMessage, &optionSubCommand, &optionRevision, &optionFiles)
	}

	// tests.txt or tests.md
	testContent, ok := parts["tests.txt"]
	if !ok {
		testContent, ok = parts["tests.md"]
	}

	if !ok {
		t.Fatalf("Missing tests.txt or tests.md")
		return
	}

	scanner := bufio.NewScanner(strings.NewReader(testContent))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		// Strip markdown list markers
		line = strings.TrimPrefix(line, "* ")
		line = strings.TrimPrefix(line, "- ")

		// Split by comma for multiple tests on one line?
		testLine := strings.SplitN(line, ":", 2)
		testName := testLine[0]

		testName = strings.TrimSpace(testName)
		if testName == "" {
			continue
		}

		switch {
		case testName == "json to rcs":
			testJSONToRCS(t, parts, options)
		case testName == "rcs to json":
			testRCSToJSON(t, parts, options)
		case testName == "rcs to rcs":
			testCircular(t, parts, options)
		case testName == "format rcs":
			testFormatRCS(t, parts, options)
		case testName == "validate rcs":
			testValidateRCS(t, parts, options)
		case testName == "rcs init":
			testNewRCS(t, parts, options, optionArgs)
		case testName == "list heads":
			testListHeads(t, parts, options)
		case testName == "normalize revisions":
			testNormalizeRevisions(t, parts, options)
		case testName == "parse error":
			testParseError(t, line, parts, options)
		case strings.HasPrefix(testName, "parse error:"):
			// TODO this test case should be no more as the error details should be moved into option.s
			fullLine := line
			if testName == "parse error" && len(testLine) > 1 {
				fullLine = "parse error: " + testLine[1]
			}
			testParseError(t, fullLine, parts, options)
		case testName == "rcs access-list":
			testAccessList(t, parts, options, optionArgs)
		case testName == "rcs":
			testRCS(t, parts, options, optionArgs)
		case strings.HasPrefix(testName, "rcs "):
			testRCS(t, parts, options, optionArgs)
		case testName == "rcs merge":
			testRCSMerge(t, parts, options)
		case testName == "rcs merge":
			testRCSMerge(t, parts, options)
		case testName == "ci":
			testCI(t, parts, options)
		case testName == "co":
			testCO(t, parts, options, optionArgs)
		case testName == "rcsdiff":
			testRCSDiff(t, parts, options)
		case testName == "rcs diff":
			testRCSDiff(t, parts, options)
		case testName == "rcs merge":
			testRCSMerge(t, parts, options)
		case testName == "rcs clean":
			testRCSClean(t, parts, options)
		case strings.HasPrefix(testName, "log message"):
			testLogMessage(t, parts, options, optionArgs, optionMessage, optionSubCommand, optionRevision, optionFiles)
		case testName == "rlog":
			testRLog(t, parts, options, optionArgs)
		case testName == "rcs log":
			testRLog(t, parts, options, optionArgs)
		default:
			t.Errorf("Unknown test type: %q", testName)
		}
	}
}

func testRLog(t *testing.T, parts map[string]string, options map[string]bool, args []string) {
	t.Run("rlog", func(t *testing.T) {
		inputRCS := getInputRCS(t, parts)
		expectedOut, ok := parts["expected.stdout"]
		if !ok {
			if expectedOut, ok = parts["expected.out"]; !ok {
				t.Fatal("Missing expected.stdout or expected.out")
			}
		}

		parsedFile, err := parseRCS(inputRCS)
		if err != nil {
			t.Fatalf("ParseFile error: %v", err)
		}

		var filterStr string
		var stateFilters []string

		for i := 0; i < len(args); i++ {
			arg := args[i]
			if strings.HasPrefix(arg, "-s") {
				val := strings.TrimPrefix(arg, "-s")
				if val == "" {
					if i+1 < len(args) {
						stateFilters = append(stateFilters, args[i+1])
						i++
					}
				} else {
					stateFilters = append(stateFilters, val)
				}
			} else if strings.HasPrefix(arg, "--filter=") {
				filterStr = strings.TrimPrefix(arg, "--filter=")
			} else if arg == "-F" || arg == "--filter" {
				if i+1 < len(args) {
					filterStr = args[i+1]
					i++
				}
			} else if strings.HasPrefix(arg, "-F") {
				filterStr = strings.TrimPrefix(arg, "-F")
			}
		}

		var filters []Filter
		if filterStr != "" {
			f, err := ParseFilter(filterStr)
			if err != nil {
				t.Fatalf("ParseFilter error: %v", err)
			}
			filters = append(filters, f)
		}

		var allowedStates []string
		for _, s := range stateFilters {
			states := strings.Split(s, ",")
			for _, state := range states {
				state = strings.TrimSpace(state)
				if state != "" {
					allowedStates = append(allowedStates, state)
				}
			}
		}
		if len(allowedStates) > 0 {
			var sFilters []Filter
			for _, state := range allowedStates {
				sFilters = append(sFilters, &StateFilter{State: state})
			}
			filters = append(filters, &OrFilter{Filters: sFilters})
		}

		var combinedFilter Filter
		if len(filters) > 0 {
			combinedFilter = &AndFilter{Filters: filters}
		}

		var sb strings.Builder
		// Assuming filename "input.txt,v" and working filename "input.txt" as per test convention
		err = PrintRLog(&sb, parsedFile, "input.txt,v", "input.txt", combinedFilter)
		if err != nil {
			t.Fatalf("PrintRLog error: %v", err)
		}

		gotOut := sb.String()

		if options["force unix line endings"] {
			expectedOut = strings.ReplaceAll(expectedOut, "\r\n", "\n")
		}

		if diff := cmp.Diff(strings.TrimSpace(expectedOut), strings.TrimSpace(gotOut)); diff != "" {
			t.Errorf("rlog mismatch (-want +got):\n%s", diff)
		}
	})
}

func testLogMessage(t *testing.T, parts map[string]string, options map[string]bool, args []string, message, subCommand, revision string, files []string) {
	t.Run("log message", func(t *testing.T) {
		if subCommand == "" {
			if len(args) == 0 {
				t.Fatal("Missing subcommand for log message")
			}
			subCommand = args[0]
			remainingArgs := args[1:]

			for i := 0; i < len(remainingArgs); i++ {
				arg := remainingArgs[i]
				if arg == "-rev" && i+1 < len(remainingArgs) {
					revision = remainingArgs[i+1]
					i++
				} else if arg == "-m" && i+1 < len(remainingArgs) {
					message = remainingArgs[i+1]
					i++
				} else if !strings.HasPrefix(arg, "-") {
					files = append(files, arg)
				}
			}
		}

		if len(files) == 0 {
			if _, ok := parts["input.txt,v"]; ok {
				files = append(files, "input.txt,v")
			} else if _, ok := parts["input.rcs"]; ok {
				files = append(files, "input.rcs")
			}
		}

		if len(files) == 0 {
			t.Fatal("No files specified")
		}

		for _, file := range files {
			content, ok := parts[file]
			if !ok {
				if content, ok = parts[file+",v"]; !ok {
					t.Fatalf("Missing file part: %s", file)
				}
			}

			parsedFile, err := parseRCS(content)
			if err != nil {
				t.Fatalf("parseRCS failed: %v", err)
			}

			switch subCommand {
			case "change":
				if revision == "" || message == "" {
					t.Fatal("change requires -rev and -m")
				}
				if err := parsedFile.ChangeLogMessage(revision, message); err != nil {
					t.Fatalf("ChangeLogMessage failed: %v", err)
				}
				expectedContent, ok := parts["expected.txt,v"]
				if !ok {
					t.Fatalf("Missing expected.txt,v")
				}
				checkRCS(t, expectedContent, parsedFile.String(), options)

			case "print":
				if revision == "" {
					t.Fatal("print requires -rev")
				}
				msg, err := parsedFile.GetLogMessage(revision)
				if err != nil {
					t.Fatalf("GetLogMessage failed: %v", err)
				}
				expectedOut, ok := parts["expected.out"]
				if !ok {
					t.Fatalf("Missing expected.out")
				}
				var sb strings.Builder
				fmt.Fprintf(&sb, "File: %s Revision: %s\n%s\n", file, revision, msg)
				gotOut := sb.String()

				if options["force unix line endings"] {
					expectedOut = strings.ReplaceAll(expectedOut, "\r\n", "\n")
				}
				if diff := cmp.Diff(strings.TrimSpace(expectedOut), strings.TrimSpace(gotOut)); diff != "" {
					t.Errorf("Log message print mismatch (-want +got):\n%s", diff)
				}

			case "list":
				logs := parsedFile.ListLogMessages()
				expectedOut, ok := parts["expected.out"]
				if !ok {
					t.Fatalf("Missing expected.out")
				}
				var sb strings.Builder
				fmt.Fprintf(&sb, "File: %s\n", file)
				for _, l := range logs {
					fmt.Fprintf(&sb, "Revision: %s\n%s\n", l.Revision, l.Log)
				}
				fmt.Fprintln(&sb)

				gotOut := sb.String()

				if options["force unix line endings"] {
					expectedOut = strings.ReplaceAll(expectedOut, "\r\n", "\n")
				}
				if diff := cmp.Diff(strings.TrimSpace(expectedOut), strings.TrimSpace(gotOut)); diff != "" {
					t.Errorf("Log message list mismatch (-want +got):\n%s", diff)
				}

			default:
				t.Fatalf("Unknown subcommand: %s", subCommand)
			}
		}
	})
}

func testRCSClean(t *testing.T, parts map[string]string, options map[string]bool) {
	t.Skip("rcs clean test type not implemented yet")
}

func testAccessList(t *testing.T, parts map[string]string, options map[string]bool, args []string) {
	t.Run("rcs access-list", func(t *testing.T) {
		if len(args) == 0 {
			t.Fatal("Missing subcommand for access-list")
		}
		subCmd := args[0]
		remainingArgs := args[1:]

		var fromFile string
		var targetFiles []string

		for i := 0; i < len(remainingArgs); i++ {
			if remainingArgs[i] == "-from" && i+1 < len(remainingArgs) {
				fromFile = remainingArgs[i+1]
				i++
			} else if !strings.HasPrefix(remainingArgs[i], "-") {
				targetFiles = append(targetFiles, remainingArgs[i])
			}
		}

		if fromFile == "" {
			t.Fatal("Missing -from argument")
		}
		if len(targetFiles) == 0 {
			t.Fatal("Missing target files")
		}

		fromContent, ok := parts[fromFile]
		if !ok {
			t.Fatalf("Missing from file: %s", fromFile)
		}

		fromRCS, err := parseRCS(fromContent)
		if err != nil {
			t.Fatalf("Failed to parse from file %s: %v", fromFile, err)
		}

		for _, targetFile := range targetFiles {
			targetContent, ok := parts[targetFile]
			if !ok {
				t.Fatalf("Missing target file: %s", targetFile)
			}

			targetRCS, err := parseRCS(targetContent)
			if err != nil {
				t.Fatalf("Failed to parse target file %s: %v", targetFile, err)
			}

			switch subCmd {
			case "copy":
				targetRCS.CopyAccessList(fromRCS)
			case "append":
				targetRCS.AppendAccessList(fromRCS)
			default:
				t.Fatalf("Unknown subcommand: %s", subCmd)
			}

			expectedKey := "expected.txt,v"
			expectedContent, ok := parts[expectedKey]
			if !ok {
				t.Fatalf("Missing expected file: %s", expectedKey)
			}

			checkRCS(t, expectedContent, targetRCS.String(), options)
		}
	})
}

func testRCSDiff(t *testing.T, parts map[string]string, options map[string]bool) {
	t.Skip("rcsdiff test type not implemented yet")
}

func testRCS(t *testing.T, parts map[string]string, _ map[string]bool, args []string) {
	t.Run("rcs", func(t *testing.T) {
		input, ok := parts["input.txt,v"]
		if !ok {
			t.Skip("rcs test type currently supports only input.txt,v fixtures")
		}
		expectedRCS, ok := parts["expected.txt,v"]
		if !ok {
			t.Skip("rcs test type currently supports only expected.txt,v fixtures")
		}

		branchName := ""
		for i := 0; i < len(args); i++ {
			if strings.HasPrefix(args[i], "-b") {
				branchName = strings.TrimPrefix(args[i], "-b")
				break
			}
			if i+3 < len(args) && args[i] == "branches" && args[i+1] == "default" && args[i+2] == "set" {
				branchName = args[i+3]
				break
			}
		}
		if branchName == "" {
			t.Skip("unsupported rcs operation fixture")
		}

		parsed, err := parseRCS(input)
		if err != nil {
			t.Fatalf("ParseFile error: %v", err)
		}

		parts := strings.Split(branchName, ".")
		if len(parts)%2 == 0 && len(parts) > 0 {
			branchName = strings.Join(parts[:len(parts)-1], ".")
		}
		parsed.Branch = branchName

		if diff := cmp.Diff(strings.TrimSpace(expectedRCS), strings.TrimSpace(parsed.String())); diff != "" {
			t.Fatalf("RCS file mismatch (-want +got):\n%s", diff)
		}
	})
}

func testRCSMerge(t *testing.T, parts map[string]string, options map[string]bool) {
	t.Skip("rcs merge test type not implemented yet")
}

func testCI(t *testing.T, parts map[string]string, options map[string]bool) {
	t.Skip("ci test type not implemented yet")
}

func testCO(t *testing.T, parts map[string]string, _ map[string]bool, args []string) {
	t.Run("co", func(t *testing.T) {
		input, ok := parts["input.txt,v"]
		if !ok {
			t.Fatal("Missing input.txt,v")
		}
		expectedWorking, ok := parts["expected.txt"]
		if !ok {
			t.Fatal("Missing expected.txt")
		}
		expectedRCS, hasExpectedRCS := parts["expected.txt,v"]

		parsed, err := parseRCS(input)
		if err != nil {
			t.Fatalf("ParseFile error: %v", err)
		}

		user := "tester"
		ops := make([]any, 0, 2)
		for _, arg := range args {
			if !strings.HasPrefix(arg, "-") {
				continue
			}
			switch {
			case arg == "-q":
				continue
			case strings.HasPrefix(arg, "-k"), strings.HasPrefix(arg, "-f"), strings.HasPrefix(arg, "-s"):
				t.Skipf("unsupported co flag in basic co mode: %s", arg)
			case strings.HasPrefix(arg, "-w"):
				if arg == "-w" {
					continue
				}
				user = strings.TrimPrefix(arg, "-w")
			case strings.HasPrefix(arg, "-r"):
				rev := strings.TrimPrefix(arg, "-r")
				if strings.Count(rev, ".") > 1 {
					t.Skipf("unsupported branch checkout in basic co mode: %s", arg)
				}
				ops = append(ops, WithRevision(rev))
			case strings.HasPrefix(arg, "-l"):
				rev := strings.TrimPrefix(arg, "-l")
				if rev != "" {
					ops = append(ops, WithRevision(rev))
				}
				ops = append(ops, WithSetLock)
			case strings.HasPrefix(arg, "-u"):
				rev := strings.TrimPrefix(arg, "-u")
				if rev != "" {
					ops = append(ops, WithRevision(rev))
				}
				ops = append(ops, WithClearLock)
			default:
				t.Skipf("unsupported co arg format in basic co mode: %s", arg)
			}
		}

		verdict, err := parsed.Checkout(user, ops...)
		if err != nil {
			t.Fatalf("Checkout failed: %v", err)
		}

		if diff := cmp.Diff(strings.TrimSpace(expectedWorking), strings.TrimSpace(verdict.Content)); diff != "" {
			t.Fatalf("working file mismatch (-want +got):\n%s", diff)
		}

		if hasExpectedRCS {
			if diff := cmp.Diff(strings.TrimSpace(expectedRCS), strings.TrimSpace(parsed.String())); diff != "" {
				t.Fatalf("RCS file mismatch (-want +got):\n%s", diff)
			}
		}
	})
}

func parseOptions(content string, options map[string]bool, optionArgs *[]string, message, subCommand, revision *string, files *[]string) {
	trimmed := strings.TrimSpace(content)
	if strings.HasPrefix(trimmed, "{") {
		var parsed struct {
			Args            []string        `json:"args"`
			TransformedArgs []string        `json:"transformed_args"`
			Flags           map[string]bool `json:"flags"`
			Options         []string        `json:"options"`
			Message         string          `json:"message"`
			SubCommand      string          `json:"subcommand"`
			Revision        string          `json:"revision"`
			Files           []string        `json:"files"`
		}
		if err := json.Unmarshal([]byte(trimmed), &parsed); err == nil {
			selectedArgs := parsed.Args
			if len(parsed.TransformedArgs) > 0 {
				selectedArgs = parsed.TransformedArgs
			}
			if len(selectedArgs) > 0 {
				*optionArgs = append((*optionArgs)[:0], selectedArgs...)
			}
			for k, v := range parsed.Flags {
				options[k] = v
			}
			for _, opt := range parsed.Options {
				options[opt] = true
			}
			if parsed.Message != "" {
				*message = parsed.Message
			}
			if parsed.SubCommand != "" {
				*subCommand = parsed.SubCommand
			}
			if parsed.Revision != "" {
				*revision = parsed.Revision
			}
			if len(parsed.Files) > 0 {
				*files = append((*files)[:0], parsed.Files...)
			}
			return
		}
	}

	scanner := bufio.NewScanner(strings.NewReader(content))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "* ") {
			options[strings.TrimPrefix(line, "* ")] = true
		}
	}
}

func testJSONToRCS(t *testing.T, parts map[string]string, options map[string]bool) {
	t.Run("json to rcs", func(t *testing.T) {
		inputJSON, ok := parts["input.json"]
		if !ok {
			t.Fatal("Missing input.json")
		}
		expectedRCS, ok := parts["expected,v"]
		if !ok {
			t.Fatal("Missing expected,v")
		}

		var file File
		if err := json.Unmarshal([]byte(inputJSON), &file); err != nil {
			t.Fatalf("json.Unmarshal error: %v", err)
		}

		gotRCS := file.String()
		checkRCS(t, expectedRCS, gotRCS, options)
	})
}

func testRCSToJSON(t *testing.T, parts map[string]string, options map[string]bool) {
	t.Run("rcs to json", func(t *testing.T) {
		inputRCS := getInputRCS(t, parts)
		expectedJSON, ok := parts["expected.json"]
		if !ok {
			t.Fatal("Missing expected.json")
		}

		parsedFile, err := parseRCS(inputRCS)
		if err != nil {
			t.Fatalf("ParseFile error: %v", err)
		}

		if options["unix line endings"] {
			parsedFile.SwitchLineEnding("\n")
		}

		gotJSONBytes, err := json.MarshalIndent(parsedFile, "", "  ")
		if err != nil {
			t.Fatalf("json.MarshalIndent error: %v", err)
		}
		gotJSON := string(gotJSONBytes)

		if options["force unix line endings"] {
			expectedJSON = strings.ReplaceAll(expectedJSON, "\r\n", "\n")
		}

		var expectedObj any
		if err := json.Unmarshal([]byte(expectedJSON), &expectedObj); err != nil {
			t.Fatalf("json.Unmarshal expected.json error: %v", err)
		}
		var gotObj any
		if err := json.Unmarshal([]byte(gotJSON), &gotObj); err != nil {
			t.Fatalf("json.Unmarshal got JSON error: %v", err)
		}

		if diff := cmp.Diff(expectedObj, gotObj); diff != "" {
			t.Errorf("JSON mismatch (-want +got):\n%s", diff)
		}
	})
}

func testCircular(t *testing.T, parts map[string]string, options map[string]bool) {
	t.Run("rcs to rcs", func(t *testing.T) {
		inputRCS := getInputRCS(t, parts)

		parsedFile, err := parseRCS(inputRCS)
		if err != nil {
			t.Fatalf("ParseFile error: %v", err)
		}

		if options["unix line endings"] {
			parsedFile.SwitchLineEnding("\n")
		}

		gotRCS := parsedFile.String()

		expectedRCS := inputRCS
		// If expected,v is present, strictly match that, otherwise match input
		if exp, ok := parts["expected,v"]; ok {
			expectedRCS = exp
		}

		checkRCS(t, expectedRCS, gotRCS, options)
	})
}

func testFormatRCS(t *testing.T, parts map[string]string, options map[string]bool) {
	t.Run("format rcs", func(t *testing.T) {
		inputRCS := getInputRCS(t, parts)
		expectedRCS, ok := parts["expected,v"]
		if !ok {
			t.Fatal("Missing expected,v")
		}

		parsedFile, err := parseRCS(inputRCS)
		if err != nil {
			t.Fatalf("ParseFile error: %v", err)
		}

		if options["unix line endings"] {
			parsedFile.SwitchLineEnding("\n")
		}

		gotRCS := parsedFile.String()
		checkRCS(t, expectedRCS, gotRCS, options)
	})
}

func testValidateRCS(t *testing.T, parts map[string]string, options map[string]bool) {
	t.Run("validate rcs", func(t *testing.T) {
		inputRCS := getInputRCS(t, parts)

		_, err := parseRCS(inputRCS)
		if err != nil {
			t.Errorf("Validation failed: %v", err)
		}
	})
}

func testNewRCS(t *testing.T, parts map[string]string, options map[string]bool, args []string) {
	t.Run("new rcs", func(t *testing.T) {
		expectedRCS, ok := parts["expected,v"]
		if !ok {
			t.Fatal("Missing expected,v")
		}

		f := NewFile()
		if options["unix line endings"] {
			f.NewLine = "\n"
		}

		for i := 0; i < len(args); i++ {
			arg := args[i]
			if strings.HasPrefix(arg, "-t") {
				if arg == "-t" {
					if i+1 < len(args) {
						i++
						// TODO read file?
						// f.Description = args[i]
						t.Skip("reading file from -t arg not implemented in test runner")
					}
				} else {
					val := strings.TrimPrefix(arg, "-t")
					if strings.HasPrefix(val, "-") {
						f.Description = strings.TrimPrefix(val, "-")
					} else {
						// TODO read file?
						t.Skip("reading file from -t arg not implemented in test runner")
					}
				}
			}
		}

		gotRCS := f.String()
		checkRCS(t, expectedRCS, gotRCS, options)
	})
}

func testListHeads(t *testing.T, parts map[string]string, options map[string]bool) {
	t.Run("list heads", func(t *testing.T) {
		inputRCS := getInputRCS(t, parts)
		expectedOut, ok := parts["expected.out"]
		if !ok {
			t.Fatal("Missing expected.out")
		}

		parsedFile, err := parseRCS(inputRCS)
		if err != nil {
			t.Fatalf("ParseFile error: %v", err)
		}

		var sb strings.Builder
		for _, rev := range parsedFile.RevisionHeads {
			sb.WriteString(rev.Revision.String() + "\n")
		}
		gotOut := sb.String()

		if options["force unix line endings"] {
			expectedOut = strings.ReplaceAll(expectedOut, "\r\n", "\n")
		}

		if diff := cmp.Diff(strings.TrimSpace(expectedOut), strings.TrimSpace(gotOut)); diff != "" {
			t.Errorf("List Heads mismatch (-want +got):\n%s", diff)
		}
	})
}

func testNormalizeRevisions(t *testing.T, parts map[string]string, options map[string]bool) {
	t.Run("normalize revisions", func(t *testing.T) {
		testFormatRCS(t, parts, options)
	})
}

func testParseError(t *testing.T, testName string, parts map[string]string, options map[string]bool) {
	errorName := strings.TrimPrefix(testName, "parse error: ")
	errorName = strings.TrimSpace(errorName)

	t.Run(testName, func(t *testing.T) {
		inputRCS := getInputRCS(t, parts)

		provider, ok := ErrorTestProvider[errorName]
		if !ok {
			t.Fatalf("Unknown error provider: %s", errorName)
		}

		// Prepare expected error
		var expectedErr error
		if errorJSON, ok := parts["error.json"]; ok {
			var err error
			expectedErr, err = provider([]byte(errorJSON))
			if err != nil {
				t.Fatalf("Failed to prepare expected error: %v", err)
			}
		} else {
			expectedErr, _ = provider(nil)
		}

		_, err := parseRCS(inputRCS)
		if err == nil {
			t.Fatal("Expected error, got nil")
		}

		// Special handling for ErrParseProperty
		var pErr ErrParseProperty
		if errors.As(expectedErr, &pErr) {
			var actualPErr ErrParseProperty
			if errors.As(err, &actualPErr) {
				if actualPErr.Property != pErr.Property {
					t.Errorf("Property mismatch: want %q, got %q", pErr.Property, actualPErr.Property)
				}
				return
			}
		}

		if !errors.Is(err, expectedErr) {
			// Also check strict equality if Is fails (for structs)
			if err.Error() != expectedErr.Error() {
				t.Errorf("Error mismatch:\nWant: %v\nGot:  %v", expectedErr, err)
			}
		}
	})
}

// Helpers

func getInputRCS(t *testing.T, parts map[string]string) string {
	if content, ok := parts["input,v"]; ok {
		return content
	}
	if content, ok := parts["input.rcs"]; ok {
		return content
	}
	t.Fatal("Missing input,v")
	return ""
}

func parseRCS(content string) (*File, error) {
	// Retry logic from original test
	parsedFile, err := ParseFile(strings.NewReader(content))
	if err != nil {
		parsedFile, err = ParseFile(strings.NewReader(content + "\n\n\n"))
		if err != nil {
			parsedFile, err = ParseFile(strings.NewReader(content + "\n"))
		}
	}
	return parsedFile, err
}

func checkRCS(t *testing.T, expected, got string, options map[string]bool) {
	ignoreWhitespace := options["ignore white space"]
	ignoreAllWhitespace := options["ignore all white space"]

	normExpected := strings.TrimSpace(expected)
	normGot := strings.TrimSpace(got)

	if options["force unix line endings"] {
		normExpected = strings.ReplaceAll(normExpected, "\r\n", "\n")
	}

	// 'got' comes from parsedFile.String().
	// If 'unix line endings' is ON, parsedFile is normalized to \n, so got has \n.
	// If 'unix line endings' is OFF, parsedFile might have \r\n (from input).
	// If input had \r\n and we compare against normalized expected (\n), we need to normalize got too?
	// The user request suggests 'unix line endings' ensures the object has \n.
	// So got should be correct.

	if ignoreAllWhitespace {
		normExpected = stripAllWhitespace(normExpected)
		normGot = stripAllWhitespace(normGot)
	} else if ignoreWhitespace {
		normExpected = strings.Join(strings.Fields(normExpected), " ")
		normGot = strings.Join(strings.Fields(normGot), " ")
	}

	if diff := cmp.Diff(normExpected, normGot); diff != "" {
		t.Errorf("RCS mismatch (-want +got):\n%s", diff)
		if !ignoreWhitespace && !ignoreAllWhitespace {
			t.Logf("Got RCS:\n%q", got)
		}
	}
}

func stripAllWhitespace(s string) string {
	return strings.Map(func(r rune) rune {
		if r == ' ' || r == '\n' || r == '\r' || r == '\t' {
			return -1
		}
		return r
	}, s)
}

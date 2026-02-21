package cli

import (
	"encoding/json"
	"flag"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"

	"golang.org/x/tools/txtar"
)

func TestCliTxtar(t *testing.T) {
	files, err := filepath.Glob("testdata/co_permissions_*.txtar")
	if err != nil {
		t.Fatal(err)
	}
	if len(files) == 0 {
		// Fallback if running from root?
		files, err = filepath.Glob("internal/cli/testdata/co_permissions_*.txtar")
		if err != nil {
			t.Fatal(err)
		}
	}

	for _, file := range files {
		t.Run(filepath.Base(file), func(t *testing.T) {
			runCliTest(t, file)
		})
	}
}

func runCliTest(t *testing.T, path string) {
	archive, err := txtar.ParseFile(path)
	if err != nil {
		t.Fatal(err)
	}

	tmpDir := t.TempDir()

	// Parse options.conf
	var opts struct {
		RCSMode string `json:"rcs_file_perm"` // Octal string
	}

	// Parse tests.txt for args
	var testCommands []string

	for _, f := range archive.Files {
		switch f.Name {
		case "options.conf", "options.json":
			if err := json.Unmarshal(f.Data, &opts); err != nil {
				t.Fatalf("json unmarshal options: %v", err)
			}
		case "tests.txt":
			lines := strings.Split(string(f.Data), "\n")
			for _, line := range lines {
				line = strings.TrimSpace(line)
				if line != "" {
					testCommands = append(testCommands, line)
				}
			}
		}
	}

	// Extract files
	for _, f := range archive.Files {
		name := f.Name
		data := f.Data
		if name == "options.conf" || name == "options.json" || name == "tests.txt" || strings.HasPrefix(name, "expected.") {
			continue
		}

		dest := filepath.Join(tmpDir, name)
		// Default permissions, might be overridden by opts
		if err := os.WriteFile(dest, data, 0644); err != nil {
			t.Fatal(err)
		}
	}

	// Apply RCS mode if needed
	if opts.RCSMode != "" {
		mode, err := strconv.ParseUint(opts.RCSMode, 8, 32)
		if err != nil {
			t.Fatalf("invalid rcs_mode: %v", err)
		}

		// Find RCS files (ending in ,v)
		entries, err := os.ReadDir(tmpDir)
		if err != nil {
			t.Fatal(err)
		}
		for _, e := range entries {
			if strings.HasSuffix(e.Name(), ",v") || strings.HasSuffix(e.Name(), ".rcs") {
				rcsFile := filepath.Join(tmpDir, e.Name())
				if err := os.Chmod(rcsFile, fs.FileMode(mode)); err != nil {
					t.Fatalf("chmod failed: %v", err)
				}
			}
		}
	}

	// Run commands
	for _, cmd := range testCommands {
		args := strings.Fields(cmd)
		if len(args) == 0 {
			continue
		}
		if args[0] == "co" || args[0] == "co:" {
			// Parse args for Co
			// Strip "co" or "co:"
			runCo(t, args[1:], tmpDir)
		}
	}

	// Verify expectations
	for _, f := range archive.Files {
		if f.Name == "expected.mode" {
			// Find working file (should be 'input' usually, but depends on test)
			// Assuming single working file for simplicity or deduced from test
			// The expected.mode file content is the mode
			// We check 'input' file permissions?

			// Hardcoded assumption for this specific test case: input file is named "input"
			workFilePath := filepath.Join(tmpDir, "input")

			fi, err := os.Stat(workFilePath)
			if err != nil {
				t.Fatal(err)
			}
			got := fi.Mode().Perm()
			wantStr := strings.TrimSpace(string(f.Data))
			want, err := strconv.ParseUint(wantStr, 8, 32)
			if err != nil {
				t.Fatalf("invalid expected.mode: %v", err)
			}

			if runtime.GOOS == "windows" {
				// Windows logic
				wantWrite := fs.FileMode(want) & 0222
				gotWrite := got & 0222

				if wantWrite == 0 && gotWrite != 0 {
					t.Errorf("Mode mismatch on Windows: want read-only, got writable (mode %o)", got)
				}
				if wantWrite != 0 && gotWrite == 0 {
					t.Errorf("Mode mismatch on Windows: want writable, got read-only (mode %o)", got)
				}
			} else {
				// Unix strict check
				if got != fs.FileMode(want) {
					t.Errorf("Mode mismatch: want %o, got %o", want, got)
				}
			}
		}
	}
}

func runCo(t *testing.T, args []string, dir string) {
	// Parse args to call Co
	f := flag.NewFlagSet("co", flag.ContinueOnError)
	revision := f.String("r", "", "")
	lock := f.Bool("l", false, "")
	unlock := f.Bool("u", false, "")
	user := f.String("w", "", "")
	quiet := f.Bool("q", false, "")
	date := f.String("d", "", "")
	zone := f.String("z", "", "")

	if err := f.Parse(args); err != nil {
		t.Fatalf("flag parse: %v", err)
	}

	files := f.Args()
	if len(files) == 0 {
		t.Fatal("no files for co")
	}

	// Prepend dir to files
	var absFiles []string
	for _, file := range files {
		absFiles = append(absFiles, filepath.Join(dir, file))
	}

	// Ensure Co is available (it's in the same package)
	err := Co(*revision, *lock, *unlock, *user, *quiet, *date, *zone, absFiles...)
	if err != nil {
		t.Fatalf("Co failed: %v", err)
	}
}

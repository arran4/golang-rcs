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

	var opts struct {
		Args     []string `json:"args"`
		WorkFile string   `json:"workfile"` // The file to run Co on (without ,v)
		RCSMode  string   `json:"rcs_mode"` // Octal string
	}

	// Extract files
	for _, f := range archive.Files {
		name := f.Name
		data := f.Data
		if name == "options.json" {
			if err := json.Unmarshal(data, &opts); err != nil {
				t.Fatalf("json unmarshal options: %v", err)
			}
			continue
		}
		if strings.HasPrefix(name, "expected.") {
			continue
		}

		dest := filepath.Join(tmpDir, name)
		if strings.HasSuffix(name, ",v") {
			// Write with 0600 initially, strict mode will be applied later if needed
			// But for now just write it.
			if err := os.WriteFile(dest, data, 0644); err != nil {
				t.Fatal(err)
			}
		} else {
			if err := os.WriteFile(dest, data, 0644); err != nil {
				t.Fatal(err)
			}
		}
	}

	// Apply RCS mode if needed
	if opts.RCSMode != "" && opts.WorkFile != "" {
		mode, err := strconv.ParseUint(opts.RCSMode, 8, 32)
		if err != nil {
			t.Fatalf("invalid rcs_mode: %v", err)
		}
		rcsFile := filepath.Join(tmpDir, opts.WorkFile+",v")
		if err := os.Chmod(rcsFile, fs.FileMode(mode)); err != nil {
			t.Fatalf("chmod failed: %v", err)
		}
	}

	// Parse args to call Co
	f := flag.NewFlagSet("co", flag.ContinueOnError)
	revision := f.String("r", "", "")
	lock := f.Bool("l", false, "")
	unlock := f.Bool("u", false, "")
	user := f.String("w", "", "")
	quiet := f.Bool("q", false, "")
	date := f.String("d", "", "")
	zone := f.String("z", "", "")

	if err := f.Parse(opts.Args); err != nil {
		t.Fatalf("flag parse: %v", err)
	}

	workFilePath := filepath.Join(tmpDir, opts.WorkFile)

	// Ensure Co is available (it's in the same package)
	err = Co(*revision, *lock, *unlock, *user, *quiet, *date, *zone, workFilePath)
	if err != nil {
		t.Fatalf("Co failed: %v", err)
	}

	// Verify
	for _, f := range archive.Files {
		if f.Name == "expected.mode" {
			// Check mode
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

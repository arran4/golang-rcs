package rcs

import (
	"bytes"
	"github.com/google/go-cmp/cmp"
	"io/fs"
	"strings"
	"testing"
)


func TestStringLocalFiles(t *testing.T) {
	testRoundTrip(t, localTests, "testdata/local")
}

func TestStringRepoFiles(t *testing.T) {
	// Placeholder for future repo data tests
	// testRoundTrip(t, repoTests, "testdata/repo")
}

func testRoundTrip(t *testing.T, fsys fs.FS, root string) {
	err := fs.WalkDir(fsys, root, func(path string, d fs.DirEntry, err error) error {
		if d == nil {
			return nil
		}
		if d.IsDir() {
			return nil
		}
		if !strings.HasSuffix(path, ",v") {
			return nil
		}
		t.Run(path, func(t *testing.T) {
			b, err := fs.ReadFile(fsys, path)
			if err != nil {
				t.Errorf("ReadFile( %s ) error = %s", path, err)
				return
			}
			got, err := ParseFile(bytes.NewReader(b))
			if err != nil {
				t.Errorf("ParseFile( %s ) error = %s", path, err)
				return
			}
			if diff := cmp.Diff(strings.Split(got.String(), "\n"), strings.Split(string(b), "\n")); diff != "" {
				t.Errorf("String(): %s", diff)
			}
		})
		return nil
	})
	if err != nil {
		t.Logf("WalkDir error: %v", err)
	}
}

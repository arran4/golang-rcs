package rcs

import (
	"bytes"
	"io/fs"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestCircularParseGenerateLocalFiles(t *testing.T) {
	path := "testdata/local"
	err := fs.WalkDir(localTests, path, func(path string, d fs.DirEntry, err error) error {
		if d.IsDir() {
			return nil
		}
		if !strings.HasSuffix(path, ",v") {
			return nil
		}
		t.Run(path, func(t *testing.T) {
			b, err := fs.ReadFile(localTests, path)
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
		t.Fatalf("Error: %s", err)
	}
}

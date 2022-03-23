package rcs

import (
	"bytes"
	"embed"
	"github.com/google/go-cmp/cmp"
	"io/fs"
	"path"
	"strings"
	"testing"
)

var (
	//go:embed "testdata/local/*"
	localTests embed.FS
)

func TestParseLocalFiles(t *testing.T) {
	dir := "testdata/local"
	d, err := localTests.ReadDir(dir)
	if err != nil {
		t.Fatalf("Error: %s", err)
	}
	for _, tt := range d {
		if !strings.HasSuffix(tt.Name(), ",v") {
			continue
		}
		t.Run(tt.Name(), func(t *testing.T) {
			b, err := fs.ReadFile(localTests, path.Join(dir, tt.Name()))
			if err != nil {
				t.Errorf("ReadFile() error = %s", err)
				return
			}
			got, err := ParseFile(bytes.NewReader(b))
			if err != nil {
				t.Errorf("ParseFile() error = %s", err)
				return
			}
			if diff := cmp.Diff(got.String(), string(b)); diff != "" {
				t.Errorf("String(): %s", diff)
			}
		})
	}
}

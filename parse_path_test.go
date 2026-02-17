package rcs

import (
	"embed"
	"os"
	"path/filepath"
	"testing"
)

//go:embed testdata/testinput.go,v
var parsePathFixtures embed.FS

func TestParsePath(t *testing.T) {
	input, err := parsePathFixtures.ReadFile("testdata/testinput.go,v")
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}

	tmpPath := filepath.Join(t.TempDir(), "testinput.go,v")
	if err := os.WriteFile(tmpPath, input, 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	t.Run("os-open", func(t *testing.T) {
		f, err := ParsePath(tmpPath)
		if err != nil {
			t.Fatalf("ParsePath() error = %v", err)
		}
		if f.Head == "" {
			t.Fatalf("ParsePath().Head = %q, want non-empty", f.Head)
		}
	})

	t.Run("mmap", func(t *testing.T) {
		f, err := ParsePath(tmpPath, WithMmap(true))
		if err != nil {
			t.Fatalf("ParsePath(WithMmap(true)) error = %v", err)
		}
		if f.Head == "" {
			t.Fatalf("ParsePath(WithMmap(true)).Head = %q, want non-empty", f.Head)
		}
	})
}

func TestParsePathErrors(t *testing.T) {
	_, err := ParsePath(filepath.Join(t.TempDir(), "missing.go,v"), WithMmap(true))
	if err == nil {
		t.Fatal("ParsePath(missing, WithMmap(true)) error = nil, want error")
	}
}

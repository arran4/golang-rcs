package cli

import (
	"os"
	"path/filepath"
	"testing"
)

func TestInitPermissions(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "gorcs_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	workFile := filepath.Join(tempDir, "testfile")
	// Create dummy work file
	if err := os.WriteFile(workFile, []byte("content"), 0644); err != nil {
		t.Fatal(err)
	}

	rcsFile := workFile + ",v"

	if err := initFile("description", workFile); err != nil {
		t.Fatalf("initFile failed: %v", err)
	}

	info, err := os.Stat(rcsFile)
	if err != nil {
		t.Fatal(err)
	}

	mode := info.Mode().Perm()
	// We expect this to fail if we consider 0444 insecure
	if mode&0004 != 0 {
		t.Errorf("RCS file is world-readable: %o", mode)
	}
}

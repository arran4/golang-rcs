package cli

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	rcs "github.com/arran4/golang-rcs"
)

func TestInitPermissions(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "gorcs_test")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = os.RemoveAll(tempDir)
	}()

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
	if runtime.GOOS != "windows" {
		if mode&0004 != 0 {
			t.Errorf("RCS file is world-readable: %o", mode)
		}
	}
}

func TestInitPermissionsEnv(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "gorcs_test_env")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = os.RemoveAll(tempDir)
	}()

	t.Setenv("GORCS_INIT_MODE", "0644")

	workFile := filepath.Join(tempDir, "testfile_env")
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
	if runtime.GOOS != "windows" {
		if mode&0004 == 0 {
			t.Errorf("RCS file is NOT world-readable but env set 0644: %o", mode)
		}
	}
}

func TestCoIntegration(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "gorcs_integration_test")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = os.RemoveAll(tempDir)
	}()

	workFile := filepath.Join(tempDir, "workfile")
	rcsFile := workFile + ",v"

	// Create dummy work file
	if err := os.WriteFile(workFile, []byte("content"), 0644); err != nil {
		t.Fatal(err)
	}

	// 1. Manually create a valid RCS file with one revision
	f := rcs.NewFile()
	f.Head = "1.1"
	f.Access = true
	f.AccessUsers = []string{"tester"}
	f.Comment = " comment "
	f.Expand = "kv"
	f.Description = "description"

	f.RevisionHeads = []*rcs.RevisionHead{
		{
			Revision: rcs.Num("1.1"),
			Date:     rcs.DateTime(time.Now().Format(rcs.DateFormat)),
			Author:   rcs.ID("tester"),
			State:    rcs.ID("Exp"),
		},
	}
	f.RevisionContents = []*rcs.RevisionContent{
		{
			Revision: "1.1",
			Log:      "Initial revision",
			Text:     "content",
		},
	}

	// Write with 0600 permissions
	if err := os.WriteFile(rcsFile, []byte(f.String()), 0600); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	info, err := os.Stat(rcsFile)
	if err != nil {
		t.Fatal(err)
	}
	// Check perm
	mode := info.Mode().Perm()
	// On Windows, 0600 might not prevent world readability as intended by mode bits.
	if runtime.GOOS != "windows" {
		if mode&0004 != 0 {
			t.Errorf("RCS file is world-readable: %o", mode)
		}
	}
	if mode&0200 == 0 {
		t.Errorf("RCS file is not writable by owner: %o", mode)
	}

	// 2. Checkout with lock (modifies RCS file)
	// We ask to lock revision 1.1
	// Co(revision string, lock, unlock bool, user string, quiet bool, checkoutDate, checkoutZone string, files ...string)
	if err := Co("1.1", true, false, "tester", true, "", "", workFile); err != nil {
		t.Fatalf("Co -l failed: %v", err)
	}

	// Check if lock is set in RCS file
	content, err := os.ReadFile(rcsFile)
	if err != nil {
		t.Fatal(err)
	}
	// Simple string check for lock
	if !strings.Contains(string(content), "tester:1.1") {
		t.Errorf("Lock not found in RCS file content: %s", content)
	}

	// 3. Check permissions again. Co might have rewritten it.
	info, err = os.Stat(rcsFile)
	if err != nil {
		t.Fatal(err)
	}
	mode = info.Mode().Perm()

	if runtime.GOOS != "windows" {
		if mode&0004 != 0 {
			t.Errorf("RCS file became world-readable after Co: %o", mode)
		}
	}
}

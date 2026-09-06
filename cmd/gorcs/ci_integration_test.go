// Integration tests that exercise the `ci` command through the real command
// execution path (RootCmd.Execute -> cli.Ci -> filesystem), as opposed to the
// flag-parsing tests in ci_test.go that inject a CommandAction.
package main

import (
	"os"
	"path/filepath"
	"testing"

	rcs "github.com/arran4/golang-rcs"
)

// runCi builds a fresh root command and runs `ci` with the given arguments.
func runCi(t *testing.T, args ...string) error {
	t.Helper()
	root, err := NewRoot("gorcs", "", "", "")
	if err != nil {
		t.Fatalf("NewRoot: %v", err)
	}
	return root.Execute(append([]string{"ci"}, args...))
}

// parseArchive parses an RCS archive file from disk.
func parseArchive(t *testing.T, path string) *rcs.File {
	t.Helper()
	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("open archive %s: %v", path, err)
	}
	defer func() { _ = f.Close() }()
	parsed, err := rcs.ParseFile(f)
	if err != nil {
		t.Fatalf("parse archive %s: %v", path, err)
	}
	return parsed
}

func TestCiIntegration_InitialCheckin(t *testing.T) {
	dir := t.TempDir()
	wf := filepath.Join(dir, "file.txt")
	if err := os.WriteFile(wf, []byte("hello\n"), 0644); err != nil {
		t.Fatalf("write working file: %v", err)
	}

	if err := runCi(t, "-l", "-w", "tester", "-m", "initial", wf); err != nil {
		t.Fatalf("ci initial: %v", err)
	}

	archive := parseArchive(t, wf+",v")
	if archive.Head != "1.1" {
		t.Errorf("head = %q, want 1.1", archive.Head)
	}
	v, err := archive.Checkout("tester", rcs.WithRevision("1.1"))
	if err != nil {
		t.Fatalf("checkout 1.1: %v", err)
	}
	if v.Content != "hello\n" {
		t.Errorf("1.1 content = %q, want %q", v.Content, "hello\n")
	}
}

func TestCiIntegration_SecondRevision(t *testing.T) {
	dir := t.TempDir()
	wf := filepath.Join(dir, "file.txt")
	if err := os.WriteFile(wf, []byte("one\n"), 0644); err != nil {
		t.Fatalf("write working file: %v", err)
	}

	// Initial checkin with lock so the strict-lock second checkin is permitted.
	if err := runCi(t, "-l", "-w", "tester", "-m", "v1", wf); err != nil {
		t.Fatalf("ci v1: %v", err)
	}
	// -l kept the working file writable; change it and check in again.
	if err := os.WriteFile(wf, []byte("one\ntwo\n"), 0644); err != nil {
		t.Fatalf("rewrite working file: %v", err)
	}
	if err := runCi(t, "-l", "-w", "tester", "-m", "v2", wf); err != nil {
		t.Fatalf("ci v2: %v", err)
	}

	archive := parseArchive(t, wf+",v")
	if archive.Head != "1.2" {
		t.Errorf("head = %q, want 1.2", archive.Head)
	}
	if len(archive.RevisionHeads) != 2 {
		t.Fatalf("revision count = %d, want 2", len(archive.RevisionHeads))
	}
	// Both revisions must reconstruct correctly (older via reverse delta).
	head, err := archive.Checkout("tester", rcs.WithRevision("1.2"))
	if err != nil {
		t.Fatalf("checkout 1.2: %v", err)
	}
	if head.Content != "one\ntwo\n" {
		t.Errorf("1.2 content = %q, want %q", head.Content, "one\ntwo\n")
	}
	prev, err := archive.Checkout("tester", rcs.WithRevision("1.1"))
	if err != nil {
		t.Fatalf("checkout 1.1: %v", err)
	}
	if prev.Content != "one\n" {
		t.Errorf("1.1 content = %q, want %q", prev.Content, "one\n")
	}
}

func TestCiIntegration_LockDisposition(t *testing.T) {
	dir := t.TempDir()
	wf := filepath.Join(dir, "file.txt")
	if err := os.WriteFile(wf, []byte("x\n"), 0644); err != nil {
		t.Fatalf("write working file: %v", err)
	}
	if err := runCi(t, "-l", "-w", "tester", "-m", "m", wf); err != nil {
		t.Fatalf("ci -l: %v", err)
	}
	info, err := os.Stat(wf)
	if err != nil {
		t.Fatalf("working file should be kept with -l: %v", err)
	}
	if info.Mode().Perm()&0200 == 0 {
		t.Errorf("working file mode = %o, want writable with -l", info.Mode().Perm())
	}
}

func TestCiIntegration_UnlockDisposition(t *testing.T) {
	dir := t.TempDir()
	wf := filepath.Join(dir, "file.txt")
	if err := os.WriteFile(wf, []byte("x\n"), 0644); err != nil {
		t.Fatalf("write working file: %v", err)
	}
	if err := runCi(t, "-u", "-w", "tester", "-m", "m", wf); err != nil {
		t.Fatalf("ci -u: %v", err)
	}
	info, err := os.Stat(wf)
	if err != nil {
		t.Fatalf("working file should be kept with -u: %v", err)
	}
	if info.Mode().Perm()&0200 != 0 {
		t.Errorf("working file mode = %o, want read-only with -u", info.Mode().Perm())
	}
}

func TestCiIntegration_RemoveDisposition(t *testing.T) {
	dir := t.TempDir()
	wf := filepath.Join(dir, "file.txt")
	if err := os.WriteFile(wf, []byte("x\n"), 0644); err != nil {
		t.Fatalf("write working file: %v", err)
	}
	if err := runCi(t, "-w", "tester", "-m", "m", wf); err != nil {
		t.Fatalf("ci (remove): %v", err)
	}
	if _, err := os.Stat(wf); !os.IsNotExist(err) {
		t.Errorf("working file should be removed without -l/-u, stat err = %v", err)
	}
	// Archive must still be created.
	parseArchive(t, wf+",v")
}

func TestCiIntegration_ForceUnchanged(t *testing.T) {
	dir := t.TempDir()
	wf := filepath.Join(dir, "file.txt")
	if err := os.WriteFile(wf, []byte("same\n"), 0644); err != nil {
		t.Fatalf("write working file: %v", err)
	}
	if err := runCi(t, "-l", "-w", "tester", "-m", "v1", wf); err != nil {
		t.Fatalf("ci v1: %v", err)
	}
	// Force a checkin of identical content; -f overrides the no-change guard,
	// and the lock held from -l satisfies strict locking.
	if err := runCi(t, "-l", "-f", "-w", "tester", "-m", "v2", wf); err != nil {
		t.Fatalf("ci -f unchanged: %v", err)
	}
	archive := parseArchive(t, wf+",v")
	if archive.Head != "1.2" {
		t.Errorf("head = %q, want 1.2 after forced checkin", archive.Head)
	}
}

func TestCiIntegration_PreservesExecutableMode(t *testing.T) {
	dir := t.TempDir()
	wf := filepath.Join(dir, "script.sh")
	if err := os.WriteFile(wf, []byte("echo hi\n"), 0755); err != nil {
		t.Fatalf("write working file: %v", err)
	}
	if err := runCi(t, "-u", "-w", "tester", "-m", "m", wf); err != nil {
		t.Fatalf("ci -u: %v", err)
	}
	// Real RCS derives the archive mode from the working file with the write
	// bits cleared (0755 -> 0555); execute/group bits must be preserved, not
	// flattened to 0444.
	archInfo, err := os.Stat(wf + ",v")
	if err != nil {
		t.Fatalf("stat archive: %v", err)
	}
	if got := archInfo.Mode().Perm(); got != 0555 {
		t.Errorf("archive mode = %o, want 0555 (execute preserved, read-only)", got)
	}
	// The -u working file must likewise stay executable but read-only.
	wfInfo, err := os.Stat(wf)
	if err != nil {
		t.Fatalf("stat working file: %v", err)
	}
	if got := wfInfo.Mode().Perm(); got != 0555 {
		t.Errorf("working file mode = %o, want 0555", got)
	}
}

func TestCiIntegration_ConflictingFlags(t *testing.T) {
	dir := t.TempDir()
	wf := filepath.Join(dir, "file.txt")
	if err := os.WriteFile(wf, []byte("x\n"), 0644); err != nil {
		t.Fatalf("write working file: %v", err)
	}
	err := runCi(t, "-l", "-u", "-w", "tester", "-m", "m", wf)
	if err == nil {
		t.Fatal("expected an error when both -l and -u are given")
	}
	// Archive must not have been created.
	if _, statErr := os.Stat(wf + ",v"); !os.IsNotExist(statErr) {
		t.Errorf("archive should not be created on conflicting flags, stat err = %v", statErr)
	}
}

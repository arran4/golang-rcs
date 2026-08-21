package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	rcs "github.com/arran4/golang-rcs"
	"github.com/google/go-cmp/cmp"
)

// writeFile creates a file under dir with the given content and mode.
func writeFile(t *testing.T, dir, name, content string, mode os.FileMode) string {
	t.Helper()
	p := filepath.Join(dir, name)
	if err := os.WriteFile(p, []byte(content), mode); err != nil {
		t.Fatalf("write %s: %v", p, err)
	}
	return p
}

// readArchive parses an RCS archive from the given path.
func readArchive(t *testing.T, path string) *rcs.File {
	t.Helper()
	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("open %s: %v", path, err)
	}
	defer func() { _ = f.Close() }()
	parsed, err := rcs.ParseFile(f)
	if err != nil {
		t.Fatalf("parse %s: %v", path, err)
	}
	return parsed
}

func TestCiInitialCheckin(t *testing.T) {
	dir := t.TempDir()
	wf := writeFile(t, dir, "test.txt", "hello world\n", 0644)

	err := Ci("", CiLock, "alice", true, "initial", false, "", "", "", wf)
	if err != nil {
		t.Fatalf("Ci: %v", err)
	}

	// Archive should exist.
	archive := readArchive(t, wf+",v")
	if archive.Head != "1.1" {
		t.Errorf("head = %s, want 1.1", archive.Head)
	}

	// Working file should still exist and be writable (CiLock).
	info, err := os.Stat(wf)
	if err != nil {
		t.Fatalf("stat working file: %v", err)
	}
	if info.Mode().Perm()&0200 == 0 {
		t.Error("working file should be writable with CiLock")
	}
}

func TestCiInitialRemove(t *testing.T) {
	dir := t.TempDir()
	wf := writeFile(t, dir, "test.txt", "content\n", 0644)

	err := Ci("", CiRemove, "alice", true, "init", false, "", "", "", wf)
	if err != nil {
		t.Fatalf("Ci: %v", err)
	}

	// Working file should be removed.
	if _, err := os.Stat(wf); !os.IsNotExist(err) {
		t.Error("working file should be removed with CiRemove")
	}

	// Archive should exist.
	readArchive(t, wf+",v")
}

func TestCiInitialUnlock(t *testing.T) {
	dir := t.TempDir()
	wf := writeFile(t, dir, "test.txt", "content\n", 0644)

	err := Ci("", CiUnlock, "alice", true, "init", false, "", "", "", wf)
	if err != nil {
		t.Fatalf("Ci: %v", err)
	}

	// Working file should be read-only.
	info, err := os.Stat(wf)
	if err != nil {
		t.Fatalf("stat: %v", err)
	}
	if info.Mode().Perm()&0200 != 0 {
		t.Error("working file should be read-only with CiUnlock")
	}
}

func TestCiSecondRevision(t *testing.T) {
	dir := t.TempDir()
	wf := writeFile(t, dir, "test.txt", "version 1\n", 0644)

	// First checkin.
	err := Ci("", CiLock, "alice", true, "first", false, "", "", "", wf)
	if err != nil {
		t.Fatalf("first Ci: %v", err)
	}

	// Modify working file.
	if err := os.WriteFile(wf, []byte("version 2\n"), 0644); err != nil {
		t.Fatal(err)
	}

	// Second checkin.
	err = Ci("", CiLock, "alice", true, "second", false, "", "", "", wf)
	if err != nil {
		t.Fatalf("second Ci: %v", err)
	}

	archive := readArchive(t, wf+",v")
	if archive.Head != "1.2" {
		t.Errorf("head = %s, want 1.2", archive.Head)
	}

	// Checkout head should give version 2.
	v, err := archive.Checkout("alice")
	if err != nil {
		t.Fatalf("checkout: %v", err)
	}
	if diff := cmp.Diff("version 2\n", v.Content); diff != "" {
		t.Errorf("head content mismatch (-want +got):\n%s", diff)
	}

	// Checkout 1.1 should give version 1.
	v, err = archive.Checkout("alice", rcs.WithRevision("1.1"))
	if err != nil {
		t.Fatalf("checkout 1.1: %v", err)
	}
	if diff := cmp.Diff("version 1\n", v.Content); diff != "" {
		t.Errorf("1.1 content mismatch (-want +got):\n%s", diff)
	}
}

func TestCiExplicitRevision(t *testing.T) {
	dir := t.TempDir()
	wf := writeFile(t, dir, "test.txt", "content\n", 0644)

	err := Ci("2.1", CiLock, "alice", true, "custom rev", false, "", "", "", wf)
	if err != nil {
		t.Fatalf("Ci: %v", err)
	}

	archive := readArchive(t, wf+",v")
	if archive.Head != "2.1" {
		t.Errorf("head = %s, want 2.1", archive.Head)
	}
}

func TestCiForceUnchanged(t *testing.T) {
	dir := t.TempDir()
	wf := writeFile(t, dir, "test.txt", "same\n", 0644)

	// Initial.
	err := Ci("", CiLock, "alice", true, "init", false, "", "", "", wf)
	if err != nil {
		t.Fatal(err)
	}

	// Attempt without force — should fail.
	err = Ci("", CiLock, "alice", true, "no change", false, "", "", "", wf)
	if err == nil {
		t.Fatal("expected error for unchanged content without force")
	}
	if !strings.Contains(err.Error(), "no changes") {
		t.Errorf("unexpected error: %v", err)
	}

	// With force — should succeed.
	err = Ci("", CiLock, "alice", true, "forced", true, "", "", "", wf)
	if err != nil {
		t.Fatalf("forced Ci: %v", err)
	}

	archive := readArchive(t, wf+",v")
	if archive.Head != "1.2" {
		t.Errorf("head = %s, want 1.2", archive.Head)
	}
}

func TestCiWithState(t *testing.T) {
	dir := t.TempDir()
	wf := writeFile(t, dir, "test.txt", "content\n", 0644)

	err := Ci("", CiLock, "alice", true, "init", false, "Rel", "", "", wf)
	if err != nil {
		t.Fatal(err)
	}

	archive := readArchive(t, wf+",v")
	if len(archive.RevisionHeads) == 0 {
		t.Fatal("no revision heads")
	}
	if string(archive.RevisionHeads[0].State) != "Rel" {
		t.Errorf("state = %s, want Rel", archive.RevisionHeads[0].State)
	}
}

func TestCiWithDate(t *testing.T) {
	dir := t.TempDir()
	wf := writeFile(t, dir, "test.txt", "content\n", 0644)

	err := Ci("", CiLock, "alice", true, "init", false, "", "2024-06-15 12:00:00", "", wf)
	if err != nil {
		t.Fatal(err)
	}

	archive := readArchive(t, wf+",v")
	if len(archive.RevisionHeads) == 0 {
		t.Fatal("no revision heads")
	}
	date := string(archive.RevisionHeads[0].Date)
	if !strings.HasPrefix(date, "2024.06.15") {
		t.Errorf("date = %s, want prefix 2024.06.15", date)
	}
}

func TestCiArchiveArgument(t *testing.T) {
	dir := t.TempDir()
	wf := writeFile(t, dir, "test.txt", "content\n", 0644)

	// Pass the archive name instead of working file name.
	err := Ci("", CiLock, "alice", true, "init", false, "", "", "", wf+",v")
	if err != nil {
		t.Fatal(err)
	}

	archive := readArchive(t, wf+",v")
	if archive.Head != "1.1" {
		t.Errorf("head = %s, want 1.1", archive.Head)
	}
}

func TestCiNoFiles(t *testing.T) {
	err := Ci("", CiRemove, "alice", true, "msg", false, "", "", "")
	if err == nil {
		t.Fatal("expected error for no files")
	}
	if !strings.Contains(err.Error(), "no files") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestCiMissingWorkingFile(t *testing.T) {
	dir := t.TempDir()
	err := Ci("", CiLock, "alice", true, "msg", false, "", "", "",
		filepath.Join(dir, "nonexistent.txt"))
	if err == nil {
		t.Fatal("expected error for missing working file")
	}
}

func TestCiDefaultUser(t *testing.T) {
	dir := t.TempDir()
	wf := writeFile(t, dir, "test.txt", "content\n", 0644)

	// Empty user should default to current logged-in user.
	err := Ci("", CiLock, "", true, "init", false, "", "", "", wf)
	if err != nil {
		t.Fatal(err)
	}

	archive := readArchive(t, wf+",v")
	author := string(archive.RevisionHeads[0].Author)
	if author == "" {
		t.Error("author should not be empty")
	}
}

func TestCiStatusOutput(t *testing.T) {
	dir := t.TempDir()
	wf := writeFile(t, dir, "test.txt", "content\n", 0644)

	// Capture stdout to verify status output in non-quiet mode.
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := Ci("", CiLock, "alice", false, "init", false, "", "", "", wf)
	if err != nil {
		_ = w.Close()
		os.Stdout = old
		t.Fatal(err)
	}

	_ = w.Close()
	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	os.Stdout = old

	output := buf.String()
	if !strings.Contains(output, "initial revision: 1.1") {
		t.Errorf("expected 'initial revision: 1.1' in output, got: %s", output)
	}
	if !strings.Contains(output, "done") {
		t.Errorf("expected 'done' in output, got: %s", output)
	}
}

func TestCiStatusOutputNewRevision(t *testing.T) {
	dir := t.TempDir()
	wf := writeFile(t, dir, "test.txt", "v1\n", 0644)

	// Initial (quiet).
	if err := Ci("", CiLock, "alice", true, "init", false, "", "", "", wf); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(wf, []byte("v2\n"), 0644); err != nil {
		t.Fatal(err)
	}

	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := Ci("", CiLock, "alice", false, "second", false, "", "", "", wf)
	if err != nil {
		_ = w.Close()
		os.Stdout = old
		t.Fatal(err)
	}

	_ = w.Close()
	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	os.Stdout = old

	output := buf.String()
	if !strings.Contains(output, "new revision: 1.2") {
		t.Errorf("expected 'new revision: 1.2' in output, got: %s", output)
	}
}

func TestCiMultipleFiles(t *testing.T) {
	dir := t.TempDir()
	wf1 := writeFile(t, dir, "a.txt", "file a\n", 0644)
	wf2 := writeFile(t, dir, "b.txt", "file b\n", 0644)

	err := Ci("", CiLock, "alice", true, "init", false, "", "", "", wf1, wf2)
	if err != nil {
		t.Fatal(err)
	}

	readArchive(t, wf1+",v")
	readArchive(t, wf2+",v")
}

func TestCiArchivePermissions(t *testing.T) {
	dir := t.TempDir()
	wf := writeFile(t, dir, "test.txt", "content\n", 0644)

	if err := Ci("", CiLock, "alice", true, "init", false, "", "", "", wf); err != nil {
		t.Fatal(err)
	}

	info, err := os.Stat(wf + ",v")
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0444 {
		t.Errorf("archive perm = %o, want 0444", info.Mode().Perm())
	}
}

func TestCiInvalidDate(t *testing.T) {
	dir := t.TempDir()
	wf := writeFile(t, dir, "test.txt", "content\n", 0644)

	err := Ci("", CiLock, "alice", true, "init", false, "", "not-a-date", "", wf)
	if err == nil {
		t.Fatal("expected error for invalid date")
	}
}

func TestCiInvalidZone(t *testing.T) {
	dir := t.TempDir()
	wf := writeFile(t, dir, "test.txt", "content\n", 0644)

	err := Ci("", CiLock, "alice", true, "init", false, "", "2024-01-01 12:00:00", "INVALID", wf)
	if err == nil {
		t.Fatal("expected error for invalid zone")
	}
}

func TestCiBadArchive(t *testing.T) {
	dir := t.TempDir()
	wf := writeFile(t, dir, "test.txt", "content\n", 0644)
	writeFile(t, dir, "test.txt,v", "not valid rcs", 0644)

	err := Ci("", CiLock, "alice", true, "init", false, "", "", "", wf)
	if err == nil {
		t.Fatal("expected error for bad archive")
	}
}

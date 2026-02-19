package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"golang.org/x/tools/txtar"
)

func TestBranchesDefaultSet_FromOperation3049(t *testing.T) {
	archivePath := filepath.Join("..", "..", "testdata", "txtar", "operations", "rcs-b.txtar")
	content, err := os.ReadFile(archivePath)
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	archive := txtar.Parse(content)
	parts := map[string]string{}
	for _, f := range archive.Files {
		parts[f.Name] = string(f.Data)
	}

	tmp := t.TempDir()
	workingPath := filepath.Join(tmp, "input.txt")
	rcsPath := workingPath + ",v"
	if err := os.WriteFile(workingPath, []byte(parts["input.txt"]), 0644); err != nil {
		t.Fatalf("write working file: %v", err)
	}
	if err := os.WriteFile(rcsPath, []byte(parts["input.txt,v"]), 0644); err != nil {
		t.Fatalf("write rcs file: %v", err)
	}

	if err != nil {
		t.Fatalf("NewRoot error: %v", err)
	}
	if err := cmd.Execute([]string{"branches", "default", "set", "-name", "1.1.1.1", workingPath}); err != nil {
		t.Fatalf("execute branches default set: %v", err)
	}

	got, err := os.ReadFile(rcsPath)
	if err != nil {
		t.Fatalf("read updated rcs: %v", err)
	}
	want := strings.TrimSpace(parts["expected.txt,v"])
	if diff := bytes.Compare([]byte(strings.TrimSpace(string(got))), []byte(want)); diff != 0 {
		t.Fatalf("updated rcs does not match expected")
	}
}

func TestBranchesDefaultSet_InvalidBranch(t *testing.T) {
	if err != nil {
		t.Fatalf("NewRoot error: %v", err)
	}
	err = cmd.Execute([]string{"branches", "default", "set", "-name", "main", "input.txt"})
	if err == nil {
		t.Fatal("expected error for invalid branch name")
	}
	if !strings.Contains(err.Error(), "invalid default branch name") {
		t.Fatalf("unexpected error: %v", err)
	}
}

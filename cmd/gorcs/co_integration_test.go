package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"golang.org/x/tools/txtar"
)

func TestCo_ForceOverwrite(t *testing.T) {
	// Load the test case content from the txtar file
	archivePath := filepath.Join("..", "..", "testdata", "txtar", "operations", "2843-co-force-overwrite-default-subst.txtar")
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

	// 1. Setup: Create RCS file and a MODIFIED working file
	if err := os.WriteFile(workingPath, []byte(parts["input.txt"]), 0644); err != nil {
		t.Fatalf("write working file: %v", err)
	}
	if err := os.WriteFile(rcsPath, []byte(parts["input.txt,v"]), 0644); err != nil {
		t.Fatalf("write rcs file: %v", err)
	}

	// 2. Test: co WITHOUT -f should fail because file is writable
	cmd, err := NewRoot("gorcs", "", "", "")
	if err != nil {
		t.Fatalf("NewRoot error: %v", err)
	}
	// args: -q -l input.txt (from options.conf in txtar, but omitting -f)
	// options.conf has ["-q","-f","-l","input.txt"]
	err = cmd.Execute([]string{"co", "-q", "-l", workingPath})
	if err == nil {
		t.Fatal("expected error when overwriting writable file without -f")
	}
	// Verify error message contains "writable" and "exists"
	if !strings.Contains(err.Error(), "writable") || !strings.Contains(err.Error(), "exists") {
		t.Errorf("unexpected error message: %v", err)
	}

	// 3. Test: co WITH -f should succeed
	cmd, err = NewRoot("gorcs", "", "", "")
	if err != nil {
		t.Fatalf("NewRoot error: %v", err)
	}
	err = cmd.Execute([]string{"co", "-q", "-f", "-l", workingPath})
	if err != nil {
		t.Fatalf("execute co -f failed: %v", err)
	}

	// Verify content matches expected
	got, err := os.ReadFile(workingPath)
	if err != nil {
		t.Fatalf("read working file: %v", err)
	}
	want := strings.TrimSpace(parts["expected.txt"])
	if strings.TrimSpace(string(got)) != want {
		t.Errorf("working file mismatch:\nwant:\n%s\ngot:\n%s", want, string(got))
	}
}

func TestCheckout_Alias(t *testing.T) {
	// Verify `checkout` is an alias for `co`
	tmp := t.TempDir()
	workingPath := filepath.Join(tmp, "input.txt")
	rcsPath := workingPath + ",v"

	// Create dummy RCS file
	rcsContent := `head	1.1;
access;
symbols;
locks;
comment	@# @;


1.1
date	2020.01.01.00.00.00;	author tester;	state Exp;
branches;
next	;


desc
@@


1.1
log
@r1
@
text
@v1
@
`
	if err := os.WriteFile(rcsPath, []byte(rcsContent), 0644); err != nil {
		t.Fatalf("write rcs file: %v", err)
	}

	cmd, err := NewRoot("gorcs", "", "", "")
	if err != nil {
		t.Fatalf("NewRoot error: %v", err)
	}
	// Use `checkout` command
	err = cmd.Execute([]string{"checkout", "-q", workingPath})
	if err != nil {
		t.Fatalf("execute checkout failed: %v", err)
	}

	got, err := os.ReadFile(workingPath)
	if err != nil {
		t.Fatalf("read working file: %v", err)
	}
	if strings.TrimSpace(string(got)) != "v1" {
		t.Errorf("checkout failed to produce expected content")
	}
}

package cli

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
)

func TestFormat_IgnoreTruncation(t *testing.T) {
	// Create a temp file with truncated date
	content := `head	1.1;
access;
symbols;
locks; strict;
comment	@# @;


1.1
date	99.01.01.00.00.00;	author user;	state Exp;
branches;
next	;


desc
@Description
@


1.1
log
@Initial revision
@
text
@Content
@
`
	tmpDir := t.TempDir()
	f, err := os.CreateTemp(tmpDir, "test*.v")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())
	if _, err := f.WriteString(content); err != nil {
		t.Fatal(err)
	}
	f.Close()

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Run Format with ignoreTruncation=true, stdout=true
	// Signature: func Format(output string, force, overwrite, stdout, ignoreTruncation bool, files ...string)
	Format("", false, false, true, true, f.Name())

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	// Check output for 4-digit year (1999)
	if !strings.Contains(output, "date\t1999.01.01.00.00.00;") {
		t.Errorf("Expected 4-digit year in output when ignoreTruncation=true, got:\n%s", output)
	}

	// Run Format with ignoreTruncation=false (default), stdout=true
	r, w, _ = os.Pipe()
	os.Stdout = w
	Format("", false, false, true, false, f.Name())
	w.Close()
	os.Stdout = oldStdout

	buf.Reset()
	io.Copy(&buf, r)
	output = buf.String()

	// Check output for 2-digit year (99)
	if !strings.Contains(output, "date\t99.01.01.00.00.00;") {
		t.Errorf("Expected 2-digit year in output when ignoreTruncation=false, got:\n%s", output)
	}
}

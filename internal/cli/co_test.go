package cli

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestCoPermissions(t *testing.T) {
	// Setup temp dir
	tmpDir := t.TempDir()

	// Create an RCS file with restricted permissions (0600)
	// Note: 0600 is rw-------
	rcsContent := `head	1.1;
access;
symbols;
locks; strict;
comment	@# @;


1.1
date	2024.01.01.00.00.00;	author user;	state Exp;
branches;
next	;


desc
@@


1.1
log
@@
text
@content
@
`
	rcsFile := filepath.Join(tmpDir, "testfile,v")
	if err := os.WriteFile(rcsFile, []byte(rcsContent), 0600); err != nil {
		t.Fatalf("failed to write rcs file: %v", err)
	}
	// Force permissions just in case WriteFile was affected by umask
	if err := os.Chmod(rcsFile, 0600); err != nil {
		t.Fatalf("failed to chmod rcs file: %v", err)
	}

	// Run Co command
	// We check out without locking.
	if err := Co("", false, false, "user", true, "", "", filepath.Join(tmpDir, "testfile")); err != nil {
		t.Fatalf("Co failed: %v", err)
	}

	// Check output file permissions
	outFile := filepath.Join(tmpDir, "testfile")
	fi, err := os.Stat(outFile)
	if err != nil {
		t.Fatalf("stat output file failed: %v", err)
	}

	mode := fi.Mode().Perm()
	t.Logf("Output file mode: %o", mode)

	// Vulnerability check:
	// If source is 0600, output should not be world readable (0644 or 0444).
	if runtime.GOOS != "windows" {
		if mode&0004 != 0 {
			t.Errorf("Security Vulnerability: Output file is world-readable (mode %o), source was 0600", mode)
		}
		if mode&0040 != 0 {
			t.Errorf("Security Vulnerability: Output file is group-readable (mode %o), source was 0600", mode)
		}
	}

	// Functionality check (RCS behavior):
	// Should be read-only if not locked.
	// On Windows, read-only is 0444, writable is 0666. 0200 bit controls write access.
	if mode&0200 != 0 {
		t.Errorf("Output file is writable (mode %o), but was not locked", mode)
	}
}

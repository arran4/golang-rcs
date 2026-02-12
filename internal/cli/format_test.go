package cli

import (
	"bytes"
	_ "embed"
	"io"
	"os"
	"strings"
	"testing"
)

//go:embed testdata/truncated_date.go,v
var truncatedDateTestFile []byte

func TestFormat_KeepTruncatedYears(t *testing.T) {
	tmpDir := t.TempDir()
	f, err := os.CreateTemp(tmpDir, "test*.v")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = os.Remove(f.Name())
	}()
	if _, err := f.Write(truncatedDateTestFile); err != nil {
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Run Format with keepTruncatedYears=true, stdout=true
	// Signature: func Format(output string, force, overwrite, stdout, keepTruncatedYears bool, files ...string)
	Format("", false, false, true, true, f.Name())

	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	os.Stdout = oldStdout

	var buf bytes.Buffer
	if _, err := io.Copy(&buf, r); err != nil {
		t.Fatalf("io.Copy error: %v", err)
	}
	output := buf.String()

	// Check output for 2-digit year (99)
	if !strings.Contains(output, "date\t99.01.01.00.00.00;") {
		t.Errorf("Expected 2-digit year in output when keepTruncatedYears=true, got:\n%s", output)
	}

	// Run Format with keepTruncatedYears=false (default), stdout=true
	r, w, _ = os.Pipe()
	os.Stdout = w
	Format("", false, false, true, false, f.Name())
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	os.Stdout = oldStdout

	buf.Reset()
	if _, err := io.Copy(&buf, r); err != nil {
		t.Fatalf("io.Copy error: %v", err)
	}
	output = buf.String()

	// Check output for 4-digit year (1999) - Default behavior matches backward compatibility (Normalize)
	if !strings.Contains(output, "date\t1999.01.01.00.00.00;") {
		t.Errorf("Expected 4-digit year in output when keepTruncatedYears=false (default), got:\n%s", output)
	}
}

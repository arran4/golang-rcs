package cli

import (
	"bytes"
	_ "embed"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

//go:embed testdata/truncated_date.go,v
var truncatedDateTestFile []byte

func TestFormat_KeepTruncatedYears(t *testing.T) {
	// Prepare stdin (input) and stdout (output capture)
	r := bytes.NewReader(truncatedDateTestFile)
	var buf bytes.Buffer

	// Run Format with keepTruncatedYears=true, stdout=true
	// Signature: func runFormat(stdin io.Reader, stdout io.Writer, output string, force, overwrite, stdout, keepTruncatedYears bool, files ...string)
	runFormat(r, &buf, "", false, false, true, true, "-")

	output := buf.String()

	// Check output for 2-digit year (99)
	if !strings.Contains(output, "date\t99.01.01.00.00.00;") {
		t.Errorf("Expected 2-digit year in output when keepTruncatedYears=true, got:\n%s", output)
	}

	// Prepare stdin (input) and stdout (output capture) for second run
	r = bytes.NewReader(truncatedDateTestFile)
	buf.Reset()

	// Run Format with keepTruncatedYears=false (default), stdout=true
	runFormat(r, &buf, "", false, false, true, false, "-")

	output = buf.String()

	// Check output for 4-digit year (1999) - Default behavior matches backward compatibility (Normalize)
	if !strings.Contains(output, "date\t1999.01.01.00.00.00;") {
		t.Errorf("Expected 4-digit year in output when keepTruncatedYears=false (default), got:\n%s", output)
	}
}

func TestFormat_MultipleFiles_KeepTruncatedYears(t *testing.T) {
	// Create temp files
	tmpDir := t.TempDir()
	file1 := filepath.Join(tmpDir, "file1.rcs")
	file2 := filepath.Join(tmpDir, "file2.rcs")

	if err := os.WriteFile(file1, truncatedDateTestFile, 0644); err != nil {
		t.Fatalf("Failed to write temp file: %v", err)
	}
	if err := os.WriteFile(file2, truncatedDateTestFile, 0644); err != nil {
		t.Fatalf("Failed to write temp file: %v", err)
	}

	var buf bytes.Buffer

	// Run Format with keepTruncatedYears=false (should expand)
	runFormat(nil, &buf, "", false, false, true, false, file1, file2)

	output := buf.String()

	// Check Txtar headers SHOULD NOT exist
	if strings.Contains(output, "-- "+file1+" --") {
		t.Errorf("Did not expect Txtar header for file1")
	}
	if strings.Contains(output, "-- "+file2+" --") {
		t.Errorf("Did not expect Txtar header for file2")
	}

	// Check expansion (should happen for both files)
	// We count occurrences or just check presence.
	if !strings.Contains(output, "date\t1999.01.01.00.00.00;") {
		t.Errorf("Expected expanded year in output, got:\n%s", output)
	}

	// Run Format with keepTruncatedYears=true (should keep)
	buf.Reset()
	runFormat(nil, &buf, "", false, false, true, true, file1, file2)
	output = buf.String()

	if !strings.Contains(output, "date\t99.01.01.00.00.00;") {
		t.Errorf("Expected kept year in output, got:\n%s", output)
	}
}

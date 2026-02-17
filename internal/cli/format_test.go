package cli

import (
	"bytes"
	_ "embed"
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
	// Signature: func runFormat(stdin io.Reader, stdout io.Writer, output string, force, overwrite, stdout, keepTruncatedYears, useMmap bool, files ...string)
	if err := runFormat(r, &buf, "", false, false, true, true, false, "-"); err != nil {
		t.Errorf("runFormat failed: %v", err)
	}

	output := buf.String()

	// Check output for 2-digit year (99)
	if !strings.Contains(output, "date\t99.01.01.00.00.00;") {
		t.Errorf("Expected 2-digit year in output when keepTruncatedYears=true, got:\n%s", output)
	}

	// Prepare stdin (input) and stdout (output capture) for second run
	r = bytes.NewReader(truncatedDateTestFile)
	buf.Reset()

	// Run Format with keepTruncatedYears=false (default), stdout=true
	if err := runFormat(r, &buf, "", false, false, true, false, false, "-"); err != nil {
		t.Errorf("runFormat failed: %v", err)
	}

	output = buf.String()

	// Check output for 4-digit year (1999) - Default behavior matches backward compatibility (Normalize)
	if !strings.Contains(output, "date\t1999.01.01.00.00.00;") {
		t.Errorf("Expected 4-digit year in output when keepTruncatedYears=false (default), got:\n%s", output)
	}
}

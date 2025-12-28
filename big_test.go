package rcs

import (
	"fmt"
	"strings"
	"testing"
	"time"
)

func TestParseFile_BigFile(t *testing.T) {
	// Generate a large RCS file with 1000 revisions
	sb := &strings.Builder{}
	sb.WriteString("head\t1000.1;\n")
	sb.WriteString("access;\n")
	sb.WriteString("symbols;\n")
	sb.WriteString("locks;\n")
	sb.WriteString("comment\t@# @;\n\n\n")

	// Generate revisions 1000.1 down to 1.1
	// Revisions are usually stored in reverse order in RCS?
	// The heads are listed.
	// Revisions in headers block:
	for i := 1000; i >= 1; i-- {
		rev := fmt.Sprintf("%d.1", i)
		next := ""
		if i > 1 {
			next = fmt.Sprintf("%d.1", i-1)
		}
		fmt.Fprintf(sb, "%s\n", rev)
		fmt.Fprintf(sb, "date\t%s;\tauthor user;\tstate Exp;\n", time.Now().Format(DateFormat))
		sb.WriteString("branches;\n")
		fmt.Fprintf(sb, "next\t%s;\n", next)
		sb.WriteString("\n")
	}

	sb.WriteString("\ndesc\n@Big file test\n@\n\n\n")

	// Generate revision contents
	for i := 1000; i >= 1; i-- {
		rev := fmt.Sprintf("%d.1", i)
		fmt.Fprintf(sb, "%s\n", rev)
		sb.WriteString("log\n")
		fmt.Fprintf(sb, "@Log message for %s\n@\n", rev)
		sb.WriteString("text\n")
		if i == 1 {
			fmt.Fprintf(sb, "@Text content for %s\nLine 2\nLine 3\n@\n\n", rev)
		} else {
			fmt.Fprintf(sb, "@Text content for %s\nLine 2\nLine 3\n@\n\n\n", rev)
		}
	}

	s := strings.NewReader(sb.String())
	start := time.Now()
	f, err := ParseFile(s)
	if err != nil {
		t.Fatalf("ParseFile failed: %v", err)
	}
	duration := time.Since(start)
	t.Logf("Parsed 1000 revisions in %v", duration)

	if len(f.RevisionHeads) != 1000 {
		t.Errorf("Expected 1000 revision heads, got %d", len(f.RevisionHeads))
	}
	if len(f.RevisionContents) != 1000 {
		t.Errorf("Expected 1000 revision contents, got %d", len(f.RevisionContents))
	}
	if f.Head != "1000.1" {
		t.Errorf("Expected head 1000.1, got %s", f.Head)
	}
}

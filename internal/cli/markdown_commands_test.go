package cli

import (
	"bytes"
	"strings"
	"testing"

	rcs "github.com/arran4/golang-rcs"
)

func TestMarkdownRoundTrip(t *testing.T) {
	// Sample RCS input
	input := []byte(`head	1.2;
access;
symbols
	tag1:1.1;
locks; strict;
comment	@# @;
expand	@o@;


1.2
date	2023.01.01.00.00.00;	author user;	state Exp;
branches;
next	1.1;

1.1
date	2022.01.01.00.00.00;	author user;	state Exp;
branches;
next	;


desc
@description
multiline
@


1.2
log
@log message 1.2
@
text
@content 1.2
line 2
@


1.1
log
@log message 1.1
@
text
@content 1.1
@
`)

	// 2. Parse original
	parsedOriginal, err := rcs.ParseFile(bytes.NewReader(input))
	if err != nil {
		t.Fatalf("parse original: %v", err)
	}

	// 3. ToMarkdown
	markdownOutput := rcsFileToMarkdown(parsedOriginal)
	t.Logf("Markdown Output:\n%s", markdownOutput)

	// 4. FromMarkdown
	parsedNew, err := parseMarkdownFile(strings.NewReader(markdownOutput))
	if err != nil {
		t.Fatalf("parse markdown: %v", err)
	}

	// 5. Compare
	if parsedOriginal.Head != parsedNew.Head {
		t.Errorf("Head mismatch: got %q, want %q", parsedNew.Head, parsedOriginal.Head)
	}
	// Description in RCS: "description\nmultiline\n" (because @...@)
	// Markdown output:
	// ```text
	// description
	// multiline
	// ```
	// Parsed back: "description\nmultiline\n" (with trailing newline)

	if normalize(parsedOriginal.Description) != normalize(parsedNew.Description) {
		t.Errorf("Description mismatch: got %q, want %q", parsedNew.Description, parsedOriginal.Description)
	}

	if len(parsedOriginal.RevisionHeads) != len(parsedNew.RevisionHeads) {
		t.Fatalf("Revision count mismatch: got %d, want %d", len(parsedNew.RevisionHeads), len(parsedOriginal.RevisionHeads))
	}

	for i, rh := range parsedOriginal.RevisionHeads {
		got := parsedNew.RevisionHeads[i]
		if rh.Revision != got.Revision {
			t.Errorf("Revision %d mismatch: got %q, want %q", i, got.Revision, rh.Revision)
		}
		if rh.Date != got.Date {
			t.Errorf("Date %d mismatch: got %q, want %q", i, got.Date, rh.Date)
		}
		if rh.Author != got.Author {
			t.Errorf("Author %d mismatch: got %q, want %q", i, got.Author, rh.Author)
		}
		if rh.State != got.State {
			t.Errorf("State %d mismatch: got %q, want %q", i, got.State, rh.State)
		}
		if rh.NextRevision != got.NextRevision {
			t.Errorf("Next %d mismatch: got %q, want %q", i, got.NextRevision, rh.NextRevision)
		}
	}

	for i, rc := range parsedOriginal.RevisionContents {
		got := parsedNew.RevisionContents[i]
		if rc.Revision != got.Revision {
			t.Errorf("Content Revision %d mismatch: got %q, want %q", i, got.Revision, rc.Revision)
		}
		if normalize(rc.Log) != normalize(got.Log) {
			t.Errorf("Log %d mismatch: got %q, want %q", i, got.Log, rc.Log)
		}
		if normalize(rc.Text) != normalize(got.Text) {
			t.Errorf("Text %d mismatch: got %q, want %q", i, got.Text, rc.Text)
		}
	}
}

func normalize(s string) string {
	return strings.TrimSpace(strings.ReplaceAll(s, "\r\n", "\n"))
}

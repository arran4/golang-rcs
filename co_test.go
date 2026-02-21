package rcs

import (
	"bytes"
	"embed"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"golang.org/x/tools/txtar"
)

//go:embed testdata/co/co-api-checkout-lock.txtar
var coTests embed.FS

func TestCheckout(t *testing.T) {
	b, err := coTests.ReadFile("testdata/co/co-api-checkout-lock.txtar")
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	archive := txtar.Parse(bytes.ReplaceAll(b, []byte("\r\n"), []byte("\n")))
	parts := map[string]string{}
	for _, f := range archive.Files {
		parts[f.Name] = string(f.Data)
	}

	input, ok := parts["input.txt,v"]
	if !ok {
		t.Fatal("missing input.txt,v")
	}
	expectedContent, ok := parts["expected-content.txt"]
	if !ok {
		t.Fatal("missing expected-content.txt")
	}
	expectedLocked, ok := parts["expected-locked.txt,v"]
	if !ok {
		t.Fatal("missing expected-locked.txt,v")
	}

	f, err := ParseFile(strings.NewReader(input))
	if err != nil {
		t.Fatalf("ParseFile() error = %v", err)
	}

	verdict, err := f.Checkout("tester", WithRevision("1.1"), WithSetLock)
	if err != nil {
		t.Fatalf("Checkout() error = %v", err)
	}
	if verdict.Revision != "1.1" {
		t.Fatalf("Checkout() revision = %q, want 1.1", verdict.Revision)
	}
	if diff := cmp.Diff(strings.TrimSpace(expectedContent), strings.TrimSpace(verdict.Content)); diff != "" {
		t.Fatalf("Checkout() content mismatch (-want +got):\n%s", diff)
	}
	if diff := cmp.Diff(strings.TrimSpace(expectedLocked), strings.TrimSpace(f.String())); diff != "" {
		t.Fatalf("Checkout() lock mutation mismatch (-want +got):\n%s", diff)
	}
}

func TestCheckout_InvalidFlags(t *testing.T) {
	_, err := NewFile().Checkout("tester", WithSetLock, WithClearLock)
	if err == nil {
		t.Fatal("Checkout() error = nil, want error")
	}
}

//go:embed testdata/co/date.rcs
var coDateTestRCS string

func TestCheckout_WithDate(t *testing.T) {
	// Normalize line endings to \n for consistent testing across platforms
	input := strings.ReplaceAll(coDateTestRCS, "\r\n", "\n")
	f, err := ParseFile(strings.NewReader(input))
	if err != nil {
		t.Fatalf("ParseFile: %v", err)
	}

	tests := []struct {
		name        string
		date        time.Time
		wantRev     string
		wantContent string
	}{
		{
			name:    "Before 1.1",
			date:    time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC),
			wantRev: "", // No revision found
		},
		{
			name:        "At 1.1",
			date:        time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC),
			wantRev:     "1.1",
			wantContent: "A\n",
		},
		{
			name:        "Between 1.1 and 1.2",
			date:        time.Date(2022, 6, 1, 0, 0, 0, 0, time.UTC),
			wantRev:     "1.1",
			wantContent: "A\n",
		},
		{
			name:        "At 1.2",
			date:        time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
			wantRev:     "1.2",
			wantContent: "A\nB\n",
		},
		{
			name:        "After 1.2",
			date:        time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			wantRev:     "1.2",
			wantContent: "A\nB\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			verdict, err := f.Checkout("user", WithDate(tt.date))
			if tt.wantRev == "" {
				if err == nil {
					t.Errorf("Checkout(date=%v) expected error, got rev %s", tt.date, verdict.Revision)
				}
			} else {
				if err != nil {
					t.Fatalf("Checkout(date=%v) error: %v", tt.date, err)
				}
				if verdict.Revision != tt.wantRev {
					t.Errorf("Checkout(date=%v) rev = %s, want %s", tt.date, verdict.Revision, tt.wantRev)
				}
				if tt.wantContent != "" && verdict.Content != tt.wantContent {
					t.Errorf("Checkout(date=%v) content = %q, want %q", tt.date, verdict.Content, tt.wantContent)
				}
			}
		})
	}
}

func TestCheckout_WithDatePrefersMostRecentTimestampOverRevisionNumber(t *testing.T) {
	f := &File{
		Head: "1.3",
		RevisionHeads: []*RevisionHead{
			{Revision: "1.3", Date: "2020.01.01.00.00.00", NextRevision: "1.2"},
			{Revision: "1.2", Date: "2022.01.01.00.00.00", NextRevision: "1.1"},
			{Revision: "1.1", Date: "2021.01.01.00.00.00", NextRevision: ""},
		},
		RevisionContents: []*RevisionContent{
			{Revision: "1.3", Text: "HEAD\n"},
			{Revision: "1.2", Text: "d1 1\na1 1\nMIDDLE\n"},
			{Revision: "1.1", Text: "d1 1\na1 1\nOLD\n"},
		},
	}

	verdict, err := f.Checkout("user", WithDate(time.Date(2021, 6, 1, 0, 0, 0, 0, time.UTC)))
	if err != nil {
		t.Fatalf("Checkout() error = %v", err)
	}
	if verdict.Revision != "1.1" {
		t.Fatalf("Checkout() revision = %q, want %q", verdict.Revision, "1.1")
	}
}

func TestCheckout_UnknownOption(t *testing.T) {
	_, err := NewFile().Checkout("tester", struct{}{})
	if err == nil {
		t.Fatal("Checkout() error = nil, want error")
	}
}

package rcs

import (
	"strings"
	"testing"
	"time"
)

// TestGenerateDeltaInterleaved tests cases where prefix/suffix matching
// produces a DIFFERENT (potentially incorrect) delta compared to LCS.
func TestGenerateDeltaInterleaved(t *testing.T) {
	cases := []struct {
		name string
		v1   string // first checkin
		v2   string // second checkin (new head)
	}{
		{
			"interleaved changes",
			"a\nX\nb\nY\nc\n", // old (v1)
			"a\nP\nb\nQ\nc\n", // new (v2)
		},
		{
			"repeated lines",
			"a\na\nb\n", // old
			"a\nb\nb\n", // new
		},
		{
			"match in middle only",
			"X\nCOMMON\nY\n", // old
			"A\nCOMMON\nB\n", // new
		},
		{
			"long prefix different suffix",
			"a\nb\nc\nd\nOLD\n", // old
			"a\nb\nc\nd\nNEW\n", // new
		},
		{
			"long suffix different prefix",
			"OLD\nb\nc\nd\n", // old
			"NEW\nb\nc\nd\n", // new
		},
		{
			"scattered single-line changes",
			"1\nA\n3\nB\n5\n", // old
			"1\nX\n3\nY\n5\n", // new
		},
		{
			"insert in middle",
			"a\nb\nc\n",      // old (3 lines)
			"a\nNEW\nb\nc\n", // new (4 lines, insert after a)
		},
		{
			"delete from middle",
			"a\nDEL\nb\nc\n", // old (4 lines)
			"a\nb\nc\n",      // new (3 lines, DEL removed)
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			f := NewFile()
			d1 := WithDate(time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC))
			d2 := WithDate(time.Date(2020, 1, 2, 0, 0, 0, 0, time.UTC))

			_, err := f.Checkin("user", "v1", tt.v1, WithInitial{}, d1, WithSetLock)
			if err != nil {
				t.Fatalf("checkin v1: %v", err)
			}
			_, err = f.Checkin("user", "v2", tt.v2, d2, WithSetLock)
			if err != nil {
				t.Fatalf("checkin v2: %v", err)
			}

			// Verify round-trip: checkout both revisions
			v2out, err := f.Checkout("user", WithRevision("1.2"))
			if err != nil {
				t.Fatalf("checkout 1.2: %v", err)
			}
			if v2out.Content != tt.v2 {
				t.Errorf("v1.2 mismatch:\n  got:  %q\n  want: %q", v2out.Content, tt.v2)
			}

			v1out, err := f.Checkout("user", WithRevision("1.1"))
			if err != nil {
				t.Fatalf("checkout 1.1: %v", err)
			}
			if v1out.Content != tt.v1 {
				t.Errorf("v1.1 mismatch:\n  got:  %q\n  want: %q\n  delta stored in 1.1 RevisionContent", v1out.Content, tt.v1)
			}

			// Also verify File.String() round-trips through parse
			serialized := f.String()
			reparsed, err := ParseFile(strings.NewReader(serialized))
			if err != nil {
				t.Fatalf("reparse failed: %v\nSerialized:\n%s", err, serialized)
			}
			v1re, err := reparsed.Checkout("user", WithRevision("1.1"))
			if err != nil {
				t.Fatalf("checkout 1.1 from reparsed: %v", err)
			}
			if v1re.Content != tt.v1 {
				t.Errorf("v1.1 after reparse mismatch:\n  got:  %q\n  want: %q", v1re.Content, tt.v1)
			}
		})
	}
}

package rcs

import (
	"testing"
	"time"
)

func TestGenerateDeltaUnit(t *testing.T) {
	tests := []struct {
		name     string
		newLines []string
		oldLines []string
		want     string
	}{
		{"identical", []string{"a", "b"}, []string{"a", "b"}, ""},
		{"empty both", nil, nil, ""},
		{"empty new", nil, []string{"a"}, "a0 1\na\n"},
		{"empty old", []string{"a"}, nil, "d1 1\n"},
		{"pure append", []string{"a", "b", "c"}, []string{"a"}, "d2 2\n"},
		{"pure prepend", []string{"c"}, []string{"a", "b", "c"}, "a0 2\na\nb\n"},
		{"delete from start", []string{"a", "b", "c"}, []string{"b", "c"}, "d1 1\n"},
		{"single line replace", []string{"new"}, []string{"old"}, "d1 1\na1 1\nold\n"},
		{"full replace", []string{"A", "B", "C"}, []string{"X", "Y", "Z"}, "d1 3\na3 3\nX\nY\nZ\n"},
		{"middle change", []string{"a", "NEW", "c"}, []string{"a", "OLD", "c"}, "d2 1\na2 1\nOLD\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := generateDelta(tt.newLines, tt.oldLines)
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestCheckinRoundTrip(t *testing.T) {
	cases := []struct {
		name string
		v1   string
		v2   string
	}{
		{"simple replace", "Hello, world!\n", "Hello, updated world!\n"},
		{"append lines", "A\nB\n", "A\nB\nC\nD\n"},
		{"delete lines", "A\nB\nC\n", "A\n"},
		{"full replace", "ONE\nTWO\nTHREE\n", "ALPHA\nBETA\n"},
		{"single line", "old\n", "new\n"},
		{"multiline to single", "A\nB\nC\n", "X\n"},
		{"single to multiline", "X\n", "A\nB\nC\n"},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			f := NewFile()
			d := WithDate(time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC))

			// Initial checkin
			_, err := f.Checkin("user", "v1", tt.v1, WithInitial{}, d, WithSetLock)
			if err != nil {
				t.Fatalf("checkin v1: %v", err)
			}

			// Second checkin
			d2 := WithDate(time.Date(2020, 1, 2, 0, 0, 0, 0, time.UTC))
			_, err = f.Checkin("user", "v2", tt.v2, d2, WithSetLock)
			if err != nil {
				t.Fatalf("checkin v2: %v", err)
			}

			// Checkout v1.2 (head)
			v2out, err := f.Checkout("user", WithRevision("1.2"))
			if err != nil {
				t.Fatalf("checkout 1.2: %v", err)
			}
			if v2out.Content != tt.v2 {
				t.Errorf("v1.2: got %q, want %q", v2out.Content, tt.v2)
			}

			// Checkout v1.1 (reconstructed via delta)
			v1out, err := f.Checkout("user", WithRevision("1.1"))
			if err != nil {
				t.Fatalf("checkout 1.1: %v", err)
			}
			if v1out.Content != tt.v1 {
				t.Errorf("v1.1: got %q, want %q", v1out.Content, tt.v1)
			}
		})
	}
}

func TestCheckinForceIdentical(t *testing.T) {
	f := NewFile()
	d := WithDate(time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC))
	_, err := f.Checkin("user", "v1", "same\n", WithInitial{}, d, WithSetLock)
	if err != nil {
		t.Fatal(err)
	}

	// Without force — should fail
	d2 := WithDate(time.Date(2020, 1, 2, 0, 0, 0, 0, time.UTC))
	_, err = f.Checkin("user", "v2", "same\n", d2)
	if err == nil {
		t.Fatal("expected error for identical content without force")
	}

	// With force — should succeed
	v, err := f.Checkin("user", "v2", "same\n", d2, WithForce{}, WithSetLock)
	if err != nil {
		t.Fatalf("force checkin failed: %v", err)
	}
	if v.Revision != "1.2" {
		t.Errorf("revision = %q, want 1.2", v.Revision)
	}
}

func TestCheckinNilFile(t *testing.T) {
	var f *File
	_, err := f.Checkin("user", "msg", "text\n")
	if err == nil {
		t.Fatal("expected error for nil file")
	}
}

func TestCheckinEmptyUser(t *testing.T) {
	f := NewFile()
	_, err := f.Checkin("", "msg", "text\n", WithInitial{})
	if err == nil {
		t.Fatal("expected error for empty user")
	}
}

func TestIncrementNum(t *testing.T) {
	tests := []struct {
		input string
		want  string
		err   bool
	}{
		{"1.1", "1.2", false},
		{"1.99", "1.100", false},
		{"1.1.1.3", "1.1.1.4", false},
		{"1", "", true},
		{"", "", true},
		{"1.abc", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := incrementNum(Num(tt.input))
			if tt.err && err == nil {
				t.Fatal("expected error")
			}
			if !tt.err && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

// TestNewlineNormalizationRoundTrip verifies that checkin normalizes
// non-empty text to end with \n, and that checkout of older revisions
// reconstructs correctly for all newline-state transitions.
func TestNewlineNormalizationRoundTrip(t *testing.T) {
	cases := []struct {
		name   string
		v1, v2 string
		wantV1 string // after normalization
		wantV2 string
	}{
		{"empty_to_content", "", "A\n", "", "A\n"},
		{"content_to_empty", "A\n", "", "A\n", ""},
		{"no_newline_v1", "old", "new\n", "old\n", "new\n"},
		{"no_newline_v2", "old\n", "new", "old\n", "new\n"},
		{"no_newline_both", "old", "new", "old\n", "new\n"},
		{"newline_only_v2", "", "\n", "", "\n"},
		{"normal_both", "first\n", "second\n", "first\n", "second\n"},
		{"multiline", "A\nB\n", "X\nY\nZ\n", "A\nB\n", "X\nY\nZ\n"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			f := NewFile()
			d := WithDate(time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC))

			_, err := f.Checkin("u", "v1", tc.v1, WithInitial{}, d, WithSetLock)
			if err != nil {
				t.Fatalf("v1 checkin: %v", err)
			}

			d2 := WithDate(time.Date(2020, 1, 2, 0, 0, 0, 0, time.UTC))
			_, err = f.Checkin("u", "v2", tc.v2, d2, WithSetLock)
			if err != nil {
				t.Fatalf("v2 checkin: %v", err)
			}

			// Verify HEAD (v2)
			v2out, err := f.Checkout("u", WithRevision("1.2"))
			if err != nil {
				t.Fatalf("checkout 1.2: %v", err)
			}
			if v2out.Content != tc.wantV2 {
				t.Errorf("1.2: got %q, want %q", v2out.Content, tc.wantV2)
			}

			// Verify old revision (v1, reconstructed via delta)
			v1out, err := f.Checkout("u", WithRevision("1.1"))
			if err != nil {
				t.Fatalf("checkout 1.1: %v", err)
			}
			if v1out.Content != tc.wantV1 {
				t.Errorf("1.1: got %q, want %q", v1out.Content, tc.wantV1)
			}
		})
	}
}

// TestWithInitialOnExistingFile verifies that WithInitial is rejected
// when the file already has revisions.
func TestWithInitialOnExistingFile(t *testing.T) {
	f := NewFile()
	d := WithDate(time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC))
	_, err := f.Checkin("user", "first", "content\n", WithInitial{}, d, WithSetLock)
	if err != nil {
		t.Fatalf("initial checkin: %v", err)
	}

	_, err = f.Checkin("user", "second", "new\n", WithInitial{})
	if err == nil {
		t.Fatal("expected error for WithInitial on file with existing revisions")
	}
}

// TestDuplicateRevisionRejected verifies that an explicit revision
// matching an existing one is rejected.
func TestDuplicateRevisionRejected(t *testing.T) {
	f := NewFile()
	d := WithDate(time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC))
	_, err := f.Checkin("user", "first", "content\n", WithInitial{}, d, WithSetLock)
	if err != nil {
		t.Fatalf("initial checkin: %v", err)
	}

	d2 := WithDate(time.Date(2020, 1, 2, 0, 0, 0, 0, time.UTC))
	_, err = f.Checkin("user", "dup", "different\n", d2, WithRevision("1.1"))
	if err == nil {
		t.Fatal("expected error for duplicate revision 1.1")
	}
}

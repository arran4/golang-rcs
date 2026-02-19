package rcs

import (
	"testing"
)

func TestParseDeltaStats(t *testing.T) {
	tests := []struct {
		name        string
		delta       string
		wantDeleted int
		wantAdded   int
	}{
		{
			name:        "empty",
			delta:       "",
			wantDeleted: 0,
			wantAdded:   0,
		},
		{
			name:        "simple add",
			delta:       "a1 1\nadded line\n",
			wantDeleted: 0,
			wantAdded:   1,
		},
		{
			name:        "simple delete",
			delta:       "d1 1\n",
			wantDeleted: 1,
			wantAdded:   0,
		},
		{
			name:        "mixed",
			delta:       "d1 2\na3 1\nnew line\n",
			wantDeleted: 2,
			wantAdded:   1,
		},
		{
			name:        "multiple adds",
			delta:       "a1 2\nline1\nline2\na5 1\nline3\n",
			wantDeleted: 0,
			wantAdded:   3,
		},
		{
			name:        "invalid format ignored",
			delta:       "x1 1\n",
			wantDeleted: 0,
			wantAdded:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotDeleted, gotAdded := parseDeltaStats(tt.delta)
			if gotDeleted != tt.wantDeleted {
				t.Errorf("parseDeltaStats() deleted = %v, want %v", gotDeleted, tt.wantDeleted)
			}
			if gotAdded != tt.wantAdded {
				t.Errorf("parseDeltaStats() added = %v, want %v", gotAdded, tt.wantAdded)
			}
		})
	}
}

func TestGetLinesStats(t *testing.T) {
	// Mock file structure
	f := &File{
		RevisionContents: []*RevisionContent{
			{Revision: "1.2", Text: "full text\n"},
			{Revision: "1.1", Text: "d1 1\n"}, // Delta for 1.1 (reverse from 1.2)
			{Revision: "1.2.1.1", Text: "a1 1\nnew line\n"}, // Delta for 1.2.1.1 (forward from 1.2)
		},
	}

	tests := []struct {
		name     string
		rh       *RevisionHead
		want     string
		wantErr  bool
	}{
		{
			name: "trunk revision (reverse delta)",
			rh: &RevisionHead{
				Revision:     "1.2",
				NextRevision: "1.1",
			},
			// 1.2 is head. 1.1 is next.
			// 1.1 contains delta "d1 1".
			// Reverse logic: lines added in 1.2 (deleted in 1.1) are dCount.
			// dCount from "d1 1" is 1.
			// Result: +1 -0.
			want: "  lines: +1 -0",
		},
		{
			name: "branch revision (forward delta)",
			rh: &RevisionHead{
				Revision: "1.2.1.1",
			},
			// 1.2.1.1 has delta "a1 1". Forward delta.
			// 1.2.1.1 is newer than 1.2.
			// "a1 1" means add 1 line to 1.2 to get 1.2.1.1.
			// Logic in rlog.go for forward: aCount -> added (+), dCount -> removed (-).
			// aCount=1, dCount=0.
			// Result: +1 -0.
			want: "  lines: +1 -0",
		},
		{
			name: "trunk tip (no next)",
			rh: &RevisionHead{
				Revision:     "1.3",
				NextRevision: "",
			},
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getLinesStats(f, tt.rh)
			if (err != nil) != tt.wantErr {
				t.Errorf("getLinesStats() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("getLinesStats() = %q, want %q", got, tt.want)
			}
		})
	}
}

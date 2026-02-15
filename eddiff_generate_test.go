package rcs_test

import (
	"reflect"
	"testing"

	"github.com/arran4/golang-rcs"
	rcstesting "github.com/arran4/golang-rcs/internal/testing"
	_ "github.com/arran4/golang-rcs/diff/lcs"
)

func TestGenerateEdDiffFromLines(t *testing.T) {
	tests := []struct {
		name string
		from []string
		to   []string
		want rcs.EdDiff
	}{
		{
			name: "Simple Add",
			from: []string{"A"},
			to:   []string{"A", "B"},
			want: rcs.EdDiff{
				rcs.Add{LineStart: 1, Lines: []string{"B"}},
			},
		},
		{
			name: "Simple Delete",
			from: []string{"A", "B"},
			to:   []string{"A"},
			want: rcs.EdDiff{
				rcs.Delete{2, 1},
			},
		},
		{
			name: "Modify (Delete then Add)",
			from: []string{"A"},
			to:   []string{"B"},
			// Logic: Delete A (d1 1), Add B (a0 1).
			// My implementation prefers Delete (move i) on ties.
			// Backtrack order: Delete A, Add B.
			// Reverse order (forward): Add B, Delete A.
			// Forward process:
			// Add B -> a0 1.
			// Delete A -> d1 1.
			want: rcs.EdDiff{
				rcs.Add{LineStart: 0, Lines: []string{"B"}},
				rcs.Delete{1, 1},
			},
		},
		{
			name: "Multiple disjoint edits",
			from: []string{"A", "B", "C"},
			to:   []string{"A", "X", "C"},
			// Match A.
			// Delete B (d2 1).
			// Add X (a1 1).
			// Match C.
			// Tie on B vs X. Prefer Delete B.
			// Ops: Delete B, Add X.
			// Reverse: Add X, Delete B.
			// Add X -> a1 1.
			// Delete B -> d2 1.
			want: rcs.EdDiff{
				rcs.Add{LineStart: 1, Lines: []string{"X"}},
				rcs.Delete{2, 1},
			},
		},
		{
			name: "Group Adds",
			from: []string{"A"},
			to:   []string{"A", "B", "C"},
			want: rcs.EdDiff{
				rcs.Add{LineStart: 1, Lines: []string{"B", "C"}},
			},
		},
		{
			name: "Group Deletes",
			from: []string{"A", "B", "C"},
			to:   []string{"A"},
			want: rcs.EdDiff{
				rcs.Delete{2, 2},
			},
		},
		{
			name: "Empty From",
			from: []string{},
			to:   []string{"A", "B"},
			want: rcs.EdDiff{
				rcs.Add{LineStart: 0, Lines: []string{"A", "B"}},
			},
		},
		{
			name: "Empty To",
			from: []string{"A", "B"},
			to:   []string{},
			want: rcs.EdDiff{
				rcs.Delete{1, 2},
			},
		},
		{
			name: "Identical",
			from: []string{"A", "B"},
			to:   []string{"A", "B"},
			want: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := rcs.GenerateEdDiffFromLines(tt.from, tt.to)
			if err != nil {
				t.Fatalf("GenerateEdDiffFromLines() error = %v", err)
			}

			// Helper to check equality of EdDiff
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GenerateEdDiffFromLines() mismatch.\nGot: %v\nWant: %v", got, tt.want)
			}

			// Verify Round Trip by applying the diff
			r := rcstesting.NewStringLineReader(tt.from)
			w := &rcstesting.StringLineWriter{}

			if err := got.Apply(r, w); err != nil {
				t.Errorf("Apply() error = %v", err)
			}

			gotLines := w.Lines()
			if len(gotLines) != len(tt.to) {
				// Handle nil/empty slice difference for DeepEqual or explicit check
				if len(gotLines) == 0 && len(tt.to) == 0 {
					// OK
				} else {
					t.Errorf("Apply result length = %d, want %d. Got: %v, Want: %v", len(gotLines), len(tt.to), gotLines, tt.to)
				}
			} else {
				for i := range gotLines {
					if gotLines[i] != tt.to[i] {
						t.Errorf("Apply result line %d = %q, want %q", i, gotLines[i], tt.to[i])
					}
				}
			}
		})
	}
}

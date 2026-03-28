package znkr_diff

import (
	rcstesting "github.com/arran4/golang-rcs/internal/testing"
	"testing"
)

func TestGenerateEdDiffFromLines(t *testing.T) {
	tests := []struct {
		name string
		from []string
		to   []string
	}{
		{
			name: "Simple Add",
			from: []string{"A"},
			to:   []string{"A", "B"},
		},
		{
			name: "Simple Delete",
			from: []string{"A", "B"},
			to:   []string{"A"},
		},
		{
			name: "Modify (Delete then Add)",
			from: []string{"A"},
			to:   []string{"B"},
		},
		{
			name: "Multiple disjoint edits",
			from: []string{"A", "B", "C"},
			to:   []string{"A", "X", "C"},
		},
		{
			name: "Group Adds",
			from: []string{"A"},
			to:   []string{"A", "B", "C"},
		},
		{
			name: "Group Deletes",
			from: []string{"A", "B", "C"},
			to:   []string{"A"},
		},
		{
			name: "Empty From",
			from: []string{},
			to:   []string{"A", "B"},
		},
		{
			name: "Empty To",
			from: []string{"A", "B"},
			to:   []string{},
		},
		{
			name: "Identical",
			from: []string{"A", "B"},
			to:   []string{"A", "B"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GenerateEdDiffFromLines(tt.from, tt.to)
			if err != nil {
				t.Fatalf("Generate() error = %v", err)
			}

			// Verify Round Trip by applying the diff
			r := rcstesting.NewStringLineReader(tt.from)
			w := &rcstesting.StringLineWriter{}

			if err := got.Apply(r, w); err != nil {
				t.Errorf("Apply() error = %v", err)
			}

			gotLines := w.Lines()
			if len(gotLines) != len(tt.to) {
				if len(gotLines) == 0 && len(tt.to) == 0 {
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

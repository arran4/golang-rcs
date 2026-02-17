package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	rcs "github.com/arran4/golang-rcs"
)

func loadTestInput(t *testing.T) []byte {
	t.Helper()
	// Adjusted path: testdata is in the root, internal/cli is in internal. So ../.. points to root.
	path := filepath.Join("..", "..", "testdata", "testinput.go,v")
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read test input from %s: %v", path, err)
	}
	return b
}

func writeSubset(t *testing.T, dir string, indices []int) string {
	t.Helper()
	input := loadTestInput(t)
	f, err := rcs.ParseFile(bytes.NewReader(input))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	var heads []*rcs.RevisionHead
	var contents []*rcs.RevisionContent

	for _, i := range indices {
		if i >= len(f.RevisionHeads) {
			t.Fatalf("index %d out of range", i)
		}
		heads = append(heads, f.RevisionHeads[i])
		contents = append(contents, f.RevisionContents[i])
	}

	// Fix NextRevision pointers
	for i := 0; i < len(heads)-1; i++ {
		heads[i].NextRevision = heads[i+1].Revision
	}
	if len(heads) > 0 {
		heads[len(heads)-1].NextRevision = ""
		f.Head = heads[0].Revision.String()
	}
	f.RevisionHeads = heads
	f.RevisionContents = contents

	p := filepath.Join(dir, "subset,v")
	// Using f.String() to serialize
	if err := os.WriteFile(p, []byte(f.String()), 0644); err != nil {
		t.Fatalf("write file: %v", err)
	}
	return p
}

func writeFull(t *testing.T, dir string) string {
	t.Helper()
	p := filepath.Join(dir, "full,v")
	input := loadTestInput(t)
	if err := os.WriteFile(p, input, 0644); err != nil {
		t.Fatalf("write file: %v", err)
	}
	return p
}

func TestPadCommits(t *testing.T) {
	dir := t.TempDir()

	// First test case: without padding
	subset := writeSubset(t, dir, []int{0, 1, 2})
	full := writeFull(t, dir)

	if err := NormalizeRevisions(false, false, subset, full); err != nil {
		t.Errorf("NormalizeRevisions failed: %v", err)
	}

	f, err := os.ReadFile(subset)
	if err != nil {
		t.Fatalf("read file: %v", err)
	}
	parsed, err := rcs.ParseFile(bytes.NewReader(f))
	if err != nil {
		t.Fatalf("parse result: %v", err)
	}
	// Original test expected 3 heads
	if got, want := len(parsed.RevisionHeads), 3; got != want {
		t.Fatalf("without padding got %d revs want %d", got, want)
	}

	// Second test case: with padding
	// Re-create subset because it was modified in place by NormalizeRevisions
	subset = writeSubset(t, dir, []int{0, 1, 2})
	// full is also modified? The command modifies files in place.
	// So we should re-create full as well to be safe.
	full = writeFull(t, dir)

	if err := NormalizeRevisions(true, false, subset, full); err != nil {
		t.Errorf("NormalizeRevisions failed: %v", err)
	}

	f, err = os.ReadFile(subset)
	if err != nil {
		t.Fatalf("read file: %v", err)
	}
	parsed, err = rcs.ParseFile(bytes.NewReader(f))
	if err != nil {
		t.Fatalf("parse result: %v", err)
	}
	// Original test expected 6 heads (because full has more revisions and padding adds them)
	if got, want := len(parsed.RevisionHeads), 6; got != want {
		t.Fatalf("with padding got %d revs want %d", got, want)
	}
}

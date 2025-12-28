package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	rcs "github.com/arran4/golang-rcs"
)

func loadTestInput(t *testing.T) []byte {
	t.Helper()
	b, err := os.ReadFile(filepath.Join("..", "..", "testdata", "testinput.go,v"))
	if err != nil {
		t.Fatalf("read test input: %v", err)
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
		heads = append(heads, f.RevisionHeads[i])
		contents = append(contents, f.RevisionContents[i])
	}
	for i := 0; i < len(heads)-1; i++ {
		heads[i].NextRevision = heads[i+1].Revision
	}
	if len(heads) > 0 {
		heads[len(heads)-1].NextRevision = ""
		f.Head = heads[0].Revision
	}
	f.RevisionHeads = heads
	f.RevisionContents = contents
	p := filepath.Join(dir, "subset,v")
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

func runNormalize(args ...string) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()
	os.Args = append([]string{"normalize-revisions"}, args...)
	main()
}

func TestPadCommits(t *testing.T) {
	dir := t.TempDir()
	subset := writeSubset(t, dir, []int{0, 1, 2})
	full := writeFull(t, dir)
	runNormalize(subset, full)
	f, err := os.ReadFile(subset)
	if err != nil {
		t.Fatalf("read file: %v", err)
	}
	parsed, err := rcs.ParseFile(bytes.NewReader(f))
	if err != nil {
		t.Fatalf("parse result: %v", err)
	}
	if got, want := len(parsed.RevisionHeads), 3; got != want {
		t.Fatalf("without padding got %d revs want %d", got, want)
	}

	subset = writeSubset(t, dir, []int{0, 1, 2})
	full = writeFull(t, dir)
	runNormalize("-pad-commits", subset, full)

	f, err = os.ReadFile(subset)
	if err != nil {
		t.Fatalf("read file: %v", err)
	}
	parsed, err = rcs.ParseFile(bytes.NewReader(f))
	if err != nil {
		t.Fatalf("parse result: %v", err)
	}
	if got, want := len(parsed.RevisionHeads), 6; got != want {
		t.Fatalf("with padding got %d revs want %d", got, want)
	}
}

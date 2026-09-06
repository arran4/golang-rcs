package rcs

import "testing"

// applyDelta must return an empty string (not "\n") when the edit script
// removes every line, and must otherwise terminate output with a newline.
func TestApplyDeltaEmptyOutput(t *testing.T) {
	// Deleting the only line yields zero output lines -> empty string.
	got, err := applyDelta("only line\n", "d1 1\n")
	if err != nil {
		t.Fatalf("applyDelta: %v", err)
	}
	if got != "" {
		t.Errorf("empty result = %q, want \"\"", got)
	}
}

func TestApplyDeltaPartialDeleteKeepsNewline(t *testing.T) {
	// Deleting the second of two lines leaves "a\n".
	got, err := applyDelta("a\nb\n", "d2 1\n")
	if err != nil {
		t.Fatalf("applyDelta: %v", err)
	}
	if got != "a\n" {
		t.Errorf("result = %q, want %q", got, "a\n")
	}
}

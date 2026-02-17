package cli

import "testing"

func TestNormalizeDefaultBranch(t *testing.T) {
	t.Run("revision to branch", func(t *testing.T) {
		got, err := normalizeDefaultBranch("1.1.1.1")
		if err != nil {
			t.Fatalf("normalizeDefaultBranch returned error: %v", err)
		}
		if got != "1.1.1" {
			t.Fatalf("normalizeDefaultBranch = %q, want %q", got, "1.1.1")
		}
	})

	t.Run("invalid", func(t *testing.T) {
		if _, err := normalizeDefaultBranch("main"); err == nil {
			t.Fatal("expected error for invalid branch name, got nil")
		}
	})
}

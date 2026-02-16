package diff

import (
	"testing"
)

func TestSetDefaultAlgorithm(t *testing.T) {
	// Setup: verify LCS is registered
	if _, err := GetAlgorithm("lcs"); err != nil {
		t.Fatalf("lcs algorithm should be registered: %v", err)
	}

	// Test invalid algorithm
	err := SetDefaultAlgorithm("invalid-algo")
	if err == nil {
		t.Errorf("SetDefaultAlgorithm(invalid) should fail")
	}

	// Test valid algorithm
	err = SetDefaultAlgorithm("lcs")
	if err != nil {
		t.Errorf("SetDefaultAlgorithm(lcs) should succeed: %v", err)
	}

	// Test Generate uses the set algorithm
	// Register a mock algorithm
	called := false
	mockAlgo := func(from, to []string) (EdDiff, error) {
		called = true
		return EdDiff{}, nil
	}
	Register("mock", mockAlgo)

	err = SetDefaultAlgorithm("mock")
	if err != nil {
		t.Fatalf("SetDefaultAlgorithm(mock) failed: %v", err)
	}

	_, _ = Generate([]string{"a"}, []string{"b"})

	if !called {
		t.Errorf("Generate did not call the set default algorithm")
	}

	// Cleanup: set back to lcs (or empty/default) to avoid affecting other tests if they run in same process
	// Note: defaultAlgo is package level variable.
	// Since lcs is what Generate falls back to, setting it to lcs is safe.
	if err := SetDefaultAlgorithm("lcs"); err != nil {
		t.Errorf("SetDefaultAlgorithm(lcs) failed during cleanup: %v", err)
	}
}

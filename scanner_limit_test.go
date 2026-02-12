package rcs

import (
	"strings"
	"testing"
)

func TestScannerLargeTokenDefault(t *testing.T) {
	// 5MB token - checks if we can scan reasonably large tokens (larger than default 64KB bufio limit)
	size := 5 * 1024 * 1024
	largeToken := strings.Repeat("a", size)
	input := "@" + largeToken + "@"

	s := NewScanner(strings.NewReader(input))
	val, err := ParseAtQuotedString(s)
	if err != nil {
		t.Fatalf("Failed to scan %d MB token with default scanner: %v", size/1024/1024, err)
	}
	if len(val) != size {
		t.Errorf("Expected token length %d, got %d", size, len(val))
	}
}

func TestScannerEnforceLimit(t *testing.T) {
	// 5MB token
	size := 5 * 1024 * 1024
	largeToken := strings.Repeat("a", size)
	input := "@" + largeToken + "@"

	// Set limit to 1MB. This should fail because token is 5MB.
	s := NewScanner(strings.NewReader(input), MaxBuffer(1024*1024))
	_, err := ParseAtQuotedString(s)
	if err == nil {
		t.Fatalf("Expected error when scanning token larger than limit, got nil")
	}
	if !strings.Contains(err.Error(), "token too long") {
		t.Logf("Got error: %v", err)
	}
}

func TestScannerDefaultLimitEnforced(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping large allocation test in short mode")
	}
	// 65MB token, slightly larger than 64MB default limit
	size := 65 * 1024 * 1024
	largeToken := strings.Repeat("a", size)
	input := "@" + largeToken + "@"

	s := NewScanner(strings.NewReader(input))
	_, err := ParseAtQuotedString(s)
	if err == nil {
		t.Fatalf("Expected error when scanning token larger than default limit (64MB), got nil. Buffer limit is not enforced.")
	}
	if !strings.Contains(err.Error(), "token too long") {
		t.Logf("Got error: %v", err)
	}
}

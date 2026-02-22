package rcs

import (
	"bytes"
	"embed"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"golang.org/x/tools/txtar"
)

//go:embed testdata/co/co-api-checkout-lock.txtar
var coTests embed.FS

func TestCheckout(t *testing.T) {
	b, err := coTests.ReadFile("testdata/co/co-api-checkout-lock.txtar")
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	archive := txtar.Parse(bytes.ReplaceAll(b, []byte("\r\n"), []byte("\n")))
	parts := map[string]string{}
	for _, f := range archive.Files {
		parts[f.Name] = string(f.Data)
	}

	input, ok := parts["input.txt,v"]
	if !ok {
		t.Fatal("missing input.txt,v")
	}
	expectedContent, ok := parts["expected-content.txt"]
	if !ok {
		t.Fatal("missing expected-content.txt")
	}
	expectedLocked, ok := parts["expected-locked.txt,v"]
	if !ok {
		t.Fatal("missing expected-locked.txt,v")
	}

	f, err := ParseFile(strings.NewReader(input))
	if err != nil {
		t.Fatalf("ParseFile() error = %v", err)
	}

	verdict, err := f.Checkout("tester", WithRevision("1.1"), WithSetLock)
	if err != nil {
		t.Fatalf("Checkout() error = %v", err)
	}
	if verdict.Revision != "1.1" {
		t.Fatalf("Checkout() revision = %q, want 1.1", verdict.Revision)
	}
	if diff := cmp.Diff(strings.TrimSpace(expectedContent), strings.TrimSpace(verdict.Content)); diff != "" {
		t.Fatalf("Checkout() content mismatch (-want +got):\n%s", diff)
	}
	if diff := cmp.Diff(strings.TrimSpace(expectedLocked), strings.TrimSpace(f.String())); diff != "" {
		t.Fatalf("Checkout() lock mutation mismatch (-want +got):\n%s", diff)
	}
}

func TestCheckout_InvalidFlags(t *testing.T) {
	_, err := NewFile().Checkout("tester", WithSetLock, WithClearLock)
	if err == nil {
		t.Fatal("Checkout() error = nil, want error")
	}
}

func TestCheckout_UnknownOption(t *testing.T) {
	_, err := NewFile().Checkout("tester", struct{}{})
	if err == nil {
		t.Fatal("Checkout() error = nil, want error")
	}
}

package rcs

import (
	"bytes"
	"errors"
	"io"
	"testing"
)

// mockReadCloser is a simple io.ReadCloser for testing.
type mockReadCloser struct {
	io.Reader
}

func (m *mockReadCloser) Close() error {
	return nil
}

func TestFileContent_Get_Success(t *testing.T) {
	expectedBytes := []byte("hello world")
	calls := 0

	fc := &FileContent{
		loader: func() (io.ReadCloser, error) {
			calls++
			return &mockReadCloser{bytes.NewReader(expectedBytes)}, nil
		},
	}

	// First call should execute loader
	b, err := fc.Get()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !bytes.Equal(b, expectedBytes) {
		t.Errorf("expected %q, got %q", expectedBytes, b)
	}
	if calls != 1 {
		t.Errorf("expected 1 loader call, got %d", calls)
	}

	// Second call should return cached result (weak pointer is likely still valid in same GC cycle)
	b2, err := fc.Get()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !bytes.Equal(b2, expectedBytes) {
		t.Errorf("expected %q, got %q", expectedBytes, b2)
	}
	if calls != 1 {
		t.Errorf("expected still 1 loader call, got %d", calls)
	}
}

func TestFileContent_Get_LoaderError(t *testing.T) {
	expectedErr := errors.New("loader failed")
	fc := &FileContent{
		loader: func() (io.ReadCloser, error) {
			return nil, expectedErr
		},
	}

	b, err := fc.Get()
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err != expectedErr {
		t.Errorf("expected %v, got %v", expectedErr, err)
	}
	if b != nil {
		t.Errorf("expected nil bytes, got %v", b)
	}
}

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

func TestFileContent_Segment(t *testing.T) {
	expectedBytes := []byte("hello world")
	fc := &FileContent{
		loader: func() (io.ReadCloser, error) {
			return &mockReadCloser{bytes.NewReader(expectedBytes)}, nil
		},
	}

	tests := []struct {
		name     string
		offset   int64
		length   int64
		expected []byte
	}{
		{"full segment", 0, 11, []byte("hello world")},
		{"partial segment", 0, 5, []byte("hello")},
		{"middle segment", 6, 5, []byte("world")},
		{"out of bounds length", 6, 10, []byte("world")},
		{"out of bounds offset", 20, 5, []byte{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fco := fc.Segment(tt.offset, tt.length)
			b, err := fco.Get()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !bytes.Equal(b, tt.expected) {
				t.Errorf("expected %q, got %q", tt.expected, b)
			}

			// Test caching of the segment
			b2, err := fco.Get()
			if err != nil {
				t.Fatalf("unexpected error on second get: %v", err)
			}
			if !bytes.Equal(b2, tt.expected) {
				t.Errorf("expected %q on second get, got %q", tt.expected, b2)
			}
		})
	}
}

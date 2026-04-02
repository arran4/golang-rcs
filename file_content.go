package rcs

import (
	"io"
	"sync"
	"weak"
)

type FileContent struct {
	weakPtr weak.Pointer[[]byte]
	loader  func() (io.ReadCloser, error)
	mu      sync.Mutex
}

func (f *FileContent) Get() ([]byte, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	if p := f.weakPtr.Value(); p != nil {
		return *p, nil
	}

	rc, err := f.loader()
	if err != nil {
		return nil, err
	}
	defer func() { _ = rc.Close() }()

	b, err := io.ReadAll(rc)
	if err != nil {
		return nil, err
	}

	f.weakPtr = weak.Make(&b)
	return b, nil
}

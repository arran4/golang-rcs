package rcs

import (
	"bytes"
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

	bp := &b
	f.weakPtr = weak.Make(bp)
	return *bp, nil
}

// Segment creates a new FileContent that represents a segment of the current FileContent.
// This allows for memory-efficient lazy loading of file segments, as the child FileContent
// will allocate its own smaller slice, allowing the larger parent slice to be garbage collected.
func (f *FileContent) Segment(offset, length int64) *FileContent {
	return &FileContent{
		loader: func() (io.ReadCloser, error) {
			b, err := f.Get()
			if err != nil {
				return nil, err
			}
			if offset >= int64(len(b)) {
				return io.NopCloser(bytes.NewReader(nil)), nil
			}
			end := offset + length
			if end > int64(len(b)) {
				end = int64(len(b))
			}
			// When the child FileContent reads from this reader, it will allocate a new []byte.
			// After the read is complete, the parent's full []byte `b` will no longer be referenced
			// strongly by this segment loader, allowing the GC to reclaim the memory if there are no other strong references.
			return io.NopCloser(bytes.NewReader(b[offset:end])), nil
		},
	}
}

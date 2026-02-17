package rcs

import (
	"fmt"
	"io"
	"os"

	"golang.org/x/exp/mmap"
)

// ParseOption configures ParsePath behavior.
type ParseOption func(*parseOptions)

type parseOptions bool

// WithMmap toggles mmap-backed file reads in ParsePath.
func WithMmap(enabled bool) ParseOption {
	return func(options *parseOptions) {
		*options = parseOptions(enabled)
	}
}

// ParsePath opens an RCS file and parses it.
//
// By default ParsePath uses os.Open. Use WithMmap(true) to open the file using
// golang.org/x/exp/mmap and stream it through an io.Reader adapter.
func ParsePath(path string, opts ...ParseOption) (*File, error) {
	var options parseOptions
	for _, opt := range opts {
		opt(&options)
	}

	r, err := openParseReader(path, bool(options))
	if err != nil {
		return nil, fmt.Errorf("open %q: %w", path, err)
	}
	defer r.Close()

	return ParseFile(r)
}

type mmapReadCloser struct {
	*io.SectionReader
	closer func() error
}

func (m *mmapReadCloser) Close() error {
	return m.closer()
}

func openParseReader(path string, useMmap bool) (io.ReadCloser, error) {
	if useMmap {
		r, err := mmap.Open(path)
		if err != nil {
			return nil, err
		}
		return &mmapReadCloser{
			SectionReader: io.NewSectionReader(r, 0, int64(r.Len())),
			closer:        r.Close,
		}, nil
	}
	return os.Open(path)
}

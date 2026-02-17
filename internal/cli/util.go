package cli

import (
	"fmt"
	"golang.org/x/exp/mmap"
	"io"
	"os"
)

// ensureFiles checks if file arguments are provided.
// If not, it checks stdin:
// - If stdin is a TTY, it returns an error.
// - If stdin is a pipe/file, it returns []string{"-"} to read from stdin.
func ensureFiles(files []string) ([]string, error) {
	if len(files) > 0 {
		return files, nil
	}

	stat, err := os.Stdin.Stat()
	if err != nil {
		return nil, fmt.Errorf("stdin stat error: %w", err)
	}

	if (stat.Mode() & os.ModeCharDevice) != 0 {
		// Stdin is a terminal, and no files provided
		return nil, fmt.Errorf("no input files provided")
	}

	// Stdin is piped or redirected
	return []string{"-"}, nil
}

type mmapReadCloser struct {
	*io.SectionReader
	closer func() error
}

func (m *mmapReadCloser) Close() error {
	return m.closer()
}

func OpenFile(filename string, useMmap bool) (io.ReadCloser, error) {
	if filename == "-" {
		return io.NopCloser(os.Stdin), nil
	}
	if useMmap {
		r, err := mmap.Open(filename)
		if err != nil {
			return nil, err
		}
		sr := io.NewSectionReader(r, 0, int64(r.Len()))
		return &mmapReadCloser{
			SectionReader: sr,
			closer:        r.Close,
		}, nil
	}
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	return f, nil
}

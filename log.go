package rcs

import (
	"fmt"
)

// SetLog updates the log message for a specific revision.
func (f *File) SetLog(revision, message string) error {
	for _, rc := range f.RevisionContents {
		if rc.Revision == revision {
			rc.Log = message
			return nil
		}
	}
	return fmt.Errorf("revision %s not found", revision)
}

// GetLog retrieves the log message for a specific revision.
func (f *File) GetLog(revision string) (string, error) {
	for _, rc := range f.RevisionContents {
		if rc.Revision == revision {
			return rc.Log, nil
		}
	}
	return "", fmt.Errorf("revision %s not found", revision)
}

// ListLogs returns a map of revision to log message.
func (f *File) ListLogs() map[string]string {
	m := make(map[string]string)
	for _, rc := range f.RevisionContents {
		m[rc.Revision] = rc.Log
	}
	return m
}

package rcs

import "fmt"

// ChangeLogMessage updates the log message for the specified revision.
func (f *File) ChangeLogMessage(revision string, message string) error {
	for _, rc := range f.RevisionContents {
		if rc.Revision == revision {
			rc.Log = message
			return nil
		}
	}
	return fmt.Errorf("revision %s not found", revision)
}

// GetLogMessage retrieves the log message for the specified revision.
func (f *File) GetLogMessage(revision string) (string, error) {
	for _, rc := range f.RevisionContents {
		if rc.Revision == revision {
			return rc.Log, nil
		}
	}
	return "", fmt.Errorf("revision %s not found", revision)
}

// LogMessage represents a log message entry for a revision.
type LogMessage struct {
	Revision string
	Log      string
}

// ListLogMessages retrieves all log messages in the RCS file.
func (f *File) ListLogMessages() []LogMessage {
	var logs []LogMessage
	for _, rc := range f.RevisionContents {
		logs = append(logs, LogMessage{
			Revision: rc.Revision,
			Log:      rc.Log,
		})
	}
	return logs
}

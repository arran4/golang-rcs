package rcs

import (
	"testing"
	"time"
)

// The RCS text model normalizes to a trailing newline (documented on Checkin).
// This is intentional, but must be reported via CIVerdict.AppendedNewline so
// the CLI can warn that the working file's content was altered.
func TestCheckinAppendedNewlineFlag(t *testing.T) {
	d := WithDate(time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC))

	// Initial checkin, input missing a trailing newline -> flag set.
	f := NewFile()
	v, err := f.Checkin("alice", "init", "no newline", WithInitial{}, d)
	if err != nil {
		t.Fatalf("initial checkin: %v", err)
	}
	if !v.AppendedNewline {
		t.Error("initial: AppendedNewline = false, want true when input lacks a trailing newline")
	}

	// Normal checkin, input already newline-terminated -> flag clear.
	f2 := strictHead(t, "alice", true)
	v2, err := f2.Checkin("alice", "v2", "two\n", d, WithSetLock)
	if err != nil {
		t.Fatalf("terminated checkin: %v", err)
	}
	if v2.AppendedNewline {
		t.Error("terminated: AppendedNewline = true, want false when input already ends with newline")
	}

	// Normal checkin, input missing a trailing newline -> flag set.
	f3 := strictHead(t, "alice", true)
	v3, err := f3.Checkin("alice", "v2", "two lines\nno final", d, WithSetLock)
	if err != nil {
		t.Fatalf("unterminated checkin: %v", err)
	}
	if !v3.AppendedNewline {
		t.Error("unterminated: AppendedNewline = false, want true when input lacks a trailing newline")
	}
}

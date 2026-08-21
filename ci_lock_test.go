package rcs

import (
	"strings"
	"testing"
	"time"
)

// strictHead returns a strict File with an initial revision 1.1 authored by
// author. When lock is true, author holds the lock on 1.1.
func strictHead(t *testing.T, author string, lock bool) *File {
	t.Helper()
	f := NewFile()
	ops := []any{WithInitial{}, WithDate(time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC))}
	if lock {
		ops = append(ops, WithSetLock)
	}
	if _, err := f.Checkin(author, "init", "one\n", ops...); err != nil {
		t.Fatalf("initial checkin: %v", err)
	}
	return f
}

// secondCheckin adds a changed revision as user, returning the error (if any).
func secondCheckin(f *File, user string, extra ...any) error {
	ops := append([]any{WithDate(time.Date(2020, 1, 2, 0, 0, 0, 0, time.UTC))}, extra...)
	_, err := f.Checkin(user, "v2", "two\n", ops...)
	return err
}

func TestStrictCheckinNoLockRejected(t *testing.T) {
	f := strictHead(t, "alice", false)
	err := secondCheckin(f, "alice")
	if err == nil {
		t.Fatal("expected strict-lock error when no lock is held")
	}
	if !strings.Contains(err.Error(), "lock") {
		t.Errorf("error = %q, want it to mention lock", err)
	}
}

func TestStrictCheckinLockHeldSameUser(t *testing.T) {
	f := strictHead(t, "alice", true)
	if err := secondCheckin(f, "alice", WithSetLock); err != nil {
		t.Fatalf("checkin with lock held by same user should succeed: %v", err)
	}
	if f.Head != "1.2" {
		t.Errorf("head = %q, want 1.2", f.Head)
	}
}

func TestStrictCheckinLockHeldOtherUserRejected(t *testing.T) {
	f := strictHead(t, "alice", true)
	err := secondCheckin(f, "bob")
	if err == nil {
		t.Fatal("expected error when the lock is held by a different user")
	}
	if !strings.Contains(err.Error(), "alice") {
		t.Errorf("error = %q, want it to name the lock holder (alice)", err)
	}
}

func TestNonStrictCheckinNoLockSucceeds(t *testing.T) {
	f := strictHead(t, "alice", false)
	f.Strict = false
	if err := secondCheckin(f, "alice"); err != nil {
		t.Fatalf("non-strict checkin without a lock should succeed: %v", err)
	}
	if f.Head != "1.2" {
		t.Errorf("head = %q, want 1.2", f.Head)
	}
}

func TestStrictCheckinForceDoesNotBypassLock(t *testing.T) {
	f := strictHead(t, "alice", false)
	err := secondCheckin(f, "alice", WithForce{})
	if err == nil {
		t.Fatal("WithForce must not bypass the strict lock check")
	}
	if !strings.Contains(err.Error(), "lock") {
		t.Errorf("error = %q, want it to mention lock", err)
	}
}

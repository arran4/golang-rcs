package rcs

import (
	"strings"
	"testing"
	"time"
)

func TestCheckinBranchRevisionRejected(t *testing.T) {
	f := strictHead(t, "alice", true)
	err := secondCheckin(f, "alice", WithRevision("1.1.1.1"), WithSetLock)
	if err == nil {
		t.Fatal("expected branch-shaped revision 1.1.1.1 to be rejected")
	}
	if !strings.Contains(err.Error(), "branch") {
		t.Errorf("error = %q, want it to explain the branch limitation", err)
	}
	if f.Head != "1.1" {
		t.Errorf("head = %q, want it left unchanged at 1.1", f.Head)
	}
}

func TestCheckinDeepBranchRevisionRejected(t *testing.T) {
	f := strictHead(t, "alice", true)
	err := secondCheckin(f, "alice", WithRevision("1.1.1.1.1.1"), WithSetLock)
	if err == nil {
		t.Fatal("expected branch-shaped revision 1.1.1.1.1.1 to be rejected")
	}
	if !strings.Contains(err.Error(), "branch") {
		t.Errorf("error = %q, want it to explain the branch limitation", err)
	}
}

func TestInitialBranchRevisionRejected(t *testing.T) {
	f := NewFile()
	_, err := f.Checkin("alice", "init", "one\n",
		WithInitial{},
		WithDate(time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)),
		WithRevision("2.1.1.1"))
	if err == nil {
		t.Fatal("expected branch-shaped initial revision 2.1.1.1 to be rejected")
	}
	if !strings.Contains(err.Error(), "branch") {
		t.Errorf("error = %q, want it to explain the branch limitation", err)
	}
}

func TestCheckinTrunkRevisionAccepted(t *testing.T) {
	f := strictHead(t, "alice", true)
	if err := secondCheckin(f, "alice", WithRevision("1.5"), WithSetLock); err != nil {
		t.Fatalf("explicit trunk revision 1.5 should be accepted: %v", err)
	}
	if f.Head != "1.5" {
		t.Errorf("head = %q, want 1.5", f.Head)
	}
}

func TestCheckinAutoIncrementStaysTrunk(t *testing.T) {
	f := strictHead(t, "alice", true)
	if err := secondCheckin(f, "alice", WithSetLock); err != nil {
		t.Fatalf("auto-increment checkin should succeed: %v", err)
	}
	if strings.Count(f.Head, ".") != 1 {
		t.Errorf("head = %q, want a two-component trunk revision", f.Head)
	}
}

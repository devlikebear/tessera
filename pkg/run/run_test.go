package run

import (
	"errors"
	"testing"
	"time"

	"github.com/devlikebear/tessera/pkg/mandate"
)

func TestRunCannotStartWithoutApprovedMandate(t *testing.T) {
	r := New("run-1")
	m := mandate.New("mandate-1", "Draft a novel", "Plan a small writing team")

	err := r.Start(m)
	if !errors.Is(err, mandate.ErrNotApproved) {
		t.Fatalf("Start() error = %v, want ErrNotApproved", err)
	}
	if r.Status != StatusAwaitingApproval {
		t.Fatalf("Status = %q, want %q", r.Status, StatusAwaitingApproval)
	}
}

func TestRunStartsWithApprovedMandateAndRecordsEvents(t *testing.T) {
	r := New("run-1")
	m, err := mandate.New("mandate-1", "Draft a novel", "Plan a small writing team").
		Approve("operator", time.Date(2026, 5, 24, 9, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("Approve() error = %v", err)
	}

	if err := r.Start(m); err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	if r.Status != StatusRunning {
		t.Fatalf("Status = %q, want %q", r.Status, StatusRunning)
	}

	events := r.Events.Events()
	if len(events) != 2 {
		t.Fatalf("events len = %d, want 2", len(events))
	}
	if events[0].To != string(StatusReady) || events[1].To != string(StatusRunning) {
		t.Fatalf("event transitions = %q -> %q, want ready -> running", events[0].To, events[1].To)
	}
}

func TestRunClosure(t *testing.T) {
	r := New("run-1")
	if err := r.Close(ClosureNormal, "all tasks succeeded"); err != nil {
		t.Fatalf("Close() error = %v", err)
	}
	if r.Status != StatusClosed {
		t.Fatalf("Status = %q, want %q", r.Status, StatusClosed)
	}
	if r.Closure != ClosureNormal {
		t.Fatalf("Closure = %q, want %q", r.Closure, ClosureNormal)
	}
}

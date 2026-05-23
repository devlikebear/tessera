package mandate

import (
	"errors"
	"testing"
	"time"
)

func TestMandateRequiresExplicitApproval(t *testing.T) {
	m := New("mandate-1", "Draft a novel", "Plan a small writing team")

	if m.IsApproved() {
		t.Fatal("new mandate should not be approved")
	}
	if err := m.RequireApproved(); !errors.Is(err, ErrNotApproved) {
		t.Fatalf("RequireApproved() error = %v, want ErrNotApproved", err)
	}
}

func TestApproveMandate(t *testing.T) {
	now := time.Date(2026, 5, 24, 9, 0, 0, 0, time.UTC)
	m, err := New("mandate-1", "Draft a novel", "Plan a small writing team").Approve("operator", now)
	if err != nil {
		t.Fatalf("Approve() error = %v", err)
	}

	if !m.IsApproved() {
		t.Fatal("approved mandate should pass IsApproved")
	}
	if err := m.RequireApproved(); err != nil {
		t.Fatalf("RequireApproved() error = %v", err)
	}
	if m.Approval.ApprovedBy != "operator" {
		t.Fatalf("ApprovedBy = %q, want operator", m.Approval.ApprovedBy)
	}
}

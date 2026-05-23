package mandate

import (
	"errors"
	"strings"
	"time"
)

var (
	ErrInvalidMandate = errors.New("invalid mandate")
	ErrNotApproved    = errors.New("mandate is not approved")
)

type Status string

const (
	StatusDraft    Status = "draft"
	StatusApproved Status = "approved"
	StatusRejected Status = "rejected"
)

type Mandate struct {
	ID          string
	Goal        string
	PlanSummary string
	Status      Status
	Approval    Approval
}

type Approval struct {
	ApprovedBy string
	ApprovedAt time.Time
}

func New(id, goal, planSummary string) Mandate {
	return Mandate{
		ID:          strings.TrimSpace(id),
		Goal:        strings.TrimSpace(goal),
		PlanSummary: strings.TrimSpace(planSummary),
		Status:      StatusDraft,
	}
}

func (m Mandate) Validate() error {
	if m.ID == "" || m.Goal == "" || m.PlanSummary == "" {
		return ErrInvalidMandate
	}
	return nil
}

func (m Mandate) IsApproved() bool {
	return m.Status == StatusApproved && m.Approval.ApprovedBy != "" && !m.Approval.ApprovedAt.IsZero()
}

func (m Mandate) RequireApproved() error {
	if !m.IsApproved() {
		return ErrNotApproved
	}
	return nil
}

func (m Mandate) Approve(actor string, at time.Time) (Mandate, error) {
	if err := m.Validate(); err != nil {
		return Mandate{}, err
	}
	actor = strings.TrimSpace(actor)
	if actor == "" || at.IsZero() {
		return Mandate{}, ErrInvalidMandate
	}
	m.Status = StatusApproved
	m.Approval = Approval{ApprovedBy: actor, ApprovedAt: at}
	return m, nil
}

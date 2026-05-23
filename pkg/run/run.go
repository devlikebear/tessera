package run

import (
	"errors"
	"strings"

	"github.com/devlikebear/tessera/pkg/mandate"
)

var (
	ErrInvalidRun = errors.New("invalid run")
	ErrClosedRun  = errors.New("run is already closed")
)

type Run struct {
	ID      string
	Status  Status
	Closure ClosureKind
	Events  EventLog
}

func New(id string) *Run {
	return &Run{
		ID:      strings.TrimSpace(id),
		Status:  StatusInitialized,
		Closure: ClosureNone,
	}
}

func (r *Run) AwaitApproval() error {
	return r.transition(StatusAwaitingApproval, "waiting for mandate approval")
}

func (r *Run) Start(m mandate.Mandate) error {
	if r.ID == "" {
		return ErrInvalidRun
	}
	if r.Status == StatusClosed {
		return ErrClosedRun
	}
	if err := m.RequireApproved(); err != nil {
		if r.Status == StatusInitialized {
			_ = r.AwaitApproval()
		}
		return err
	}
	if r.Status == StatusInitialized || r.Status == StatusAwaitingApproval {
		if err := r.transition(StatusReady, "mandate approved"); err != nil {
			return err
		}
	}
	return r.transition(StatusRunning, "run started")
}

func (r *Run) Close(kind ClosureKind, message string) error {
	if r.ID == "" {
		return ErrInvalidRun
	}
	if r.Status == StatusClosed {
		return ErrClosedRun
	}
	if kind != ClosureNormal && kind != ClosureAbnormal {
		return ErrInvalidRun
	}
	from := r.Status
	r.Status = StatusClosed
	r.Closure = kind
	r.Events.Append(Event{
		Type:    EventClosure,
		RunID:   r.ID,
		From:    string(from),
		To:      string(kind),
		Message: strings.TrimSpace(message),
	})
	return nil
}

func (r *Run) transition(to Status, message string) error {
	if r.ID == "" {
		return ErrInvalidRun
	}
	if r.Status == StatusClosed {
		return ErrClosedRun
	}
	from := r.Status
	if from == to {
		return nil
	}
	r.Status = to
	r.Events.Append(Event{
		Type:    EventRunTransition,
		RunID:   r.ID,
		From:    string(from),
		To:      string(to),
		Message: strings.TrimSpace(message),
	})
	return nil
}

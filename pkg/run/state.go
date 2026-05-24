package run

import (
	"context"
	"time"

	"github.com/devlikebear/tessera/pkg/queue"
)

type Status string

const (
	StatusInitialized      Status = "initialized"
	StatusAwaitingApproval Status = "awaiting_approval"
	StatusReady            Status = "ready"
	StatusRunning          Status = "running"
	StatusClosed           Status = "closed"
)

type ClosureKind string

const (
	ClosureNone     ClosureKind = "none"
	ClosureNormal   ClosureKind = "normal"
	ClosureAbnormal ClosureKind = "abnormal"
)

type EventType string

const (
	EventRunTransition  EventType = "run.transition"
	EventTaskTransition EventType = "task.transition"
	EventClosure        EventType = "run.closure"
)

type Event struct {
	Seq     int       `json:"seq"`
	At      time.Time `json:"at"`
	Type    EventType `json:"type"`
	RunID   string    `json:"run_id"`
	TaskID  string    `json:"task_id,omitempty"`
	Role    string    `json:"role,omitempty"`
	From    string    `json:"from,omitempty"`
	To      string    `json:"to,omitempty"`
	Message string    `json:"message,omitempty"`
}

type EventSink interface {
	OnEvent(context.Context, Event) error
}

type EventSinkFunc func(context.Context, Event) error

func (f EventSinkFunc) OnEvent(ctx context.Context, event Event) error {
	return f(ctx, event)
}

type EventLog struct {
	events []Event
}

func (l *EventLog) Append(event Event) Event {
	event.Seq = len(l.events) + 1
	if event.At.IsZero() {
		event.At = time.Now().UTC()
	}
	l.events = append(l.events, event)
	return event
}

func (l *EventLog) Events() []Event {
	out := make([]Event, len(l.events))
	copy(out, l.events)
	return out
}

func (l *EventLog) RecordTask(runID, taskID string, from, to queue.TaskStatus, message string) Event {
	return l.RecordTaskWithRole(runID, taskID, "", from, to, message)
}

func (l *EventLog) RecordTaskWithRole(runID, taskID, role string, from, to queue.TaskStatus, message string) Event {
	return l.Append(Event{
		Type:    EventTaskTransition,
		RunID:   runID,
		TaskID:  taskID,
		Role:    role,
		From:    string(from),
		To:      string(to),
		Message: message,
	})
}

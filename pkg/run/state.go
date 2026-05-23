package run

import (
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
	Seq     int
	At      time.Time
	Type    EventType
	RunID   string
	TaskID  string
	From    string
	To      string
	Message string
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
	return l.Append(Event{
		Type:    EventTaskTransition,
		RunID:   runID,
		TaskID:  taskID,
		From:    string(from),
		To:      string(to),
		Message: message,
	})
}

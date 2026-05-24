package run

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/devlikebear/tessera/pkg/queue"
)

const EventSchemaVersion = 1

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
	EventTaskQueued     EventType = "task.queued"
	EventTaskStarted    EventType = "task.started"
	EventTaskRetrying   EventType = "task.retrying"
	EventTaskSucceeded  EventType = "task.succeeded"
	EventTaskFailed     EventType = "task.failed"
	EventClosure        EventType = "run.closure"

	EventAgentIterationStarted   EventType = "agent.iteration.started"
	EventAgentIterationCompleted EventType = "agent.iteration.completed"
	EventLLMRequestStarted       EventType = "llm.request.started"
	EventLLMResponseCompleted    EventType = "llm.response.completed"
	EventToolCallStarted         EventType = "tool.call.started"
	EventToolCallCompleted       EventType = "tool.call.completed"
)

type Event struct {
	SchemaVersion int               `json:"schema_version"`
	Seq           int               `json:"seq"`
	At            time.Time         `json:"at"`
	Type          EventType         `json:"type"`
	RunID         string            `json:"run_id"`
	TaskID        string            `json:"task_id,omitempty"`
	Role          string            `json:"role,omitempty"`
	Stage         string            `json:"stage,omitempty"`
	Attempt       int               `json:"attempt,omitempty"`
	MaxAttempts   int               `json:"max_attempts,omitempty"`
	WorkerID      string            `json:"worker_id,omitempty"`
	LeaseID       string            `json:"lease_id,omitempty"`
	DependsOn     []string          `json:"depends_on,omitempty"`
	From          string            `json:"from,omitempty"`
	To            string            `json:"to,omitempty"`
	Message       string            `json:"message,omitempty"`
	InputSummary  string            `json:"input_summary,omitempty"`
	OutputSummary string            `json:"output_summary,omitempty"`
	Error         string            `json:"error,omitempty"`
	Attributes    map[string]string `json:"attributes,omitempty"`
}

type EventSink interface {
	OnEvent(context.Context, Event) error
}

type EventSinkFunc func(context.Context, Event) error

func (f EventSinkFunc) OnEvent(ctx context.Context, event Event) error {
	return f(ctx, event)
}

type EventLog struct {
	mu     sync.Mutex
	events []Event
}

func (l *EventLog) Append(event Event) Event {
	l.mu.Lock()
	defer l.mu.Unlock()

	if event.SchemaVersion == 0 {
		event.SchemaVersion = EventSchemaVersion
	}
	event.Seq = len(l.events) + 1
	if event.At.IsZero() {
		event.At = time.Now().UTC()
	}
	event = cloneEvent(event)
	l.events = append(l.events, event)
	return event
}

func (l *EventLog) Events() []Event {
	l.mu.Lock()
	defer l.mu.Unlock()

	out := make([]Event, len(l.events))
	for i := range l.events {
		out[i] = cloneEvent(l.events[i])
	}
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

func NewTaskEvent(eventType EventType, runID string, task queue.Task, from, to queue.TaskStatus, message string) Event {
	return Event{
		Type:         eventType,
		RunID:        runID,
		TaskID:       task.ID,
		Role:         task.Role,
		Stage:        task.Stage,
		Attempt:      task.Attempts,
		MaxAttempts:  task.MaxAttempts,
		DependsOn:    append([]string(nil), task.DependsOn...),
		From:         string(from),
		To:           string(to),
		Message:      strings.TrimSpace(message),
		InputSummary: summarizeEventText(task.Payload),
	}
}

func NewLeaseEvent(eventType EventType, runID string, lease queue.Lease, from, to queue.TaskStatus, message string) Event {
	event := NewTaskEvent(eventType, runID, lease.Task, from, to, message)
	event.WorkerID = lease.WorkerID
	event.LeaseID = lease.ID
	return event
}

func cloneEvent(event Event) Event {
	event.DependsOn = append([]string(nil), event.DependsOn...)
	if len(event.Attributes) > 0 {
		attrs := make(map[string]string, len(event.Attributes))
		for key, value := range event.Attributes {
			attrs[key] = value
		}
		event.Attributes = attrs
	}
	return event
}

func summarizeEventText(text string) string {
	text = strings.Join(strings.Fields(text), " ")
	if len(text) <= 240 {
		return text
	}
	return text[:237] + "..."
}

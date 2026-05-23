package queue

import (
	"context"
	"errors"
	"time"
)

var (
	ErrEmpty        = errors.New("queue is empty")
	ErrTaskNotFound = errors.New("task not found")
	ErrInvalidTask  = errors.New("invalid task")
)

type TaskStatus string

const (
	TaskQueued    TaskStatus = "queued"
	TaskRunning   TaskStatus = "running"
	TaskSucceeded TaskStatus = "succeeded"
	TaskRetrying  TaskStatus = "retrying"
	TaskFailed    TaskStatus = "failed"
)

type Task struct {
	ID             string
	Role           string
	Stage          string
	Payload        string
	Status         TaskStatus
	Attempts       int
	MaxAttempts    int
	IdempotencyKey string
	DependsOn      []string
}

type Lease struct {
	ID        string
	Task      Task
	WorkerID  string
	ExpiresAt time.Time
}

type Queue interface {
	Enqueue(context.Context, Task) error
	Dequeue(context.Context, string) (Lease, error)
	Ack(context.Context, string) error
	Retry(context.Context, string, string) error
	Fail(context.Context, string, string) error
}

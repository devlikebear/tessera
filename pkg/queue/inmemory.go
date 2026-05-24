package queue

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"
)

var ErrInvalidLease = errors.New("invalid lease")

type InMemoryOption func(*InMemory)

type InMemory struct {
	mu           sync.Mutex
	tasks        map[string]Task
	ready        []string
	leases       map[string]Lease
	deadLetters  []Task
	leaseTimeout time.Duration
	leaseSeq     int
	now          func() time.Time
}

func NewInMemory(opts ...InMemoryOption) *InMemory {
	q := &InMemory{
		tasks:        make(map[string]Task),
		leases:       make(map[string]Lease),
		leaseTimeout: 30 * time.Second,
		now:          func() time.Time { return time.Now().UTC() },
	}
	for _, opt := range opts {
		opt(q)
	}
	return q
}

func WithLeaseTimeout(timeout time.Duration) InMemoryOption {
	return func(q *InMemory) {
		if timeout > 0 {
			q.leaseTimeout = timeout
		}
	}
}

func WithClock(now func() time.Time) InMemoryOption {
	return func(q *InMemory) {
		if now != nil {
			q.now = now
		}
	}
}

func (q *InMemory) Enqueue(_ context.Context, task Task) error {
	task.ID = strings.TrimSpace(task.ID)
	if task.ID == "" {
		return ErrInvalidTask
	}
	q.mu.Lock()
	defer q.mu.Unlock()

	if task.MaxAttempts <= 0 {
		task.MaxAttempts = 1
	}
	task.Status = TaskQueued
	q.tasks[task.ID] = cloneTask(task)
	q.ready = append(q.ready, task.ID)
	return nil
}

func (q *InMemory) Dequeue(_ context.Context, workerID string) (Lease, error) {
	workerID = strings.TrimSpace(workerID)
	if workerID == "" {
		return Lease{}, ErrInvalidLease
	}
	q.mu.Lock()
	defer q.mu.Unlock()

	for len(q.ready) > 0 {
		id := q.ready[0]
		q.ready = q.ready[1:]
		task, ok := q.tasks[id]
		if !ok || task.Status != TaskQueued {
			continue
		}
		task.Status = TaskRunning
		task.Attempts++
		q.tasks[id] = task
		q.leaseSeq++
		lease := Lease{
			ID:        fmt.Sprintf("lease-%d", q.leaseSeq),
			Task:      cloneTask(task),
			WorkerID:  workerID,
			ExpiresAt: q.now().Add(q.leaseTimeout),
		}
		q.leases[lease.ID] = lease
		return lease, nil
	}
	return Lease{}, ErrEmpty
}

func (q *InMemory) Ack(_ context.Context, leaseID string) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	lease, ok := q.leases[leaseID]
	if !ok {
		return ErrInvalidLease
	}
	delete(q.leases, leaseID)
	task := lease.Task
	task.Status = TaskSucceeded
	q.tasks[task.ID] = task
	return nil
}

func (q *InMemory) Retry(_ context.Context, leaseID, _ string) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	lease, ok := q.leases[leaseID]
	if !ok {
		return ErrInvalidLease
	}
	delete(q.leases, leaseID)
	return q.retryLocked(lease.Task)
}

func (q *InMemory) Fail(_ context.Context, leaseID, _ string) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	lease, ok := q.leases[leaseID]
	if !ok {
		return ErrInvalidLease
	}
	delete(q.leases, leaseID)
	q.failLocked(lease.Task)
	return nil
}

func (q *InMemory) RequeueExpired(now time.Time) []Task {
	q.mu.Lock()
	defer q.mu.Unlock()

	var recovered []Task
	for leaseID, lease := range q.leases {
		if now.Before(lease.ExpiresAt) {
			continue
		}
		delete(q.leases, leaseID)
		before := lease.Task
		if before.Attempts < before.MaxAttempts {
			_ = q.retryLocked(before)
			after := before
			after.Status = TaskQueued
			recovered = append(recovered, cloneTask(after))
			continue
		}
		q.failLocked(before)
	}
	return recovered
}

func (q *InMemory) Snapshot() []Task {
	q.mu.Lock()
	defer q.mu.Unlock()

	out := make([]Task, 0, len(q.tasks))
	for _, task := range q.tasks {
		out = append(out, cloneTask(task))
	}
	return out
}

func (q *InMemory) DeadLetters() []Task {
	q.mu.Lock()
	defer q.mu.Unlock()

	out := make([]Task, len(q.deadLetters))
	for i := range q.deadLetters {
		out[i] = cloneTask(q.deadLetters[i])
	}
	return out
}

func (q *InMemory) retryLocked(task Task) error {
	if task.Attempts >= task.MaxAttempts {
		q.failLocked(task)
		return nil
	}
	task.Status = TaskQueued
	q.tasks[task.ID] = task
	q.ready = append(q.ready, task.ID)
	return nil
}

func (q *InMemory) failLocked(task Task) {
	task.Status = TaskFailed
	q.tasks[task.ID] = task
	q.deadLetters = append(q.deadLetters, cloneTask(task))
}

func cloneTask(task Task) Task {
	task.DependsOn = append([]string(nil), task.DependsOn...)
	return task
}

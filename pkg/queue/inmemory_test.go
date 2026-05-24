package queue

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestInMemoryQueueAckLifecycle(t *testing.T) {
	q := NewInMemory()
	ctx := context.Background()

	if err := q.Enqueue(ctx, Task{ID: "task-1", Role: "writer"}); err != nil {
		t.Fatalf("Enqueue() error = %v", err)
	}
	lease, err := q.Dequeue(ctx, "worker-1")
	if err != nil {
		t.Fatalf("Dequeue() error = %v", err)
	}
	if lease.Task.Status != TaskRunning {
		t.Fatalf("lease status = %q, want %q", lease.Task.Status, TaskRunning)
	}
	if err := q.Ack(ctx, lease.ID); err != nil {
		t.Fatalf("Ack() error = %v", err)
	}

	tasks := q.Snapshot()
	if len(tasks) != 1 || tasks[0].Status != TaskSucceeded {
		t.Fatalf("snapshot = %+v, want one succeeded task", tasks)
	}
}

func TestInMemoryQueueRetryBudgetMovesToDeadLetter(t *testing.T) {
	q := NewInMemory()
	ctx := context.Background()

	if err := q.Enqueue(ctx, Task{ID: "task-1", Role: "writer", MaxAttempts: 1}); err != nil {
		t.Fatalf("Enqueue() error = %v", err)
	}
	lease, err := q.Dequeue(ctx, "worker-1")
	if err != nil {
		t.Fatalf("Dequeue() error = %v", err)
	}
	if err := q.Retry(ctx, lease.ID, "failed once"); err != nil {
		t.Fatalf("Retry() error = %v", err)
	}

	dead := q.DeadLetters()
	if len(dead) != 1 {
		t.Fatalf("dead letters len = %d, want 1", len(dead))
	}
	if dead[0].Status != TaskFailed {
		t.Fatalf("dead status = %q, want %q", dead[0].Status, TaskFailed)
	}
}

func TestInMemoryQueueRecoversExpiredLeaseWithinBudget(t *testing.T) {
	base := time.Date(2026, 5, 24, 11, 0, 0, 0, time.UTC)
	q := NewInMemory(WithLeaseTimeout(time.Second), WithClock(func() time.Time { return base }))
	ctx := context.Background()

	if err := q.Enqueue(ctx, Task{ID: "task-1", Role: "writer", MaxAttempts: 2}); err != nil {
		t.Fatalf("Enqueue() error = %v", err)
	}
	if _, err := q.Dequeue(ctx, "worker-1"); err != nil {
		t.Fatalf("Dequeue() error = %v", err)
	}

	recovered := q.RequeueExpired(base.Add(2 * time.Second))
	if len(recovered) != 1 {
		t.Fatalf("recovered len = %d, want 1", len(recovered))
	}
	lease, err := q.Dequeue(ctx, "worker-2")
	if err != nil {
		t.Fatalf("Dequeue() after recovery error = %v", err)
	}
	if lease.Task.Attempts != 2 {
		t.Fatalf("attempts = %d, want 2", lease.Task.Attempts)
	}
}

func TestInMemoryQueueReturnsErrEmpty(t *testing.T) {
	_, err := NewInMemory().Dequeue(context.Background(), "worker-1")
	if !errors.Is(err, ErrEmpty) {
		t.Fatalf("Dequeue() error = %v, want ErrEmpty", err)
	}
}

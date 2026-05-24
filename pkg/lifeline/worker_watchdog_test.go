package lifeline

import (
	"context"
	"testing"
	"time"

	"github.com/devlikebear/tessera/pkg/queue"
)

func TestWatchdogRequeuesExpiredWork(t *testing.T) {
	base := time.Date(2026, 5, 24, 12, 0, 0, 0, time.UTC)
	q := queue.NewInMemory(queue.WithLeaseTimeout(time.Second), queue.WithClock(func() time.Time { return base }))
	ctx := context.Background()
	if err := q.Enqueue(ctx, queue.Task{ID: "task-1", Role: "writer", MaxAttempts: 2}); err != nil {
		t.Fatalf("Enqueue() error = %v", err)
	}
	if _, err := q.Dequeue(ctx, "worker-1"); err != nil {
		t.Fatalf("Dequeue() error = %v", err)
	}

	recovered := Watchdog{
		Queue: q,
		Now:   func() time.Time { return base.Add(2 * time.Second) },
	}.RecoverStalled(ctx)
	if len(recovered) != 1 {
		t.Fatalf("recovered len = %d, want 1", len(recovered))
	}
}

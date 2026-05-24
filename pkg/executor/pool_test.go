package executor

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/devlikebear/tessera/pkg/queue"
)

func TestPoolRunsTasksInParallel(t *testing.T) {
	ctx := context.Background()
	q := queue.NewInMemory()
	for _, id := range []string{"a", "b"} {
		if err := q.Enqueue(ctx, queue.Task{ID: id, Role: "writer"}); err != nil {
			t.Fatalf("Enqueue() error = %v", err)
		}
	}

	started := make(chan struct{}, 2)
	release := make(chan struct{})
	pool := Pool{
		Queue:   q,
		Workers: 2,
		Executor: TaskHandler(func(ctx context.Context, task queue.Task) (Result, error) {
			started <- struct{}{}
			<-release
			return Result{Output: task.ID}, nil
		}),
	}

	done := make(chan Summary, 1)
	go func() {
		summary, err := pool.RunUntilIdle(ctx)
		if err != nil {
			t.Errorf("RunUntilIdle() error = %v", err)
		}
		done <- summary
	}()

	<-started
	<-started
	close(release)

	select {
	case summary := <-done:
		if summary.Succeeded != 2 {
			t.Fatalf("Succeeded = %d, want 2", summary.Succeeded)
		}
	case <-time.After(time.Second):
		t.Fatal("pool did not finish")
	}
}

func TestPoolRetriesThenSucceeds(t *testing.T) {
	ctx := context.Background()
	q := queue.NewInMemory()
	if err := q.Enqueue(ctx, queue.Task{ID: "task-1", Role: "writer", MaxAttempts: 2}); err != nil {
		t.Fatalf("Enqueue() error = %v", err)
	}

	var attempts int32
	pool := Pool{
		Queue:   q,
		Workers: 1,
		Executor: TaskHandler(func(ctx context.Context, task queue.Task) (Result, error) {
			if atomic.AddInt32(&attempts, 1) == 1 {
				return Result{}, errors.New("transient")
			}
			return Result{Output: "ok"}, nil
		}),
	}

	summary, err := pool.RunUntilIdle(ctx)
	if err != nil {
		t.Fatalf("RunUntilIdle() error = %v", err)
	}
	if summary.Succeeded != 1 || summary.Retried != 1 {
		t.Fatalf("summary = %+v, want one success and one retry", summary)
	}
}

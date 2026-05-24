package executor

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/devlikebear/tessera/pkg/queue"
)

type Queue interface {
	Dequeue(context.Context, string) (queue.Lease, error)
	Ack(context.Context, string) error
	Retry(context.Context, string, string) error
}

type Pool struct {
	Queue      Queue
	Executor   TaskExecutor
	Workers    int
	RoleLimits map[string]int
}

type Summary struct {
	Succeeded int
	Retried   int
	Failed    int
}

func (p Pool) RunUntilIdle(ctx context.Context) (Summary, error) {
	if p.Queue == nil || p.Executor == nil {
		return Summary{}, errors.New("executor pool requires queue and executor")
	}
	workers := p.Workers
	if workers <= 0 {
		workers = 1
	}

	limits := make(map[string]chan struct{}, len(p.RoleLimits))
	for role, n := range p.RoleLimits {
		if n > 0 {
			limits[role] = make(chan struct{}, n)
		}
	}

	var wg sync.WaitGroup
	var mu sync.Mutex
	var summary Summary
	var firstErr error

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			workerID := fmt.Sprintf("worker-%d", index+1)
			for {
				if ctx.Err() != nil {
					mu.Lock()
					if firstErr == nil {
						firstErr = ctx.Err()
					}
					mu.Unlock()
					return
				}
				lease, err := p.Queue.Dequeue(ctx, workerID)
				if errors.Is(err, queue.ErrEmpty) {
					return
				}
				if err != nil {
					mu.Lock()
					if firstErr == nil {
						firstErr = err
					}
					mu.Unlock()
					return
				}

				release := acquireRole(limits, lease.Task.Role)
				_, execErr := p.Executor.Execute(ctx, lease.Task)
				release()
				if execErr != nil {
					if retryErr := p.Queue.Retry(ctx, lease.ID, execErr.Error()); retryErr != nil {
						mu.Lock()
						summary.Failed++
						if firstErr == nil {
							firstErr = retryErr
						}
						mu.Unlock()
						return
					}
					mu.Lock()
					summary.Retried++
					mu.Unlock()
					continue
				}
				if err := p.Queue.Ack(ctx, lease.ID); err != nil {
					mu.Lock()
					if firstErr == nil {
						firstErr = err
					}
					mu.Unlock()
					return
				}
				mu.Lock()
				summary.Succeeded++
				mu.Unlock()
			}
		}(i)
	}
	wg.Wait()
	return summary, firstErr
}

func acquireRole(limits map[string]chan struct{}, role string) func() {
	sem, ok := limits[role]
	if !ok {
		return func() {}
	}
	sem <- struct{}{}
	return func() { <-sem }
}

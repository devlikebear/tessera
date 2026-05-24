package run

import (
	"context"

	"github.com/devlikebear/tessera/pkg/queue"
)

type DispatchQueue interface {
	Enqueue(context.Context, queue.Task) error
}

func DispatchReady(ctx context.Context, graph TaskGraph, completed, enqueued map[string]struct{}, q DispatchQueue, maxAttempts int) ([]queue.Task, error) {
	if maxAttempts <= 0 {
		maxAttempts = 1
	}
	var dispatched []queue.Task
	for _, node := range (Scheduler{Graph: graph}).Ready(completed, enqueued) {
		task := queue.Task{
			ID:             node.ID,
			Role:           node.Role,
			Stage:          node.Stage,
			Payload:        node.Description,
			Status:         queue.TaskQueued,
			MaxAttempts:    maxAttempts,
			IdempotencyKey: graph.PlanID + "/" + node.ID,
			DependsOn:      append([]string(nil), node.DependsOn...),
		}
		if err := q.Enqueue(ctx, task); err != nil {
			return nil, err
		}
		enqueued[node.ID] = struct{}{}
		dispatched = append(dispatched, task)
	}
	return dispatched, nil
}

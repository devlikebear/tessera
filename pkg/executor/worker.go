package executor

import (
	"context"

	"github.com/devlikebear/tessera/pkg/queue"
)

type Result struct {
	Output string
}

type TaskExecutor interface {
	Execute(context.Context, queue.Task) (Result, error)
}

type TaskHandler func(context.Context, queue.Task) (Result, error)

func (h TaskHandler) Execute(ctx context.Context, task queue.Task) (Result, error) {
	return h(ctx, task)
}

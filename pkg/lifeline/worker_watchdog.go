package lifeline

import (
	"context"
	"time"

	"github.com/devlikebear/tessera/pkg/queue"
)

type RecoverableQueue interface {
	RequeueExpired(time.Time) []queue.Task
}

type Watchdog struct {
	Queue  RecoverableQueue
	Now    func() time.Time
	Policy RequeuePolicy
}

func (w Watchdog) RecoverStalled(ctx context.Context) []queue.Task {
	if ctx.Err() != nil || w.Queue == nil {
		return nil
	}
	now := time.Now().UTC()
	if w.Now != nil {
		now = w.Now()
	}
	recovered := w.Queue.RequeueExpired(now)
	if !w.Policy.Allows(len(recovered)) {
		return nil
	}
	return recovered
}

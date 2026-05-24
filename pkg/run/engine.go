package run

import (
	"context"
	"errors"

	"github.com/devlikebear/tessera/pkg/executor"
	"github.com/devlikebear/tessera/pkg/mandate"
	"github.com/devlikebear/tessera/pkg/queue"
)

var ErrInvalidExecution = errors.New("invalid execution")

type ExecutionConfig struct {
	RunID       string
	Mandate     mandate.Mandate
	Graph       TaskGraph
	Queue       *queue.InMemory
	Executor    executor.TaskExecutor
	EventSink   EventSink
	Workers     int
	RoleLimits  map[string]int
	MaxAttempts int
}

type ExecutionResult struct {
	Run        *Run
	Checkpoint Checkpoint
	Report     Report
}

func ExecuteTaskGraph(ctx context.Context, cfg ExecutionConfig) (ExecutionResult, error) {
	if cfg.Graph.PlanID == "" || len(cfg.Graph.Nodes) == 0 || cfg.Queue == nil || cfg.Executor == nil {
		return ExecutionResult{}, ErrInvalidExecution
	}
	runID := cfg.RunID
	if runID == "" {
		runID = cfg.Graph.PlanID + "-run"
	}
	r := New(runID)
	if err := r.Start(cfg.Mandate); err != nil {
		_ = emitNewEvents(ctx, cfg.EventSink, r, new(int))
		return ExecutionResult{}, err
	}
	lastEmitted := 0
	if err := emitNewEvents(ctx, cfg.EventSink, r, &lastEmitted); err != nil {
		return finishExecution(r, cfg.Queue), err
	}

	completed := make(map[string]struct{})
	enqueued := make(map[string]struct{})
	maxAttempts := cfg.MaxAttempts
	if maxAttempts <= 0 {
		maxAttempts = 1
	}

	for {
		if len(completed) == len(cfg.Graph.Nodes) {
			_ = r.Close(ClosureNormal, "all tasks succeeded")
			if emitErr := emitNewEvents(ctx, cfg.EventSink, r, &lastEmitted); emitErr != nil {
				return finishExecution(r, cfg.Queue), emitErr
			}
			return finishExecution(r, cfg.Queue), nil
		}

		dispatched, err := DispatchReady(ctx, cfg.Graph, completed, enqueued, cfg.Queue, maxAttempts)
		if err != nil {
			return ExecutionResult{}, err
		}
		for _, task := range dispatched {
			r.Events.RecordTaskWithRole(r.ID, task.ID, task.Role, "", queue.TaskQueued, "task queued")
		}
		if err := emitNewEvents(ctx, cfg.EventSink, r, &lastEmitted); err != nil {
			return finishExecution(r, cfg.Queue), err
		}

		summary, err := (executor.Pool{
			Queue:      cfg.Queue,
			Executor:   cfg.Executor,
			Workers:    cfg.Workers,
			RoleLimits: cfg.RoleLimits,
		}).RunUntilIdle(ctx)
		if err != nil {
			_ = r.Close(ClosureAbnormal, err.Error())
			if emitErr := emitNewEvents(ctx, cfg.EventSink, r, &lastEmitted); emitErr != nil {
				return finishExecution(r, cfg.Queue), emitErr
			}
			return finishExecution(r, cfg.Queue), err
		}

		progressed := false
		for _, task := range cfg.Queue.Snapshot() {
			switch task.Status {
			case queue.TaskSucceeded:
				if _, ok := completed[task.ID]; !ok {
					completed[task.ID] = struct{}{}
					progressed = true
					r.Events.RecordTaskWithRole(r.ID, task.ID, task.Role, queue.TaskRunning, queue.TaskSucceeded, "task succeeded")
				}
			case queue.TaskFailed:
				_ = r.Close(ClosureAbnormal, "task failed: "+task.ID)
				if emitErr := emitNewEvents(ctx, cfg.EventSink, r, &lastEmitted); emitErr != nil {
					return finishExecution(r, cfg.Queue), emitErr
				}
				return finishExecution(r, cfg.Queue), nil
			}
		}
		if err := emitNewEvents(ctx, cfg.EventSink, r, &lastEmitted); err != nil {
			return finishExecution(r, cfg.Queue), err
		}

		if len(dispatched) == 0 && !progressed && summary.Retried == 0 {
			_ = r.Close(ClosureAbnormal, "no dispatchable tasks remain")
			if emitErr := emitNewEvents(ctx, cfg.EventSink, r, &lastEmitted); emitErr != nil {
				return finishExecution(r, cfg.Queue), emitErr
			}
			return finishExecution(r, cfg.Queue), nil
		}
	}
}

func emitNewEvents(ctx context.Context, sink EventSink, r *Run, lastEmitted *int) error {
	if sink == nil {
		return nil
	}
	for _, event := range r.Events.Events() {
		if event.Seq <= *lastEmitted {
			continue
		}
		if err := sink.OnEvent(ctx, event); err != nil {
			return err
		}
		*lastEmitted = event.Seq
	}
	return nil
}

func finishExecution(r *Run, q *queue.InMemory) ExecutionResult {
	tasks := q.Snapshot()
	return ExecutionResult{
		Run:        r,
		Checkpoint: NewCheckpoint(r, tasks),
		Report:     NewReport(r, tasks, q.DeadLetters()),
	}
}

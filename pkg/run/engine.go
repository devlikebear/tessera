package run

import (
	"context"
	"errors"
	"sync"

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
	publisher := eventPublisher{run: r, sink: cfg.EventSink}
	if err := r.Start(cfg.Mandate); err != nil {
		_ = publisher.Emit(ctx)
		return ExecutionResult{}, err
	}
	if err := publisher.Emit(ctx); err != nil {
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
			if emitErr := publisher.Emit(ctx); emitErr != nil {
				return finishExecution(r, cfg.Queue), emitErr
			}
			return finishExecution(r, cfg.Queue), nil
		}

		dispatched, err := DispatchReady(ctx, cfg.Graph, completed, enqueued, cfg.Queue, maxAttempts)
		if err != nil {
			return ExecutionResult{}, err
		}
		for _, task := range dispatched {
			r.Events.Append(NewTaskEvent(EventTaskQueued, r.ID, task, "", queue.TaskQueued, "task queued"))
		}
		if err := publisher.Emit(ctx); err != nil {
			return finishExecution(r, cfg.Queue), err
		}

		summary, err := (executor.Pool{
			Queue:    cfg.Queue,
			Executor: cfg.Executor,
			Hooks: executor.Hooks{
				OnTaskStarted: func(ctx context.Context, lease queue.Lease) error {
					return publisher.Record(ctx, NewLeaseEvent(EventTaskStarted, r.ID, lease, queue.TaskQueued, queue.TaskRunning, "task started"))
				},
				OnTaskSucceeded: func(ctx context.Context, lease queue.Lease, result executor.Result) error {
					event := NewLeaseEvent(EventTaskSucceeded, r.ID, lease, queue.TaskRunning, queue.TaskSucceeded, "task succeeded")
					event.OutputSummary = summarizeEventText(result.Output)
					return publisher.Record(ctx, event)
				},
				OnTaskRetrying: func(ctx context.Context, lease queue.Lease, err error) error {
					event := NewLeaseEvent(EventTaskRetrying, r.ID, lease, queue.TaskRunning, queue.TaskRetrying, "task retrying")
					event.Error = err.Error()
					return publisher.Record(ctx, event)
				},
				OnTaskFailed: func(ctx context.Context, lease queue.Lease, err error) error {
					event := NewLeaseEvent(EventTaskFailed, r.ID, lease, queue.TaskRunning, queue.TaskFailed, "task failed")
					event.Error = err.Error()
					return publisher.Record(ctx, event)
				},
			},
			Workers:    cfg.Workers,
			RoleLimits: cfg.RoleLimits,
		}).RunUntilIdle(ctx)
		if err != nil {
			_ = r.Close(ClosureAbnormal, err.Error())
			if emitErr := publisher.Emit(ctx); emitErr != nil {
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
				}
			case queue.TaskFailed:
				_ = r.Close(ClosureAbnormal, "task failed: "+task.ID)
				if emitErr := publisher.Emit(ctx); emitErr != nil {
					return finishExecution(r, cfg.Queue), emitErr
				}
				return finishExecution(r, cfg.Queue), nil
			}
		}
		if err := publisher.Emit(ctx); err != nil {
			return finishExecution(r, cfg.Queue), err
		}

		if len(dispatched) == 0 && !progressed && summary.Retried == 0 {
			_ = r.Close(ClosureAbnormal, "no dispatchable tasks remain")
			if emitErr := publisher.Emit(ctx); emitErr != nil {
				return finishExecution(r, cfg.Queue), emitErr
			}
			return finishExecution(r, cfg.Queue), nil
		}
	}
}

type eventPublisher struct {
	mu          sync.Mutex
	run         *Run
	sink        EventSink
	lastEmitted int
}

func (p *eventPublisher) Record(ctx context.Context, event Event) error {
	p.run.Events.Append(event)
	return p.Emit(ctx)
}

func (p *eventPublisher) Emit(ctx context.Context) error {
	if p.sink == nil {
		return nil
	}
	p.mu.Lock()
	defer p.mu.Unlock()

	for _, event := range p.run.Events.Events() {
		if event.Seq <= p.lastEmitted {
			continue
		}
		if err := p.sink.OnEvent(ctx, event); err != nil {
			return err
		}
		p.lastEmitted = event.Seq
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

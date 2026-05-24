package run

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/devlikebear/tessera/pkg/executor"
	"github.com/devlikebear/tessera/pkg/leader"
	"github.com/devlikebear/tessera/pkg/mandate"
	"github.com/devlikebear/tessera/pkg/queue"
)

func TestExecuteTaskGraphNormalClosure(t *testing.T) {
	ctx := context.Background()
	plan, m, graph := approvedGraph(t)

	result, err := ExecuteTaskGraph(ctx, ExecutionConfig{
		RunID:       "run-1",
		Mandate:     m,
		Graph:       graph,
		Queue:       queue.NewInMemory(),
		Workers:     2,
		MaxAttempts: 2,
		Executor: executor.TaskHandler(func(ctx context.Context, task queue.Task) (executor.Result, error) {
			return executor.Result{Output: task.ID}, nil
		}),
	})
	if err != nil {
		t.Fatalf("ExecuteTaskGraph() error = %v", err)
	}
	if result.Report.Closure != ClosureNormal {
		t.Fatalf("closure = %q, want %q", result.Report.Closure, ClosureNormal)
	}
	if result.Report.Succeeded != len(plan.Steps) {
		t.Fatalf("succeeded = %d, want %d", result.Report.Succeeded, len(plan.Steps))
	}
}

func TestExecuteTaskGraphAbnormalClosureOnRetryBudget(t *testing.T) {
	ctx := context.Background()
	_, m, graph := approvedGraph(t)

	result, err := ExecuteTaskGraph(ctx, ExecutionConfig{
		RunID:       "run-1",
		Mandate:     m,
		Graph:       graph,
		Queue:       queue.NewInMemory(),
		Workers:     1,
		MaxAttempts: 1,
		Executor: executor.TaskHandler(func(ctx context.Context, task queue.Task) (executor.Result, error) {
			return executor.Result{}, errors.New("permanent")
		}),
	})
	if err != nil {
		t.Fatalf("ExecuteTaskGraph() error = %v", err)
	}
	if result.Report.Closure != ClosureAbnormal {
		t.Fatalf("closure = %q, want %q", result.Report.Closure, ClosureAbnormal)
	}
	if len(result.Report.DeadLetters) == 0 {
		t.Fatal("expected dead letters")
	}
}

func TestExecuteTaskGraphEmitsDetailedStreamingEventsToSink(t *testing.T) {
	ctx := context.Background()
	_, m, graph := approvedGraph(t)
	var events []Event

	result, err := ExecuteTaskGraph(ctx, ExecutionConfig{
		RunID:       "run-1",
		Mandate:     m,
		Graph:       graph,
		Queue:       queue.NewInMemory(),
		Workers:     2,
		MaxAttempts: 2,
		EventSink: EventSinkFunc(func(_ context.Context, event Event) error {
			events = append(events, event)
			return nil
		}),
		Executor: executor.TaskHandler(func(ctx context.Context, task queue.Task) (executor.Result, error) {
			return executor.Result{Output: "completed " + task.ID}, nil
		}),
	})
	if err != nil {
		t.Fatalf("ExecuteTaskGraph() error = %v", err)
	}
	if result.Report.Closure != ClosureNormal {
		t.Fatalf("closure = %q, want %q", result.Report.Closure, ClosureNormal)
	}

	var sawRunning, sawWriterQueued, sawWriterStarted, sawWriterSucceeded, sawClosure bool
	for _, event := range events {
		if event.SchemaVersion != EventSchemaVersion || event.Seq == 0 || event.At.IsZero() {
			t.Fatalf("event missing sequence or timestamp: %+v", event)
		}
		if event.Type == EventRunTransition && event.To == string(StatusRunning) {
			sawRunning = true
		}
		if event.Type == EventTaskQueued && event.TaskID == "draft-chapter" && event.Role == "writer" && event.Stage == "execution" && event.To == string(queue.TaskQueued) {
			sawWriterQueued = true
		}
		if event.Type == EventTaskStarted && event.TaskID == "draft-chapter" && event.Attempt == 1 && event.MaxAttempts == 2 && event.WorkerID != "" && event.InputSummary != "" {
			sawWriterStarted = true
		}
		if event.Type == EventTaskSucceeded && event.TaskID == "draft-chapter" && event.OutputSummary == "completed draft-chapter" {
			sawWriterSucceeded = true
		}
		if event.Type == EventClosure && event.To == string(ClosureNormal) {
			sawClosure = true
		}
	}
	if !sawRunning || !sawWriterQueued || !sawWriterStarted || !sawWriterSucceeded || !sawClosure {
		t.Fatalf("events = %+v, want running transition, detailed writer lifecycle, and normal closure", events)
	}
}

func TestExecuteTaskGraphEmitsRetryAndFailureDetails(t *testing.T) {
	ctx := context.Background()
	_, m, graph := approvedGraph(t)
	var events []Event
	failures := 0

	result, err := ExecuteTaskGraph(ctx, ExecutionConfig{
		RunID:       "run-1",
		Mandate:     m,
		Graph:       graph,
		Queue:       queue.NewInMemory(),
		Workers:     1,
		MaxAttempts: 2,
		EventSink: EventSinkFunc(func(_ context.Context, event Event) error {
			events = append(events, event)
			return nil
		}),
		Executor: executor.TaskHandler(func(ctx context.Context, task queue.Task) (executor.Result, error) {
			if task.ID == "research-world" && failures == 0 {
				failures++
				return executor.Result{}, errors.New("temporary research outage")
			}
			return executor.Result{Output: "completed " + task.ID}, nil
		}),
	})
	if err != nil {
		t.Fatalf("ExecuteTaskGraph() error = %v", err)
	}
	if result.Report.Closure != ClosureNormal {
		t.Fatalf("closure = %q, want %q", result.Report.Closure, ClosureNormal)
	}

	var sawRetrying, sawSecondAttempt bool
	for _, event := range events {
		if event.TaskID != "research-world" {
			continue
		}
		if event.Type == EventTaskRetrying && event.Attempt == 1 && event.MaxAttempts == 2 && event.Error == "temporary research outage" {
			sawRetrying = true
		}
		if event.Type == EventTaskStarted && event.Attempt == 2 {
			sawSecondAttempt = true
		}
	}
	if !sawRetrying || !sawSecondAttempt {
		t.Fatalf("events = %+v, want retrying event and second attempt start", events)
	}
}

func approvedGraph(t *testing.T) (leader.Plan, mandate.Mandate, TaskGraph) {
	t.Helper()
	plan, err := leader.NewNovelTeamPlanner().Plan(context.Background(), leader.Goal{
		ID:   "goal-1",
		Text: "Draft a hopeful climate fiction opening",
	})
	if err != nil {
		t.Fatalf("Plan() error = %v", err)
	}
	m, err := mandate.New("mandate-1", plan.Goal.Text, plan.Summary).
		Approve("operator", time.Date(2026, 5, 24, 12, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("Approve() error = %v", err)
	}
	graph, err := NewTaskGraphFromPlan(plan, m)
	if err != nil {
		t.Fatalf("NewTaskGraphFromPlan() error = %v", err)
	}
	return plan, m, graph
}

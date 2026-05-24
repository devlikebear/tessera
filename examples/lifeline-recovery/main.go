package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/devlikebear/tessera/pkg/executor"
	"github.com/devlikebear/tessera/pkg/leader"
	"github.com/devlikebear/tessera/pkg/lifeline"
	"github.com/devlikebear/tessera/pkg/mandate"
	"github.com/devlikebear/tessera/pkg/queue"
	"github.com/devlikebear/tessera/pkg/run"
)

type appReport struct {
	App       string          `json:"app"`
	Closure   run.ClosureKind `json:"closure"`
	Recovered int             `json:"recovered"`
	Succeeded int             `json:"succeeded"`
	Failed    int             `json:"failed"`
}

func main() {
	if err := runApp(context.Background(), os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func runApp(ctx context.Context, out io.Writer) error {
	base := time.Date(2026, 5, 24, 13, 30, 0, 0, time.UTC)
	plan, err := leader.NewStaticPlanner("lifeline-recovery-plan", []leader.Step{
		{
			ID:          "recover-stalled-work",
			Title:       "Recover stalled work",
			Role:        "rescuer",
			Stage:       "recovery",
			Description: "Complete a task after its original worker stops heartbeating.",
		},
	}).Plan(ctx, leader.Goal{
		ID:       "lifeline-recovery-goal",
		Text:     "Demonstrate queue recovery for an expired in-flight task.",
		Template: "lifeline-recovery",
	})
	if err != nil {
		return err
	}
	m, err := mandate.New("lifeline-recovery-mandate", plan.Goal.Text, plan.Summary).
		Approve("example-operator", base)
	if err != nil {
		return err
	}
	graph, err := run.NewTaskGraphFromPlan(plan, m)
	if err != nil {
		return err
	}

	r := run.New("lifeline-recovery-run")
	if err := r.Start(m); err != nil {
		return err
	}

	q := queue.NewInMemory(
		queue.WithLeaseTimeout(time.Second),
		queue.WithClock(func() time.Time { return base }),
	)
	completed := make(map[string]struct{})
	enqueued := make(map[string]struct{})
	if _, err := run.DispatchReady(ctx, graph, completed, enqueued, q, 2); err != nil {
		return err
	}

	if _, err := q.Dequeue(ctx, "stalled-worker"); err != nil {
		return err
	}
	recovered := lifeline.Watchdog{
		Queue: q,
		Now:   func() time.Time { return base.Add(2 * time.Second) },
	}.RecoverStalled(ctx)

	summary, err := (executor.Pool{
		Queue:   q,
		Workers: 1,
		Executor: executor.TaskHandler(func(ctx context.Context, task queue.Task) (executor.Result, error) {
			return executor.Result{Output: "recovered " + task.ID}, nil
		}),
	}).RunUntilIdle(ctx)
	if err != nil {
		_ = r.Close(run.ClosureAbnormal, err.Error())
		return err
	}
	if summary.Succeeded == len(graph.Nodes) {
		if err := r.Close(run.ClosureNormal, "stalled task recovered"); err != nil {
			return err
		}
	} else {
		if err := r.Close(run.ClosureAbnormal, "recovery did not complete every task"); err != nil {
			return err
		}
	}

	report := run.NewReport(r, q.Snapshot(), q.DeadLetters())
	enc := json.NewEncoder(out)
	enc.SetIndent("", "  ")
	return enc.Encode(appReport{
		App:       "lifeline-recovery",
		Closure:   report.Closure,
		Recovered: len(recovered),
		Succeeded: report.Succeeded,
		Failed:    report.Failed,
	})
}

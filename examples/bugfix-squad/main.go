package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"github.com/devlikebear/tessera/pkg/council"
	"github.com/devlikebear/tessera/pkg/executor"
	"github.com/devlikebear/tessera/pkg/leader"
	"github.com/devlikebear/tessera/pkg/mandate"
	"github.com/devlikebear/tessera/pkg/queue"
	"github.com/devlikebear/tessera/pkg/run"
)

type appReport struct {
	App              string            `json:"app"`
	Closure          run.ClosureKind   `json:"closure"`
	Succeeded        int               `json:"succeeded"`
	TransientRetries int               `json:"transient_retries"`
	Artifacts        map[string]string `json:"artifacts"`
}

func main() {
	if err := runApp(context.Background(), os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func runApp(ctx context.Context, out io.Writer) error {
	plan, err := leader.NewStaticPlanner("bugfix-squad-plan", []leader.Step{
		{
			ID:          "reproduce-failure",
			Title:       "Reproduce failure",
			Role:        "tester",
			Stage:       "diagnosis",
			Description: "Reproduce a failing regression test and capture the symptom.",
		},
		{
			ID:          "patch-fix",
			Title:       "Patch fix",
			Role:        "engineer",
			Stage:       "execution",
			Description: "Apply a small fix for the reproduced failure.",
			DependsOn:   []string{"reproduce-failure"},
		},
		{
			ID:          "run-tests",
			Title:       "Run tests",
			Role:        "tester",
			Stage:       "verification",
			Description: "Run focused tests for the bugfix.",
			DependsOn:   []string{"patch-fix"},
		},
		{
			ID:          "summarize-fix",
			Title:       "Summarize fix",
			Role:        "lead",
			Stage:       "closure",
			Description: "Produce a concise fix report.",
			DependsOn:   []string{"run-tests"},
		},
	}).Plan(ctx, leader.Goal{
		ID:       "bugfix-squad-goal",
		Text:     "Fix a regression through a delegated bugfix cell.",
		Template: "bugfix-squad",
	})
	if err != nil {
		return err
	}

	councilReport, err := council.DefaultCouncil().Review(ctx, plan)
	if err != nil {
		return err
	}
	m, err := mandate.New("bugfix-squad-mandate", plan.Goal.Text, plan.Summary).
		WithReviews(councilReport.MandateReviews()).
		Approve("example-operator", time.Date(2026, 5, 24, 13, 0, 0, 0, time.UTC))
	if err != nil {
		return err
	}
	graph, err := run.NewTaskGraphFromPlan(plan, m)
	if err != nil {
		return err
	}

	artifacts := newArtifactStore()
	var patchAttempts atomic.Int32
	var transientRetries atomic.Int32
	result, err := run.ExecuteTaskGraph(ctx, run.ExecutionConfig{
		RunID:       "bugfix-squad-run",
		Mandate:     m,
		Graph:       graph,
		Queue:       queue.NewInMemory(),
		Workers:     2,
		MaxAttempts: 2,
		RoleLimits: map[string]int{
			"engineer": 1,
			"tester":   2,
			"lead":     1,
		},
		Executor: executor.TaskHandler(func(ctx context.Context, task queue.Task) (executor.Result, error) {
			if task.ID == "patch-fix" && patchAttempts.Add(1) == 1 {
				transientRetries.Add(1)
				return executor.Result{}, errors.New("simulated flaky patch application")
			}
			output := fmt.Sprintf("%s finished %s", task.Role, task.ID)
			artifacts.set(task.ID, output)
			return executor.Result{Output: output}, nil
		}),
	})
	if err != nil {
		return err
	}

	enc := json.NewEncoder(out)
	enc.SetIndent("", "  ")
	return enc.Encode(appReport{
		App:              "bugfix-squad",
		Closure:          result.Report.Closure,
		Succeeded:        result.Report.Succeeded,
		TransientRetries: int(transientRetries.Load()),
		Artifacts:        artifacts.snapshot(),
	})
}

type artifactStore struct {
	mu    sync.Mutex
	items map[string]string
}

func newArtifactStore() *artifactStore {
	return &artifactStore{items: make(map[string]string)}
}

func (s *artifactStore) set(id, value string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.items[id] = value
}

func (s *artifactStore) snapshot() map[string]string {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make(map[string]string, len(s.items))
	for key, value := range s.items {
		out[key] = value
	}
	return out
}

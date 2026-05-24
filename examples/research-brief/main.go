package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"github.com/devlikebear/tessera/pkg/council"
	"github.com/devlikebear/tessera/pkg/executor"
	"github.com/devlikebear/tessera/pkg/leader"
	"github.com/devlikebear/tessera/pkg/mandate"
	"github.com/devlikebear/tessera/pkg/queue"
	"github.com/devlikebear/tessera/pkg/run"
)

type appReport struct {
	App        string            `json:"app"`
	Closure    run.ClosureKind   `json:"closure"`
	Succeeded  int               `json:"succeeded"`
	Artifacts  map[string]string `json:"artifacts"`
	EventCount int               `json:"event_count"`
}

func main() {
	if err := runApp(context.Background(), os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func runApp(ctx context.Context, out io.Writer) error {
	plan, err := leader.NewStaticPlanner("research-brief-plan", []leader.Step{
		{
			ID:          "collect-notes",
			Title:       "Collect notes",
			Role:        "researcher",
			Stage:       "research",
			Description: "Collect key facts for a concise Tessera adoption brief.",
		},
		{
			ID:          "synthesize-angle",
			Title:       "Synthesize angle",
			Role:        "analyst",
			Stage:       "planning",
			Description: "Turn notes into the brief's central argument.",
			DependsOn:   []string{"collect-notes"},
		},
		{
			ID:          "draft-brief",
			Title:       "Draft brief",
			Role:        "writer",
			Stage:       "execution",
			Description: "Draft a short research brief from the synthesized angle.",
			DependsOn:   []string{"synthesize-angle"},
		},
		{
			ID:          "review-brief",
			Title:       "Review brief",
			Role:        "editor",
			Stage:       "closure",
			Description: "Review the brief and produce final notes.",
			DependsOn:   []string{"draft-brief"},
		},
	}).Plan(ctx, leader.Goal{
		ID:       "research-brief-goal",
		Text:     "Create a concise research brief about using Tessera for delegated agent teams.",
		Template: "research-brief",
	})
	if err != nil {
		return err
	}

	councilReport, err := council.DefaultCouncil().Review(ctx, plan)
	if err != nil {
		return err
	}
	m, err := mandate.New("research-brief-mandate", plan.Goal.Text, plan.Summary).
		WithReviews(councilReport.MandateReviews()).
		Approve("example-operator", time.Date(2026, 5, 24, 12, 30, 0, 0, time.UTC))
	if err != nil {
		return err
	}
	graph, err := run.NewTaskGraphFromPlan(plan, m)
	if err != nil {
		return err
	}

	artifacts := newArtifactStore()
	result, err := run.ExecuteTaskGraph(ctx, run.ExecutionConfig{
		RunID:       "research-brief-run",
		Mandate:     m,
		Graph:       graph,
		Queue:       queue.NewInMemory(),
		Workers:     3,
		MaxAttempts: 2,
		RoleLimits: map[string]int{
			"researcher": 1,
			"analyst":    1,
			"writer":     2,
			"editor":     1,
		},
		Executor: executor.TaskHandler(func(ctx context.Context, task queue.Task) (executor.Result, error) {
			output := fmt.Sprintf("%s completed: %s", task.Role, task.Payload)
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
		App:        "research-brief",
		Closure:    result.Report.Closure,
		Succeeded:  result.Report.Succeeded,
		Artifacts:  artifacts.snapshot(),
		EventCount: len(result.Report.Events),
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

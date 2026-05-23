package run

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/devlikebear/tessera/pkg/leader"
	"github.com/devlikebear/tessera/pkg/mandate"
)

func TestTaskGraphRequiresApprovedMandate(t *testing.T) {
	plan, err := leader.NewNovelTeamPlanner().Plan(context.Background(), leader.Goal{
		ID:   "goal-1",
		Text: "Draft a hopeful climate fiction opening",
	})
	if err != nil {
		t.Fatalf("Plan() error = %v", err)
	}

	_, err = NewTaskGraphFromPlan(plan, mandate.New("mandate-1", plan.Goal.Text, plan.Summary))
	if !errors.Is(err, mandate.ErrNotApproved) {
		t.Fatalf("NewTaskGraphFromPlan() error = %v, want ErrNotApproved", err)
	}
}

func TestTaskGraphPreservesPlanDependencies(t *testing.T) {
	plan, err := leader.NewNovelTeamPlanner().Plan(context.Background(), leader.Goal{
		ID:   "goal-1",
		Text: "Draft a hopeful climate fiction opening",
	})
	if err != nil {
		t.Fatalf("Plan() error = %v", err)
	}
	m, err := mandate.New("mandate-1", plan.Goal.Text, plan.Summary).
		Approve("operator", time.Date(2026, 5, 24, 10, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("Approve() error = %v", err)
	}

	graph, err := NewTaskGraphFromPlan(plan, m)
	if err != nil {
		t.Fatalf("NewTaskGraphFromPlan() error = %v", err)
	}
	if len(graph.Nodes) != len(plan.Steps) {
		t.Fatalf("nodes len = %d, want %d", len(graph.Nodes), len(plan.Steps))
	}
	if graph.Nodes[1].DependsOn[0] != "research-world" {
		t.Fatalf("dependency = %q, want research-world", graph.Nodes[1].DependsOn[0])
	}
}

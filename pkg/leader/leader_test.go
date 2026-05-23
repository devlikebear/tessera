package leader

import (
	"context"
	"testing"
)

func TestNovelTeamPlannerBuildsDeterministicPlan(t *testing.T) {
	plan, err := NewNovelTeamPlanner().Plan(context.Background(), Goal{
		ID:       "goal-1",
		Text:     "Draft a hopeful climate fiction opening",
		Template: "novel-team",
	})
	if err != nil {
		t.Fatalf("Plan() error = %v", err)
	}

	if plan.ID != "novel-team-plan" {
		t.Fatalf("Plan ID = %q, want novel-team-plan", plan.ID)
	}
	if len(plan.Steps) != 4 {
		t.Fatalf("steps len = %d, want 4", len(plan.Steps))
	}
	if plan.Steps[1].DependsOn[0] != "research-world" {
		t.Fatalf("outline dependency = %q, want research-world", plan.Steps[1].DependsOn[0])
	}
}

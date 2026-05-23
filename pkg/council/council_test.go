package council

import (
	"context"
	"testing"

	"github.com/devlikebear/tessera/pkg/leader"
)

func TestDefaultCouncilReviewsAllFourRoles(t *testing.T) {
	plan, err := leader.NewNovelTeamPlanner().Plan(context.Background(), leader.Goal{
		ID:   "goal-1",
		Text: "Draft a hopeful climate fiction opening",
	})
	if err != nil {
		t.Fatalf("Plan() error = %v", err)
	}

	report, err := DefaultCouncil().Review(context.Background(), plan)
	if err != nil {
		t.Fatalf("Review() error = %v", err)
	}

	for _, role := range []Role{RolePositive, RoleNegative, RoleMediator, RoleResearcher} {
		if !report.HasRole(role) {
			t.Fatalf("report missing role %q", role)
		}
	}
	if len(report.MandateReviews()) != 4 {
		t.Fatalf("mandate review len = %d, want 4", len(report.MandateReviews()))
	}
}

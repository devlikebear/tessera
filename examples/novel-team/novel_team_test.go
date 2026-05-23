package novelteam

import (
	"context"
	"testing"
	"time"
)

func TestBuildNovelTeamFlow(t *testing.T) {
	result, err := Build(
		context.Background(),
		"Draft a hopeful climate fiction opening",
		"operator",
		time.Date(2026, 5, 24, 10, 0, 0, 0, time.UTC),
	)
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}

	if len(result.Report.Reviews) != 4 {
		t.Fatalf("reviews len = %d, want 4", len(result.Report.Reviews))
	}
	if !result.Mandate.IsApproved() {
		t.Fatal("mandate should be approved")
	}
	if len(result.Graph.Nodes) != 4 {
		t.Fatalf("graph nodes len = %d, want 4", len(result.Graph.Nodes))
	}
}

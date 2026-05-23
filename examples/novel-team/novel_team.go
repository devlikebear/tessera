package novelteam

import (
	"context"
	"time"

	"github.com/devlikebear/tessera/pkg/council"
	"github.com/devlikebear/tessera/pkg/leader"
	"github.com/devlikebear/tessera/pkg/mandate"
	"github.com/devlikebear/tessera/pkg/run"
)

type Result struct {
	Plan    leader.Plan
	Report  council.Report
	Mandate mandate.Mandate
	Graph   run.TaskGraph
}

func Build(ctx context.Context, goalText, approvedBy string, approvedAt time.Time) (Result, error) {
	plan, err := leader.NewNovelTeamPlanner().Plan(ctx, leader.Goal{
		ID:       "novel-team-goal",
		Text:     goalText,
		Template: "novel-team",
	})
	if err != nil {
		return Result{}, err
	}
	report, err := council.DefaultCouncil().Review(ctx, plan)
	if err != nil {
		return Result{}, err
	}
	m := mandate.New("novel-team-mandate", plan.Goal.Text, plan.Summary).
		WithReviews(report.MandateReviews())
	m, err = m.Approve(approvedBy, approvedAt)
	if err != nil {
		return Result{}, err
	}
	graph, err := run.NewTaskGraphFromPlan(plan, m)
	if err != nil {
		return Result{}, err
	}
	return Result{Plan: plan, Report: report, Mandate: m, Graph: graph}, nil
}

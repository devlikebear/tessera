package leader

import (
	"context"
	"errors"
	"strings"
)

var ErrInvalidGoal = errors.New("invalid goal")

type Goal struct {
	ID       string
	Text     string
	Template string
}

type Step struct {
	ID          string
	Title       string
	Role        string
	Stage       string
	Description string
	DependsOn   []string
}

type Plan struct {
	ID      string
	Goal    Goal
	Summary string
	Steps   []Step
}

type Planner interface {
	Plan(context.Context, Goal) (Plan, error)
}

type StaticPlanner struct {
	planID string
	steps  []Step
}

func NewStaticPlanner(planID string, steps []Step) StaticPlanner {
	return StaticPlanner{
		planID: strings.TrimSpace(planID),
		steps:  cloneSteps(steps),
	}
}

func NewNovelTeamPlanner() StaticPlanner {
	return NewStaticPlanner("novel-team-plan", []Step{
		{
			ID:          "research-world",
			Title:       "Research the story world",
			Role:        "researcher",
			Stage:       "planning",
			Description: "Collect grounded constraints and inspiration for the story world.",
		},
		{
			ID:          "outline-plot",
			Title:       "Outline the plot",
			Role:        "leader",
			Stage:       "planning",
			Description: "Turn the goal and research into a concise plot outline.",
			DependsOn:   []string{"research-world"},
		},
		{
			ID:          "draft-chapter",
			Title:       "Draft the first chapter",
			Role:        "writer",
			Stage:       "execution",
			Description: "Write a first-pass chapter from the approved outline.",
			DependsOn:   []string{"outline-plot"},
		},
		{
			ID:          "review-draft",
			Title:       "Review the draft",
			Role:        "editor",
			Stage:       "closure",
			Description: "Review the draft and produce final notes.",
			DependsOn:   []string{"draft-chapter"},
		},
	})
}

func (p StaticPlanner) Plan(_ context.Context, goal Goal) (Plan, error) {
	goal.ID = strings.TrimSpace(goal.ID)
	goal.Text = strings.TrimSpace(goal.Text)
	goal.Template = strings.TrimSpace(goal.Template)
	if goal.ID == "" || goal.Text == "" {
		return Plan{}, ErrInvalidGoal
	}
	planID := p.planID
	if planID == "" {
		planID = goal.ID + "-plan"
	}
	return Plan{
		ID:      planID,
		Goal:    goal,
		Summary: "Delegate goal through a mandate-gated Tessera execution cell.",
		Steps:   cloneSteps(p.steps),
	}, nil
}

func cloneSteps(steps []Step) []Step {
	out := make([]Step, len(steps))
	for i := range steps {
		out[i] = steps[i]
		out[i].DependsOn = append([]string(nil), steps[i].DependsOn...)
	}
	return out
}

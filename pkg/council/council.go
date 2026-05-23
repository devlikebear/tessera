package council

import (
	"context"
	"errors"

	"github.com/devlikebear/tessera/pkg/leader"
	"github.com/devlikebear/tessera/pkg/mandate"
)

var ErrInvalidPlan = errors.New("invalid plan")

type Role string

const (
	RolePositive   Role = "positive"
	RoleNegative   Role = "negative"
	RoleMediator   Role = "mediator"
	RoleResearcher Role = "researcher"
)

type Verdict string

const (
	VerdictApprove Verdict = "approve"
	VerdictRevise  Verdict = "revise"
)

type Review struct {
	Role    Role
	Verdict Verdict
	Notes   string
}

type Reviewer interface {
	Review(context.Context, leader.Plan) (Review, error)
}

type Council struct {
	reviewers []Reviewer
}

type Report struct {
	PlanID  string
	Reviews []Review
}

func DefaultCouncil() Council {
	return Council{reviewers: []Reviewer{
		staticReviewer{role: RolePositive, notes: "The plan has a clear mandate boundary and concrete execution stages."},
		staticReviewer{role: RoleNegative, notes: "Watch retry budgets and closure rules so recovery does not loop forever."},
		staticReviewer{role: RoleMediator, notes: "The plan is acceptable if approval gates remain mandatory."},
		staticReviewer{role: RoleResearcher, notes: "The task graph preserves inspectable dependencies for later evidence gathering."},
	}}
}

func (c Council) Review(ctx context.Context, plan leader.Plan) (Report, error) {
	if plan.ID == "" || len(plan.Steps) == 0 {
		return Report{}, ErrInvalidPlan
	}
	report := Report{PlanID: plan.ID}
	for _, reviewer := range c.reviewers {
		review, err := reviewer.Review(ctx, plan)
		if err != nil {
			return Report{}, err
		}
		report.Reviews = append(report.Reviews, review)
	}
	return report, nil
}

func (r Report) HasRole(role Role) bool {
	for _, review := range r.Reviews {
		if review.Role == role {
			return true
		}
	}
	return false
}

func (r Report) MandateReviews() []mandate.Review {
	reviews := make([]mandate.Review, len(r.Reviews))
	for i, review := range r.Reviews {
		reviews[i] = mandate.Review{
			Role:    string(review.Role),
			Verdict: string(review.Verdict),
			Notes:   review.Notes,
		}
	}
	return reviews
}

type staticReviewer struct {
	role  Role
	notes string
}

func (r staticReviewer) Review(_ context.Context, _ leader.Plan) (Review, error) {
	return Review{Role: r.role, Verdict: VerdictApprove, Notes: r.notes}, nil
}

package run

import (
	"errors"

	"github.com/devlikebear/tessera/pkg/leader"
	"github.com/devlikebear/tessera/pkg/mandate"
	"github.com/devlikebear/tessera/pkg/queue"
)

var ErrInvalidTaskGraph = errors.New("invalid task graph")

type TaskNode struct {
	ID          string
	Role        string
	Stage       string
	Description string
	DependsOn   []string
	Status      queue.TaskStatus
}

type TaskGraph struct {
	PlanID string
	Nodes  []TaskNode
}

func NewTaskGraphFromPlan(plan leader.Plan, m mandate.Mandate) (TaskGraph, error) {
	if err := m.RequireApproved(); err != nil {
		return TaskGraph{}, err
	}
	if plan.ID == "" || len(plan.Steps) == 0 {
		return TaskGraph{}, ErrInvalidTaskGraph
	}

	graph := TaskGraph{PlanID: plan.ID}
	for _, step := range plan.Steps {
		if step.ID == "" || step.Role == "" || step.Stage == "" {
			return TaskGraph{}, ErrInvalidTaskGraph
		}
		graph.Nodes = append(graph.Nodes, TaskNode{
			ID:          step.ID,
			Role:        step.Role,
			Stage:       step.Stage,
			Description: step.Description,
			DependsOn:   append([]string(nil), step.DependsOn...),
			Status:      queue.TaskQueued,
		})
	}
	return graph, nil
}

func (g TaskGraph) TaskIDs() map[string]struct{} {
	ids := make(map[string]struct{}, len(g.Nodes))
	for _, node := range g.Nodes {
		ids[node.ID] = struct{}{}
	}
	return ids
}

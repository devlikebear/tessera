package tarsagent

import (
	"context"
	"errors"

	"github.com/devlikebear/tars/pkg/agentloop"
	"github.com/devlikebear/tars/pkg/llm"
	"github.com/devlikebear/tars/pkg/tools"
	"github.com/devlikebear/tessera/pkg/executor"
	"github.com/devlikebear/tessera/pkg/queue"
)

var ErrMissingClient = errors.New("tars agentloop executor requires an llm client")

type AgentLoopExecutor struct {
	Client        llm.Client
	Registry      *tools.Registry
	MaxIterations int
}

func New(client llm.Client, registry *tools.Registry) AgentLoopExecutor {
	return AgentLoopExecutor{Client: client, Registry: registry}
}

func (e AgentLoopExecutor) Execute(ctx context.Context, task queue.Task) (executor.Result, error) {
	if e.Client == nil {
		return executor.Result{}, ErrMissingClient
	}
	registry := e.Registry
	if registry == nil {
		registry = tools.NewRegistry()
	}
	maxIterations := e.MaxIterations
	if maxIterations <= 0 {
		maxIterations = 1
	}
	loop := agentloop.New(e.Client, registry)
	resp, err := loop.Run(ctx, []llm.ChatMessage{
		{Role: "user", Content: task.Payload},
	}, agentloop.RunOptions{
		MaxIterations: maxIterations,
		ToolChoice:    llm.ToolChoiceNone(),
	})
	if err != nil {
		return executor.Result{}, err
	}
	return executor.Result{Output: resp.Message.Content}, nil
}

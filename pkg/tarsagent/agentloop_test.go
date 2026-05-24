package tarsagent

import (
	"context"
	"testing"

	"github.com/devlikebear/tars/pkg/llm"
	"github.com/devlikebear/tars/pkg/tools"
	"github.com/devlikebear/tessera/pkg/queue"
)

func TestAgentLoopExecutorUsesPublicTARSPackages(t *testing.T) {
	exec := New(fakeClient{}, tools.NewRegistry())
	result, err := exec.Execute(context.Background(), queue.Task{
		ID:      "task-1",
		Role:    "writer",
		Payload: "write one paragraph",
	})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if result.Output != "done: write one paragraph" {
		t.Fatalf("output = %q, want done response", result.Output)
	}
}

type fakeClient struct{}

func (fakeClient) Ask(ctx context.Context, prompt string) (string, error) {
	return prompt, nil
}

func (fakeClient) Chat(ctx context.Context, messages []llm.ChatMessage, opts llm.ChatOptions) (llm.ChatResponse, error) {
	return llm.ChatResponse{
		Message: llm.ChatMessage{
			Role:    "assistant",
			Content: "done: " + messages[0].Content,
		},
	}, nil
}

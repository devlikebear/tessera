package main

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"

	"github.com/devlikebear/tessera/pkg/run"
)

func TestResearchBriefAppRunsToNormalClosure(t *testing.T) {
	var out bytes.Buffer
	if err := runApp(context.Background(), &out); err != nil {
		t.Fatalf("runApp() error = %v", err)
	}

	var report appReport
	if err := json.Unmarshal(out.Bytes(), &report); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	if report.App != "research-brief" {
		t.Fatalf("app = %q, want research-brief", report.App)
	}
	if report.Closure != run.ClosureNormal {
		t.Fatalf("closure = %q, want %q", report.Closure, run.ClosureNormal)
	}
	if report.Succeeded != 4 {
		t.Fatalf("succeeded = %d, want 4", report.Succeeded)
	}
	if report.Artifacts["draft-brief"] == "" {
		t.Fatal("expected draft-brief artifact")
	}
}

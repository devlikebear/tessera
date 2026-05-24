package main

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"

	"github.com/devlikebear/tessera/pkg/run"
)

func TestRunMainNovelTeam(t *testing.T) {
	var out bytes.Buffer
	if err := runMain(context.Background(), []string{"--example", "novel-team", "--workers", "2"}, &out); err != nil {
		t.Fatalf("runMain() error = %v", err)
	}

	var report run.Report
	if err := json.Unmarshal(out.Bytes(), &report); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	if report.Closure != run.ClosureNormal {
		t.Fatalf("closure = %q, want %q", report.Closure, run.ClosureNormal)
	}
	if report.Succeeded != 4 {
		t.Fatalf("succeeded = %d, want 4", report.Succeeded)
	}
}

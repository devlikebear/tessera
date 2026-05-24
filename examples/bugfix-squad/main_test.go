package main

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"

	"github.com/devlikebear/tessera/pkg/run"
)

func TestBugfixSquadAppDemonstratesRetryRecovery(t *testing.T) {
	var out bytes.Buffer
	if err := runApp(context.Background(), &out); err != nil {
		t.Fatalf("runApp() error = %v", err)
	}

	var report appReport
	if err := json.Unmarshal(out.Bytes(), &report); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	if report.App != "bugfix-squad" {
		t.Fatalf("app = %q, want bugfix-squad", report.App)
	}
	if report.Closure != run.ClosureNormal {
		t.Fatalf("closure = %q, want %q", report.Closure, run.ClosureNormal)
	}
	if report.Succeeded != 4 {
		t.Fatalf("succeeded = %d, want 4", report.Succeeded)
	}
	if report.TransientRetries != 1 {
		t.Fatalf("transient retries = %d, want 1", report.TransientRetries)
	}
}

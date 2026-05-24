package main

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"

	"github.com/devlikebear/tessera/pkg/run"
)

func TestLifelineRecoveryAppRequeuesStalledWork(t *testing.T) {
	var out bytes.Buffer
	if err := runApp(context.Background(), &out); err != nil {
		t.Fatalf("runApp() error = %v", err)
	}

	var report appReport
	if err := json.Unmarshal(out.Bytes(), &report); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	if report.App != "lifeline-recovery" {
		t.Fatalf("app = %q, want lifeline-recovery", report.App)
	}
	if report.Closure != run.ClosureNormal {
		t.Fatalf("closure = %q, want %q", report.Closure, run.ClosureNormal)
	}
	if report.Recovered != 1 {
		t.Fatalf("recovered = %d, want 1", report.Recovered)
	}
	if report.Succeeded != 1 {
		t.Fatalf("succeeded = %d, want 1", report.Succeeded)
	}
}

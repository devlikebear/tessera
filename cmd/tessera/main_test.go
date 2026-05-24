package main

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
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

func TestRunMainWritesEventsAndVisualizesReport(t *testing.T) {
	dir := t.TempDir()
	eventsPath := filepath.Join(dir, "events.jsonl")
	reportPath := filepath.Join(dir, "report.html")

	var out bytes.Buffer
	if err := runMain(context.Background(), []string{"--example", "novel-team", "--events-out", eventsPath}, &out); err != nil {
		t.Fatalf("runMain() with events-out error = %v", err)
	}
	events, err := os.ReadFile(eventsPath)
	if err != nil {
		t.Fatalf("ReadFile(events) error = %v", err)
	}
	if !strings.Contains(string(events), `"type":"run.closure"`) {
		t.Fatalf("events file missing closure event:\n%s", events)
	}

	if err := runMain(context.Background(), []string{"visualize", eventsPath, "--out", reportPath}, &out); err != nil {
		t.Fatalf("runMain() visualize error = %v", err)
	}
	html, err := os.ReadFile(reportPath)
	if err != nil {
		t.Fatalf("ReadFile(report) error = %v", err)
	}
	for _, want := range []string{"novel-team-run", "normal", "draft-chapter", "writer"} {
		if !strings.Contains(string(html), want) {
			t.Fatalf("report missing %q:\n%s", want, html)
		}
	}
}

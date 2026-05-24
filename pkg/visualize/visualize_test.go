package visualize

import (
	"bytes"
	"strings"
	"testing"

	"github.com/devlikebear/tessera/pkg/run"
)

func TestReadEventsJSONLRejectsMalformedLine(t *testing.T) {
	_, err := ReadEventsJSONL(strings.NewReader(`{"seq":1,"type":"run.transition"}
not-json
`))
	if err == nil || !strings.Contains(err.Error(), "line 2") {
		t.Fatalf("ReadEventsJSONL() error = %v, want line 2 parse error", err)
	}
}

func TestProjectEventsTracksClosureTasksAndRoles(t *testing.T) {
	events := []run.Event{
		{Seq: 1, Type: run.EventRunTransition, RunID: "run-1", To: string(run.StatusRunning)},
		{Seq: 2, Type: run.EventTaskQueued, RunID: "run-1", TaskID: "draft-chapter", Role: "writer", To: "queued"},
		{Seq: 3, Type: run.EventTaskSucceeded, RunID: "run-1", TaskID: "draft-chapter", Role: "writer", To: "succeeded"},
		{Seq: 4, Type: run.EventClosure, RunID: "run-1", To: string(run.ClosureNormal)},
	}

	projection := Project(events)
	if projection.RunID != "run-1" {
		t.Fatalf("RunID = %q, want run-1", projection.RunID)
	}
	if projection.Closure != run.ClosureNormal {
		t.Fatalf("Closure = %q, want normal", projection.Closure)
	}
	if projection.EventCount != 4 {
		t.Fatalf("EventCount = %d, want 4", projection.EventCount)
	}
	if len(projection.Tasks) != 1 {
		t.Fatalf("tasks len = %d, want 1", len(projection.Tasks))
	}
	if projection.Tasks[0].ID != "draft-chapter" || projection.Tasks[0].Status != "succeeded" {
		t.Fatalf("task projection = %+v, want draft-chapter succeeded", projection.Tasks[0])
	}
	if len(projection.Roles) != 1 || projection.Roles[0].Role != "writer" || projection.Roles[0].Tasks != 1 {
		t.Fatalf("role projection = %+v, want one writer task", projection.Roles)
	}
}

func TestWriteHTMLReportIncludesSummaryAndTaskTable(t *testing.T) {
	projection := Project([]run.Event{
		{Seq: 1, Type: run.EventTaskSucceeded, RunID: "run-1", TaskID: "draft-chapter", Role: "writer", To: "succeeded"},
		{Seq: 2, Type: run.EventClosure, RunID: "run-1", To: string(run.ClosureNormal)},
	})

	var buf bytes.Buffer
	if err := WriteHTMLReport(&buf, projection); err != nil {
		t.Fatalf("WriteHTMLReport() error = %v", err)
	}
	html := buf.String()
	for _, want := range []string{"run-1", "normal", "draft-chapter", "writer", "succeeded"} {
		if !strings.Contains(html, want) {
			t.Fatalf("html missing %q:\n%s", want, html)
		}
	}
	if strings.Contains(html, "https://") || strings.Contains(html, "http://") {
		t.Fatalf("html should not depend on external assets:\n%s", html)
	}
}

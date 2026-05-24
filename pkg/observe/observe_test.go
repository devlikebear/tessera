package observe

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"testing"

	"github.com/devlikebear/tessera/pkg/run"
)

func TestJSONLinesSinkWritesOneEventPerLine(t *testing.T) {
	var buf bytes.Buffer
	sink := NewJSONLinesSink(&buf)

	events := []run.Event{
		{Seq: 1, Type: run.EventRunTransition, RunID: "run-1", From: string(run.StatusReady), To: string(run.StatusRunning)},
		{Seq: 2, Type: run.EventClosure, RunID: "run-1", To: string(run.ClosureNormal)},
	}
	for _, event := range events {
		if err := sink.OnEvent(context.Background(), event); err != nil {
			t.Fatalf("OnEvent() error = %v", err)
		}
	}

	scanner := bufio.NewScanner(bytes.NewReader(buf.Bytes()))
	var decoded []run.Event
	for scanner.Scan() {
		var event run.Event
		if err := json.Unmarshal(scanner.Bytes(), &event); err != nil {
			t.Fatalf("json.Unmarshal() error = %v", err)
		}
		decoded = append(decoded, event)
	}
	if err := scanner.Err(); err != nil {
		t.Fatalf("scanner error = %v", err)
	}
	if len(decoded) != 2 {
		t.Fatalf("decoded len = %d, want 2", len(decoded))
	}
	if decoded[0].Type != run.EventRunTransition || decoded[1].Type != run.EventClosure {
		t.Fatalf("decoded = %+v, want run transition then closure", decoded)
	}
}

func TestMultiSinkFanoutAndMemorySnapshot(t *testing.T) {
	var buf bytes.Buffer
	memory := &MemorySink{}
	multi := MultiSink{memory, NewJSONLinesSink(&buf)}
	event := run.Event{Seq: 1, Type: run.EventTaskTransition, RunID: "run-1", TaskID: "task-1", Role: "writer", To: "queued"}

	if err := multi.OnEvent(context.Background(), event); err != nil {
		t.Fatalf("OnEvent() error = %v", err)
	}

	events := memory.Events()
	if len(events) != 1 {
		t.Fatalf("memory events len = %d, want 1", len(events))
	}
	events[0].TaskID = "mutated"
	if got := memory.Events()[0].TaskID; got != "task-1" {
		t.Fatalf("memory snapshot was mutable, got task id %q", got)
	}
	if buf.Len() == 0 {
		t.Fatal("expected JSONL output")
	}
}

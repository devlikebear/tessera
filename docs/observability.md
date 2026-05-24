# Observability

Tessera treats visualization as an application concern. The framework streams
run events and provides a default JSONL sink plus static HTML report, but host
apps can render the same data however they want.

The customization boundary is `run.EventSink`:

```go
type EventSink interface {
    OnEvent(context.Context, run.Event) error
}
```

Apps can implement this interface to send events to SwiftUI state, a web
dashboard, SSE, WebSocket, a database, logs, or any other runtime surface.

## Built-In Flow

```bash
go run ./cmd/tessera --example novel-team --events-out .tessera/runs/novel-team/events.jsonl
go run ./cmd/tessera visualize .tessera/runs/novel-team/events.jsonl --out .tessera/runs/novel-team/report.html
```

The HTML report is a reference view. It is not the only supported UI.

## Event Types

Run lifecycle:

- `run.transition`
- `run.closure`

Task lifecycle:

- `task.queued`
- `task.started`
- `task.retrying`
- `task.succeeded`
- `task.failed`

Reserved agent-level event types:

- `agent.iteration.started`
- `agent.iteration.completed`
- `llm.request.started`
- `llm.response.completed`
- `tool.call.started`
- `tool.call.completed`

The reserved event types define the direction for deeper agent instrumentation.
They let apps plan richer visualizations without coupling to a specific UI.

## Event Fields

Each framework-emitted event includes:

- `schema_version`: event contract version, currently `1`
- `seq`: per-run event sequence
- `at`: UTC timestamp
- `type`: event type
- `run_id`: run identifier
- `task_id`: task identifier when the event is task-scoped
- `role`: agent role when available
- `stage`: task stage when available
- `attempt` and `max_attempts`: retry context
- `worker_id` and `lease_id`: execution context for started/completed tasks
- `depends_on`: task dependencies
- `from` and `to`: state transition values
- `message`: short framework message
- `input_summary`: sanitized task input summary
- `output_summary`: sanitized executor output summary
- `error`: error string for retrying or failed tasks
- `attributes`: extension map for app or integration-specific details

Tessera intentionally emits summaries rather than raw prompts, tool payloads, or
secret-bearing values by default. Apps that need deeper detail should add their
own sink or event attributes at the integration boundary where they can apply
their own redaction policy.

## Sink Failure Policy

Sink errors fail the execution path. This keeps observability loss visible
instead of silently dropping the record of what happened. Apps that prefer
best-effort telemetry can wrap their sink and return `nil` after recording the
failure elsewhere.

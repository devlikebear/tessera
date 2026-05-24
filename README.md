# Tessera

Tessera is a delegated autonomous agent-team framework that runs on top of TARS.

That sentence is the architectural baseline for code review, package boundaries,
and pull requests in this repository.

## MVP Boundary

Tessera starts as a local-first Go library and CLI. The first MVP proves:

- goal intake and leader-led planning
- planning council review
- mandate generation with an explicit approval gate
- execution-cell task graph generation
- queue-backed parallel execution
- inspectable state and event logs
- lifeline recovery for stalled or failed work
- normal or abnormal closure with a final report

## Version

Current version: `0.1.0`

## Quickstart

Run the deterministic novel-team example:

```bash
go run ./cmd/tessera --example novel-team
```

Write a JSONL event stream while the run executes:

```bash
go run ./cmd/tessera --example novel-team --events-out .tessera/runs/novel-team/events.jsonl
```

Generate a static HTML report from those events:

```bash
go run ./cmd/tessera visualize .tessera/runs/novel-team/events.jsonl --out .tessera/runs/novel-team/report.html
```

The built-in HTML report is a reference visualizer. Apps that embed Tessera can
implement `run.EventSink` and render the same event stream in their own UI.

## Observability Contract

Tessera emits versioned `run.Event` values for run and task lifecycle changes.
Task events include role, stage, attempt count, worker and lease identifiers,
dependency IDs, input/output summaries, and errors.

Use the built-in sinks in `pkg/observe` for JSONL or in-memory capture, or pass a
custom sink to `run.ExecutionConfig.EventSink` for app-specific dashboards.

See [docs/observability.md](docs/observability.md) for the event schema and
customization guidance.

## TARS Dependency Rule

Tessera may depend only on public TARS packages under:

```text
github.com/devlikebear/tars/pkg/...
```

Imports from `github.com/devlikebear/tars/internal/...` are not allowed.

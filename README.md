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

## TARS Dependency Rule

Tessera may depend only on public TARS packages under:

```text
github.com/devlikebear/tars/pkg/...
```

Imports from `github.com/devlikebear/tars/internal/...` are not allowed.

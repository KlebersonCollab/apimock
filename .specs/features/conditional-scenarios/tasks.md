# Task List: Conditional Multi-Scenarios & Rule Engine

## Sequence Guidelines (MetaGPT SOP)
- **Strict Sequential Order**: Tasks must be executed top-to-bottom without reordering or cherry-picking.
- **Atomic File Boundaries**: Each task must modify at most 1–3 specific target files.
- **Decoupled Test Setup**: Test definition / scaffolding tasks (`Type: test`) MUST precede implementation tasks (`Type: feat`).
- **Sensor Evidence Gate**: Mark complete `[x]` ONLY after passing build, lint, and test sensors with recorded evidence.

## Implementation Tasks

| Status | ID | Type | Description | Target Files | Dependencies | Evidence |
|---|---|---|---|---|---|---|
| [x] | TASK-01 | test | Scaffold unit tests for Scenario condition evaluation and multi-branch endpoint matching | `pkg/engine/scenarios_test.go` | None | Commit `71d15d4`, `go test -v ./pkg/engine` PASS |
| [x] | TASK-02 | feat | Define Scenario, Condition, Operator, and MatchMode domain structs and validation logic | `pkg/models/models.go`, `pkg/models/models_test.go` | TASK-01 | Commit `71d15d4`, `go test -v ./pkg/models` PASS |
| [x] | TASK-03 | feat | Implement predicate evaluation engine supporting query, headers, params, and JSON body paths | `pkg/engine/scenarios.go` | TASK-02 | Commit `71d15d4`, `go test -v ./pkg/engine` PASS |
| [x] | TASK-04 | feat | Integrate first-match scenario dispatching with template interpolation and fallback in engine | `pkg/engine/engine.go` | TASK-03 | Commit `71d15d4`, `go test -v ./...` PASS |
| [x] | TASK-05 | feat | Implement tabbed modal UI and visual Scenario & Condition Builder in Endpoint Studio | `web/js/views/endpoints.js`, `web/css/style.css` | TASK-04 | Commit `71d15d4`, Spec Drift OK |
| [x] | TASK-06 | feat | Add rich multi-scenario demo seeds showcasing conditional login, roles, and error states | `pkg/engine/engine.go` | TASK-05 | Commit `71d15d4`, `go test -v ./...` PASS |
| [x] | TASK-07 | review | Run full sensor suite (Go tests, Spec Drift, Go Vet, Build) and verify all BDD ACs | `pkg/engine/`, `web/` | TASK-06 | Commit `71d15d4`, 100% sensors pass |
| [x] | TASK-08 | feat | Add missing password validation scenario and quick scenario test presets to Tester View | `pkg/engine/engine.go`, `web/js/views/tester.js` | TASK-07 | Commit `49265f3`, `go test -v ./...` PASS |


## Schema Dictionary
- **Status**: `[ ]` (Pending) \| `[x]` (Verified Complete).
- **ID**: `TASK-01`, `TASK-02`, etc.
- **Type**: `test` \| `feat` \| `fix` \| `refactor` \| `docs` \| `rules` \| `skill` \| `review`.
- **Target Files**: Concrete comma-separated file paths (relative to workspace root).
- **Dependencies**: Comma-separated list of preceding task IDs or `None`.
- **Evidence**: Commit hash (`git rev-parse --short HEAD`) + sensor output snippet.

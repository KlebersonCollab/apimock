# Task List: MockForge Studio Engine (tasks.md) — MetaGPT SOP Schema Contract

# Task List: MockForge Studio Engine

## Sequence Guidelines (MetaGPT SOP)
- **Strict Sequential Order**: Tasks must be executed top-to-bottom without reordering or cherry-picking.
- **Atomic File Boundaries**: Each task must modify at most 1–3 specific target files (or declared sets).
- **Decoupled Test Setup**: Test definition / scaffolding tasks precede implementation.
- **Sensor Evidence Gate**: Mark complete `[x]` ONLY after passing build, lint, and test sensors with recorded evidence.

## Implementation Tasks

| Status | ID | Type | Description | Target Files | Dependencies | Evidence |
|---|---|---|---|---|---|---|
| [x] | TASK-01 | feat | Initialize Go module and implement domain models with validation tests | `go.mod`, `pkg/models/models.go`, `pkg/models/models_test.go` | None | cbc8d2b - PASS |
| [x] | TASK-02 | feat | Implement dynamic template interpolation and synthetic faker generator with unit tests | `pkg/template/faker.go`, `pkg/template/engine.go`, `pkg/template/engine_test.go` | TASK-01 | de31a0f - PASS |
| [x] | TASK-03 | feat | Implement Auth Guard Simulator with Bearer, API Key, Basic, and JWT token generator with unit tests | `pkg/auth/auth.go`, `pkg/auth/auth_test.go` | TASK-01 | 8815478 - PASS |
| [x] | TASK-04 | feat | Implement Stateful Auto-CRUD Collection Store with filtering, pagination, search, and unit tests | `pkg/store/store.go`, `pkg/store/store_test.go` | TASK-01 | b677f46 - PASS |
| [ ] | TASK-05 | feat | Implement Circular Buffer Traffic Inspector and SSE Broadcaster with unit tests | `pkg/traffic/traffic.go`, `pkg/traffic/traffic_test.go` | TASK-01 | Pending |
| [ ] | TASK-06 | feat | Implement OpenAPI 3.0 import/export and JSON workspace serializer with unit tests | `pkg/openapi/openapi.go`, `pkg/openapi/openapi_test.go` | TASK-01 | Pending |
| [ ] | TASK-07 | feat | Implement Central Mock Engine Router, Latency Jitter, Chaos Simulator, and Admin API with tests | `pkg/engine/engine.go`, `pkg/engine/engine_test.go` | TASK-02, TASK-03, TASK-04, TASK-05, TASK-06 | Pending |
| [ ] | TASK-08 | feat | Implement embedded web asset handler, index.html shell, and Linear Design System CSS adhering to DESIGN.md | `pkg/web/web.go`, `web/index.html`, `web/css/style.css` | TASK-07 | Pending |
| [ ] | TASK-09 | feat | Implement Frontend SPA JS modules: core app, API client, and views (Dashboard, Endpoints, Collections, Traffic, Auth, OpenAPI, Tester) | `web/js/app.js`, `web/js/api.js`, `web/js/views/dashboard.js`, `web/js/views/endpoints.js`, `web/js/views/collections.js`, `web/js/views/traffic.js`, `web/js/views/auth.js`, `web/js/views/openapi.js`, `web/js/views/tester.js` | TASK-08 | Pending |
| [ ] | TASK-10 | feat | Implement main.go CLI entrypoint with graceful shutdown, demo seeds, and documentation | `main.go`, `README.md` | TASK-07, TASK-08, TASK-09 | Pending |
| [ ] | TASK-11 | review | Run full sensor verification suite (Go tests, Go vet, Spec Drift, Build) and generate formal Verification Report | `pkg/`, `web/`, `main.go` | TASK-10 | Pending |

## Schema Dictionary
- **Status**: `[ ]` (Pending) \| `[x]` (Verified Complete).
- **ID**: `TASK-01`, `TASK-02`, etc.
- **Type**: `test` \| `feat` \| `fix` \| `refactor` \| `docs` \| `rules` \| `skill` \| `review`.
- **Target Files**: Concrete comma-separated file paths (relative to workspace root).
- **Dependencies**: Comma-separated list of preceding task IDs or `None`.
- **Evidence**: Commit hash (`git rev-parse --short HEAD`) + sensor output snippet.

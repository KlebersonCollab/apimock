# Project Roadmap (ROADMAP.md)

## Milestone 1: Core Engine & Dynamic Routing (Completed)
- [x] Go module setup and project skeleton.
- [x] High-performance path matcher with parameter support (`:id`, `*wildcard`).
- [x] Dynamic template & synthetic faker engine (`{{faker.*}}`, `{{req.*}}`, repeat loops).
- [x] Network latency, jitter, and chaos failure rate simulator.
- [x] Security & Auth guard subsystem (JWT, Bearer, API Key, Basic Auth) + Mock Auth token endpoint.
- [x] Stateful Auto-CRUD collection engine with search, sort, pagination, and persistence.
- [x] In-memory ring buffer request inspector with real-time SSE stream.
- [x] OpenAPI 3.0 import/export and workspace persistence.

## Milestone 2: Modern Embedded Web Interface (Completed)
- [x] Embedded web server via Go `embed.FS`.
- [x] Dark canvas Linear design system (`#010102`, `#0f1011`, `#5e6ad2`).
- [x] Interactive Dashboard & Quick Metrics.
- [x] Visual Endpoint Studio & Template Editor with dynamic live preview.
- [x] Stateful Resource DB Table & Record Editor.
- [x] Real-time Live Traffic Stream & Request Inspector with cURL and Replay.
- [x] Mock Auth Generator & Token Playground.
- [x] OpenAPI & Workspace Import/Export Hub.
- [x] In-browser API Test Console.

## Milestone 3: Comprehensive Verification & Release (Completed)
- [x] Automated Go test suite with 100% core coverage.
- [x] Spec Drift validation and clean build verification.
- [x] Documentation and user guide.

## Milestone 4: Conditional Multi-Scenarios & Rule Engine (Active)
- [ ] Domain models for Scenario, Condition, Operator, and MatchMode.
- [ ] Predicate evaluation engine (query, headers, params, body JSON paths).
- [ ] First-match cascading dispatcher with fallback to default response.
- [ ] Tabbed Linear Web Studio UI in Endpoint Studio.
- [ ] Interactive Scenario Builder cards with condition rows.
- [ ] Verification with comprehensive unit & integration tests.

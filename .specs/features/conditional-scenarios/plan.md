# Plan: Conditional Multi-Scenarios & Rule Engine

## 1. Problem Statement & Motivation
In real-world frontend and fullstack development, endpoints have distinct execution paths based on inputs (e.g. valid credentials returning HTTP 200, missing password returning HTTP 400, unauthorized role returning HTTP 403, and nonexistent resource returning HTTP 404). MockForge currently only allows a single static response per endpoint path, forcing developers to alter their code or create artificial route paths. We need a flexible, declarative conditional scenario engine with complete editing ergonomics in the Linear Web Studio.

## 2. Scope & Boundaries
- **In Scope**:
  - Domain modeling for `Scenario`, `Condition`, `MatchMode`, and response overrides in `pkg/models/models.go`.
  - Condition evaluation engine supporting `query`, `header`, `param`, and `body` JSON-path evaluation with operators (`equals`, `not_equals`, `contains`, `regex`, `gt`, `gte`, `lt`, `lte`, `is_empty`, `is_not_empty`).
  - First-match cascading dispatcher in `pkg/engine/engine.go` (evaluates scenarios in priority order; falls back to default response).
  - Tabbed Linear Web Studio UI in `web/js/views/endpoints.js` (General, Default Response, Scenarios & Rules, Auth & Latency).
  - Rich Scenario Editor in the UI allowing adding, editing, reordering, deleting, and toggling individual scenarios and conditions.
  - Comprehensive unit and integration test suite covering all condition operators, body extraction, match modes, and fallback.
- **Out of Scope**:
  - Arbitrary JavaScript / Lua script execution (rejected in ADR 0002 for security and determinism).
  - Complex boolean AST expressions (e.g. `(A AND B) OR (C AND NOT D)`); scenarios use `ALL` or `ANY` match modes.

## 3. High-Level Approach
1. **Model Extension**: Extend `models.Endpoint` with `Scenarios []Scenario`. Define `Scenario` with `ID`, `Name`, `Enabled`, `Priority`, `MatchMode` (`all` / `any`), `Conditions []Condition`, `Response ResponseMock`, and optional `Latency` and `Chaos`.
2. **Evaluation Logic**: Add `EvaluateScenarios(r *http.Request, ep *models.Endpoint, params map[string]string, reqBody string)` in engine/matcher.
3. **Dispatcher Wiring**: Update `dispatchMockEndpoint` to check if a matching scenario is active. If matched, apply the scenario's response, latency, and chaos overrides, interpolated with template Faker tags.
4. **Web UI Tabbed Interface**: Refactor Endpoint Studio modal into clean Linear tabs:
   - Tab 1: Basic Route (Method, Path, Name, Description, Tags)
   - Tab 2: Default Response (Status, Content-Type, Headers, Body)
   - Tab 3: Conditional Scenarios (Interactive cards to add/edit rules, conditions, and custom responses)
   - Tab 4: Latency & Chaos & Auth Guards
5. **Validation & Test Suite**: Add tests for condition evaluation, JSON body path extraction, regex matching, numeric comparisons, and UI integration.

## 4. Dependencies & Prerequisites
- Pure Go standard library (no third-party dependencies).
- Conforms strictly to `DESIGN.md` dark canvas tokens.

## 5. Architectural Decision Records (ADRs)
- [ADR 0002: Conditional Multi-Scenarios & Rule Engine](../../project/ADRs/0002-conditional-multi-scenarios.md)

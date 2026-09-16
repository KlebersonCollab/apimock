# Specification: Conditional Multi-Scenarios & Rule Engine

## 1. User Stories
- **US-1**: As a frontend developer, I want an endpoint to return different status codes and payloads depending on request parameters or body fields, so that I can test edge cases and error states without changing my frontend code.
- **US-2**: As a developer, I want to create and manage multiple scenarios in the Web Studio using intuitive tabs and condition builders, so that I can visually configure my API simulation.
- **US-3**: As a QA engineer, I want request conditions to evaluate query strings, headers, route params, and nested JSON body fields with various operators (equals, contains, regex, numeric comparisons), so that I can simulate complex business logic.

## 2. Acceptance Criteria (BDD)

### AC-1: Query Parameter Matching
- **Given** an endpoint `GET /api/v1/users` with a scenario requiring query `status == "archived"` returning HTTP 200 `[]`
- **When** a request `GET /api/v1/users?status=archived` is dispatched
- **Then** the scenario response is returned instead of the default response.

### AC-2: Request Body JSON-Path Evaluation
- **Given** an endpoint `POST /api/v1/login` with a scenario matching `body.email == "admin@example.com"` returning HTTP 200 with admin token, and another matching `body.password == "wrong"` returning HTTP 401
- **When** a request `POST /api/v1/login` with `{"email":"admin@example.com","password":"secret"}` is sent
- **Then** HTTP 200 with the admin response is returned.
- **When** a request with `{"email":"admin@example.com","password":"wrong"}` is sent
- **Then** HTTP 401 Unauthorized is returned.

### AC-3: Fallback to Default Response
- **Given** an endpoint with 2 conditional scenarios
- **When** an incoming request does not satisfy any scenario's conditions
- **Then** the endpoint's default `ResponseMock` is evaluated and returned.

### AC-4: First-Match Priority Execution
- **Given** scenario A with Priority 1 and scenario B with Priority 2, where both could match a request
- **When** the request arrives
- **Then** scenario A is selected and executed, terminating further evaluation.

### AC-5: Match Modes (ALL vs ANY)
- **Given** a scenario configured with `MatchMode: "all"` having 2 conditions
- **When** only 1 condition is met
- **Then** the scenario does NOT match.
- **When** `MatchMode: "any"` is configured and 1 condition is met
- **Then** the scenario DOES match.

### AC-6: Web Studio Portal Tabbed UI
- **Given** the user opens an Endpoint in the Endpoint Studio
- **When** viewing the modal dialog
- **Then** 4 clean navigation tabs are displayed: "General", "Default Response", "Scenarios & Rules", "Auth & Latency".
- **When** clicking "Scenarios & Rules", the user can view, add, edit conditions, and delete scenarios.

## 3. Verification Sensors
| Sensor | Command / Target | Success Threshold |
|---|---|---|
| Go Test Suite | `go test -v -race ./...` | 100% pass |
| Spec Drift Sensor | `node .agents/scripts/check-spec-drift.js` | 0 drifted files |
| Go Vet | `go vet ./...` | 0 issues |
| Build Sensor | `go build -o mockforge.exe .` | Exit code 0 |

## 4. UI & Design System Tokens (DESIGN.md)
- **Canvas & Surfaces**: Background `--color-canvas` (`#010102`), cards `--color-surface-1` (`#0f1011`), inner rows `--color-surface-2` (`#141516`).
- **Hairlines**: `--color-hairline` (`#23252a`) and `--color-hairline-strong` (`#34343a`).
- **Accent & Badges**: `--color-primary` (`#5e6ad2`), status badges matching method colors.
- **Interactive Tabs**: Linear pill/tab style with active bottom indicator or active background highlight.

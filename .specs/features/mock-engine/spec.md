# Feature Specification: MockForge Studio Engine (spec.md)

# Specification: MockForge Studio Engine

## 1. User Stories
- **US-1**: As a frontend developer, I want to create dynamic mock endpoints with realistic fake data templates so that I can prototype UI features without waiting for backend implementation.
- **US-2**: As a developer, I want to simulate network latency, jitter, and chaos error rates so that I can test my frontend's loading skeletons, error states, and retry resilience.
- **US-3**: As a developer, I want to protect mock endpoints with simulated authentication (Bearer, API Key, Basic, JWT) and generate test tokens so that I can develop auth guards and login flows.
- **US-4**: As a developer, I want stateful auto-CRUD collections with filtering and pagination so that my UI can perform realistic create, update, and delete actions with persistent state.
- **US-5**: As a developer, I want a real-time live traffic stream so that I can inspect inbound requests, view latency and payload details, and replay calls.
- **US-6**: As a developer, I want an integrated, beautiful web interface based on Linear's design system so that I can manage my mocks effortlessly in one place.
- **US-7**: As a developer, I want to import and export OpenAPI 3.0 specs and workspace backups so that I can sync with API specifications and share configurations with my team.

## 2. Acceptance Criteria (BDD)

### AC-1: Dynamic Parameterized Mock Route Handling & Templates
- **Given** an active mock endpoint configured for `GET /api/v1/users/:id` with body `{"id": "{{req.params.id}}", "name": "{{faker.name}}", "email": "{{faker.email}}"}`
- **When** an HTTP GET request is received at `/api/v1/users/42`
- **Then** the engine returns HTTP 200 with `application/json`, `"id": "42"`, and a realistic name and email.

### AC-2: Network Latency & Jitter Simulation
- **Given** a mock endpoint configured with a min latency of 100ms and max latency of 200ms
- **When** a client sends a request to the endpoint
- **Then** the response is delayed by at least 100ms and no more than 250ms before returning.

### AC-3: Chaos Error Rate Simulation
- **Given** a mock endpoint configured with a 100% chaos failure rate returning status HTTP 503
- **When** a client sends a request to the endpoint
- **Then** the engine immediately intercepts and returns HTTP 503 with the configured chaos error response payload.

### AC-4: Authentication Guard & JWT Validation
- **Given** a mock endpoint protected by Bearer auth with token `secret-token-xyz`
- **When** a request is made without `Authorization: Bearer secret-token-xyz`
- **Then** the engine returns HTTP 401 Unauthorized with error details.
- **When** a request is made with the valid header `Authorization: Bearer secret-token-xyz`
- **Then** the engine processes and returns the mock response.

### AC-5: Stateful Auto-CRUD Collection
- **Given** an empty stateful collection named `products`
- **When** a POST request is made to `/api/resources/products` with body `{"title": "Mechanical Keyboard", "price": 120}`
- **Then** the resource is stored, returns HTTP 201 with generated ID and timestamps, and is queryable via `GET /api/resources/products` and `GET /api/resources/products/:id`.

### AC-6: Real-time Live Traffic Stream
- **Given** a client connected to SSE `/api/admin/traffic/stream`
- **When** any mock endpoint is requested
- **Then** an SSE event is broadcast containing method, path, status, latency, headers, and payload.

### AC-7: OpenAPI 3.0 Import
- **Given** a valid OpenAPI 3.0 JSON specification
- **When** imported via `/api/admin/openapi/import`
- **Then** endpoints corresponding to paths and methods are automatically created with default mock response payloads.

## 3. Verification Sensors
| Sensor | Command / Target | Success Threshold |
|---|---|---|
| Go Tests | `go test -v -race ./...` | 100% pass |
| Go Vet | `go vet ./...` | 0 errors |
| Spec Drift | `node .agents/scripts/check-spec-drift.js` | 0 drifting files |
| Build | `go build -o mockforge.exe .` | Clean exit 0 |

## 4. UI & Design System Tokens (Conforming to DESIGN.md)
- **Background Canvas**: `{colors.canvas}` (`#010102`)
- **Card Surfaces**: `{colors.surface-1}` (`#0f1011`), `{colors.surface-2}` (`#141516`), `{colors.surface-3}` (`#18191a`)
- **Hairline Borders**: `{colors.hairline}` (`#23252a`), `{colors.hairline-strong}` (`#34343a`)
- **Primary Accent**: `{colors.primary}` (`#5e6ad2`), Hover `{colors.primary-hover}` (`#828fff`), Focus `{colors.primary-focus}` (`#5e69d1`)
- **Text Hierarchy**: `{colors.ink}` (`#f7f8f8`), `{colors.ink-muted}` (`#d0d6e0`), `{colors.ink-subtle}` (`#8a8f98`)
- **Semantic Badges**: `{colors.semantic-success}` (`#27a644`), Warning amber (`#e59b20`), Danger rose (`#e5484d`)
- **Typography**: Display/Heading with negative tracking (`-0.6px` to `-1.0px`), Body 14-16px, Mono for code and endpoints.
- **Border Radius**: `{rounded.md}` (8px) for buttons/inputs, `{rounded.lg}` (12px) for cards, `{rounded.xl}` (16px) for panels, `{rounded.pill}` (9999px) for status badges.

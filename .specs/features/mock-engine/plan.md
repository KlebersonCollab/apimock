# Feature Plan: MockForge Studio Engine (plan.md)

# Plan: MockForge Studio Engine

## 1. Problem Statement & Motivation
Frontend developers often experience severe delivery blockers waiting for backend APIs to be specified, coded, or deployed. Existing mock solutions are either too rigid (returning static JSON without realistic variability), too difficult to configure (requiring extensive third-party services or heavy node dependencies), or lack critical real-world features such as auth validation (JWT/Bearer), realistic network latency simulation, chaos injection, stateful REST CRUD, and live traffic observability.

MockForge solves this problem by delivering a single-binary, high-performance mock API engine written in Go with an integrated, ultra-elegant web management interface designed according to Linear's aesthetic.

## 2. Scope & Boundaries
- **In Scope**:
  - High-performance parameterized and wildcard HTTP mock router.
  - Template & synthetic faker engine for dynamic realistic data (`{{faker.*}}`, `{{req.*}}`, `{{#repeat N}}`).
  - Latency & chaos error simulator (fixed delay, jitter range, probabilistic chaos status codes).
  - Auth Guard Simulator (No Auth, Bearer Token, API Key, Basic Auth, JWT verification & Mock Login token generator).
  - Stateful Auto-CRUD Collection Store with search, filtering, pagination, and sorting.
  - Real-time request traffic inspector with circular buffer and Server-Sent Events (SSE) streaming.
  - OpenAPI 3.0 import/export and JSON workspace backup/restore.
  - Interactive web application adhering to `DESIGN.md` embedded into the Go binary.
  - Built-in In-Browser API Test Console.
- **Out of Scope**:
  - WebSocket protocol mocking (planned for subsequent milestone).
  - GraphQL schema mock execution (planned for subsequent milestone).

## 3. High-Level Approach
1. **Domain Models & Core Router**: Implement strongly-typed models (`Endpoint`, `Collection`, `TrafficLog`, `AuthConfig`) and a fast thread-safe path matcher supporting parameterized tokens (`:id`) and wildcards (`*path`).
2. **Template & Faker Engine**: Build a lightweight evaluation engine capable of parsing interpolation tokens, generating deterministic or random synthetic fake data (names, emails, addresses, numbers, dates), and injecting request context.
3. **Stateful Store**: Implement a thread-safe in-memory key-value and collection store supporting full REST verbs, automatic ID generation, query filtering (`?field=val`), pagination (`?page=1&limit=10`), sorting, and search.
4. **Auth & Latency Modules**: Implement auth middleware guards, JWT signature/claim verification, and latency/chaos jitter injection.
5. **Traffic Inspector & SSE**: Implement circular buffer log and SSE broker pushing live request telemetry to connected UI clients.
6. **OpenAPI Hub**: Implement JSON/YAML OpenAPI 3.0 parser converting specs into MockForge endpoints and vice-versa.
7. **Frontend Web UI**: Implement an embedded Single Page App in `web/` using HTML5, modern modular JS, and Linear Design System styles (`DESIGN.md`).
8. **Embed & Main Entrypoint**: Wire the Go `embed.FS` to serve both the mock engine, admin API, and static assets seamlessly on a single port (with CLI flags for customization).

## 4. Dependencies & Prerequisites
- Go 1.22+ runtime.
- Git repository with pre-commit spec drift sensor installed.

## 5. Architectural Decision Records (ADRs)
- [ADR 0001: Mock Engine Architecture and Embedded Web UI](../../project/ADRs/0001-mock-engine-architecture.md)

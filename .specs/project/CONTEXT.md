# Domain Glossary & Project Context (CONTEXT.md)

## 1. Domain Glossary

| Term | Canonical Definition | Usage Context |
|---|---|---|
| **MockForge Engine** | The core Go HTTP mock server responsible for intercepting, evaluating, and serving simulated responses. | Backend core runtime |
| **Mock Endpoint** | A configured HTTP route with method, path pattern, matching rules, response status, headers, and dynamic body template. | Routing & dispatching |
| **Route Pattern** | An endpoint path specification supporting path parameters (e.g. `/api/v1/users/:id`), wildcards (`/assets/*path`), and query filters. | Path matcher |
| **Template Engine** | An evaluation system processing tags like `{{faker.name}}`, `{{req.params.id}}`, `{{req.query.q}}`, and `{{#repeat N}}`. | Dynamic response generation |
| **Faker Generator** | A realistic synthetic data generator producing emails, names, UUIDs, dates, lorem ipsum, addresses, companies, and prices. | Data realism |
| **Stateful Collection** | An in-memory/persisted JSON resource collection (e.g. `users`, `products`) providing automatic full REST CRUD (`GET`, `POST`, `PUT`, `PATCH`, `DELETE`) with search, filter, sort, and pagination. | Dynamic resource mocking |
| **Network & Latency Simulator** | Engine module capable of applying fixed latency, randomized jitter ranges, or chaos failure rates (e.g. HTTP 500/503/429 injection). | Resilience and latency testing |
| **Auth Guard Simulator** | Security subsystem validating Bearer Tokens, API Keys, Basic Auth, or JWT claims/signatures, accompanied by a mock token generator. | Frontend auth decoupling |
| **Live Traffic Inspector** | A real-time SSE stream and log buffer of inbound HTTP requests with headers, payloads, timings, and replay functionality. | Observability & debugging |
| **OpenAPI Hub** | Subsystem for bidirectional import and export of OpenAPI 3.0 specifications and full workspace configurations. | Interoperability |
| **Linear Canvas UI** | Embedded single-page frontend styled strictly according to `DESIGN.md` tokens (near-black `#010102` canvas, `#5e6ad2` lavender accent). | User interface |

## 2. Project Boundaries
- **Primary Goal**: Provide a single-binary, high-performance mock API server in Go with an integrated, ultra-elegant web UI for rapid frontend-backend decoupling.
- **Audience**: Frontend engineers, fullstack developers, QA testers, and software architects requiring immediate, reliable, realistic APIs without backend bottlenecks.

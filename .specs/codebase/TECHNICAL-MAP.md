# Technical Map (TECHNICAL-MAP.md)

## Subsystems & Package Responsibilities

| Package / Module | Responsibility | Key Structs / Functions |
|---|---|---|
| `main.go` | CLI entry point, flag parsing, graceful shutdown, server startup. | `main()` |
| `pkg/models` | Data contracts for endpoints, responses, collections, auth, logs. | `Endpoint`, `ResponseMock`, `Collection`, `TrafficEntry`, `AuthConfig` |
| `pkg/template` | Dynamic template tag resolution, faker generators, repeat blocks. | `Engine`, `Evaluate(template, reqContext)`, `Faker` |
| `pkg/auth` | Bearer, API Key, Basic Auth, and JWT validation / generation. | `Guard`, `ValidateRequest()`, `GenerateMockJWT()` |
| `pkg/store` | In-memory CRUD data store with filtering, pagination, sorting. | `CollectionStore`, `Create()`, `List()`, `Update()`, `Delete()` |
| `pkg/traffic` | Circular buffer for request logging and SSE streaming. | `Logger`, `StreamClients`, `LogRequest()`, `Subscribe()` |
| `pkg/engine` | Central HTTP dispatch, routing, latency & chaos simulation. | `Server`, `ServeHTTP()`, `RegisterEndpoint()`, `ApplyLatency()` |
| `pkg/openapi` | OpenAPI 3.0 import & export parser and serializer. | `ImportOpenAPI()`, `ExportOpenAPI()`, `Workspace` |
| `web/` | HTML, CSS, JS single-page web UI adhering to `DESIGN.md`. | Embedded in Go binary via `embed.FS` |

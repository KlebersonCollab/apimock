# ADR 0002: Conditional Multi-Scenarios & Rule Engine

## Context & Problem Statement
MockForge endpoints currently support only a single static response mock per route pattern (`[METHOD] /path`). In real-world frontend testing and integration, developers need to simulate diverse edge cases and branching behaviors on the same endpoint (e.g. valid credentials returning HTTP 200, invalid password returning HTTP 401, missing query parameters returning HTTP 400, or specific IDs returning HTTP 404). Creating separate mock routes for each variation is clumsy, unrealistic, and breaks standard REST contracts.

## Decision Drivers
1. **Expressive Rule Matching**: Support evaluating request query params, HTTP headers, URL route parameters, and JSON request body fields.
2. **First-Match Cascading (Deterministic Execution)**: Evaluate rules in explicit priority order. The first matching scenario executes; if none match, the endpoint's Default Response serves as fallback.
3. **100% Backward Compatibility**: Existing endpoints without scenarios must continue operating identically without migration or breakage.
4. **Full Portal Ergonomics**: Allow creating, editing, reordering, enabling/disabling, and testing scenarios directly in the Linear Canvas Web UI via dedicated modal tabs.

## Considered Options
1. **Separate Endpoints with Query Hacks**: Force developers to configure `/api/users?scenario=error` as distinct paths. (Rejected: Unrealistic, pollutes route table, doesn't support body/header inspection).
2. **Custom JavaScript / Lua Scripting Sandbox**: Embed an interpreter to execute custom user code per request. (Rejected: Heavy runtime overhead, security risks, steep learning curve).
3. **Declarative Predicate Rule Engine**: Add a `Scenarios` list to `Endpoint` with declarative conditions (`source`, `property`, `operator`, `value`), logical match modes (`ALL` vs `ANY`), and response overrides.

## Decision Outcome
**Option 3**: Declarative Predicate Rule Engine with First-Match Cascading and Tabbed Linear Web Studio UI.

### Positive Consequences
- Zero script-injection or sandbox vulnerabilities.
- Sub-millisecond rule evaluation in pure Go.
- Seamless editing in the UI with intuitive dropdowns and dynamic condition rows.
- Full compatibility with OpenAPI and Workspace JSON serialization.

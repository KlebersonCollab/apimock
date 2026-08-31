# ADR 0001: Mock Engine Architecture and Embedded Web UI

## Context & Problem Statement
Developers need to decouple frontend development from backend readiness without setting up heavy infrastructure or depending on complex runtime environments. The tool must be extremely fast, light, cross-platform, and provide a top-tier visual experience out of the box with zero setup steps.

## Decision Drivers
1. **Single-Binary Portability**: Users should be able to run `mockforge` or `go run main.go` and have both the high-performance mock server and the web interface immediately live.
2. **Deterministic & Realistic Mocking**: Dynamic templates with rich faker tags, parameter interpolation, configurable latency, chaos errors, and realistic auth verification.
3. **Low Latency & High Concurrency**: Pure Go concurrency primitives (`sync.RWMutex`, atomic counters, channels for SSE live streaming).
4. **Visual Aesthetics**: Strict alignment with `DESIGN.md` (Linear aesthetic: `#010102` canvas, hairline borders, lavender accents).

## Considered Options
1. Go Backend with separate Node.js/React frontend requiring `npm install` / `npm run dev` at runtime.
2. Pure Node.js / Express tool with React.
3. Single Go application utilizing `net/http` and `embed.FS` with an integrated modern SPA.

## Decision Outcome
**Option 3**: Pure Go application embedding static SPA assets via `embed.FS`.

### Positive Consequences
- Instant execution (`go run .` or compiled binary) with zero runtime dependencies.
- Sub-millisecond mock response handling and minimal memory footprint (< 25MB RAM).
- Built-in SSE live traffic streaming without third-party brokers.
- Embeddable and shareable as a single static executable for teams and CI/CD pipelines.

### Negative Consequences / Trade-offs
- Frontend build must be bundled into static files embedded into Go, requiring modular vanilla JS/ES modules rather than heavy runtime transpilers during runtime.

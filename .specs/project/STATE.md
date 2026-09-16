# Project State & Context (STATE.md)

## 🏁 Session Status
- **Current Task**: Implementing Conditional Multi-Scenarios & Rule Engine (`conditional-scenarios`).
- **Progress**: Planning approved, moving to TDD execution (TASK-01 to TASK-07).
- **Next Steps**: Scaffold unit tests in `pkg/engine/scenarios_test.go` and implement scenario data structures.

## 💡 Decisions Log
- **2026-08-31 - Architecture**: Selected single-binary Go standard library architecture with embedded SPA frontend via `embed.FS` to maximize portable execution without node runtime dependency for users.
- **2026-08-31 - UI Design**: Adopted strict Linear design tokens (`#010102` canvas, `#0f1011` surface-1, `#5e6ad2` primary) conforming to `DESIGN.md`.
- **2026-08-31 - Verification**: 100% pass on all Go unit and integration tests, Go vet static analysis, Spec Drift sensor, and executable build.
- **2026-09-16 - Feature Multi-Scenarios**: Selected declarative first-match cascading predicate rule engine with tabbed modal UI in Endpoint Studio (ADR 0002).

## 🚧 Active Blockers
- None.

## ❄️ Deferred Ideas / Icebox
- WebSocket mocking engine (future milestone).
- GraphQL mock schema resolver (future milestone).

## ⚠️ Known Technical Debts
- None.

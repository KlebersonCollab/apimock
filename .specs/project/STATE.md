# Project State & Context (STATE.md)

## 🏁 Session Status
- **Current Task**: Initialize SDD artifacts, technical mapping, architectural decision records, and feature plan for MockForge.
- **Progress**: Phase 0 (Discovery & Planning) in progress.
- **Next Steps**: Complete technical specs, feature plan, acceptance criteria, and 7-column tasks.md.

## 💡 Decisions Log
- **2026-08-31 - Architecture**: Selected single-binary Go standard library architecture with embedded SPA frontend via `embed.FS` to maximize portable execution without node runtime dependency for users.
- **2026-08-31 - UI Design**: Adopted strict Linear design tokens (`#010102` canvas, `#0f1011` surface-1, `#5e6ad2` primary) conforming to `DESIGN.md`.

## 🚧 Active Blockers
- None.

## ❄️ Deferred Ideas / Icebox
- WebSocket mocking engine (future milestone).
- GraphQL mock schema resolver (future milestone).

## ⚠️ Known Technical Debts
- None (Greenfield development).

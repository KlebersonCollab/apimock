# Project Vision & North Star (PROJECT.md)

## 1. Executive Summary
**MockForge** is a fast, powerful, and elegant developer tool written in Golang with an embedded modern web interface. It empowers frontend and fullstack engineers to completely decouple frontend development from backend readiness. With MockForge, developers can spin up instant REST endpoints, generate deeply realistic data with an integrated template faker engine, simulate network latency and chaos failure rates, enforce simulated authentication guards (JWT, Bearer, API Key, Basic), manage stateful auto-CRUD resource collections, inspect live request streams via Server-Sent Events, and import/export OpenAPI 3.0 specs.

## 2. Core Pillars
1. **Instant Developer Ergonomics**: Zero complex configuration, single-binary distribution with embedded UI, instant startup under 10 milliseconds.
2. **True-to-Life Simulation**: Dynamic templates, realistic data synthesis, customizable response delay and jitter, probabilistic chaos errors.
3. **Decoupled Auth & Security**: Built-in mock JWT issuer and claim verifier, header/query API key guards, Basic Auth.
4. **Stateful REST Resource Engine**: Auto-CRUD database for realistic interactive frontend state manipulation without writing backend code.
5. **Quiet Luxury UX**: Designed with Linear-inspired typography, dark surface hierarchy (`#010102`), and high-density product control panels.

## 3. Success Metrics
- Sub-millisecond mock dispatch latency (excluding user-configured artificial delay).
- 100% standalone binary capability with `embed.FS`.
- Zero runtime crashes with concurrent thread safety.
- Complete compliance with `DESIGN.md` design tokens.

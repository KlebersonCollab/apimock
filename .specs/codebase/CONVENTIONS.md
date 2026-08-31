# Codebase Conventions (CONVENTIONS.md)

## Go Code Conventions
1. **Error Handling**: Explicit `if err != nil` checks. No silent swallowing of errors.
2. **Concurrency & Thread Safety**: All shared engine structures MUST be protected by `sync.RWMutex`.
3. **Idiomatic Formatting**: Standard `gofmt` and `go vet` rules.
4. **Package Organization**: Packages under `pkg/` should be decoupled and focused on single responsibilities.
5. **No External Heavy Dependencies**: Standard library preferred whenever possible for ultra-lightweight binaries.

## Frontend Conventions
1. **Design System Adherence**: Strictly use CSS tokens from `DESIGN.md` (`--color-canvas: #010102`, `--color-primary: #5e6ad2`, `--color-surface-1: #0f1011`).
2. **Component Modularity**: UI views separated into focused modules (`views/dashboard.js`, `views/endpoints.js`, etc.).
3. **XSS Protection & Escaping**: All dynamic user-provided template text rendered in UI must be sanitized/escaped before DOM insertion.
4. **State Management**: Reactive centralized state with pub/sub event emitters for live updates.

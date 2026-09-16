# Critical Risks & Technical Debt (CONCERNS.md)

## 1. High-Risk Areas (Mandatory Caution Paths)

### A. Linear Search in Route Matching (`pkg/engine/engine.go`)
- **Location**: `matchEndpoint()` in `pkg/engine/engine.go` (lines 472–491)
- **Reality**: On every incoming request, the server acquires a read lock and iterates over all registered endpoints via `for _, ep := range s.endpoints`, parsing route tokens dynamically using `splitPath(ep.Path)` on every iteration.
- **Risk**: Latency scales as $O(N \times M)$ where $N$ is the number of endpoints and $M$ is the route segment depth. For workspaces with hundreds of endpoints, this can degrade dispatch performance.
- **Mitigation Path**: Pre-parse route tokens on endpoint insertion/update, or migrate to a Radix Tree / Trie matcher.

### B. Synchronous File I/O under Store Mutex (`pkg/store/store.go`)
- **Location**: `saveToFileUnsafe()` called inside `CreateItem`, `UpdateItem`, `DeleteItem`, `CreateCollection`, `UpdateCollection`, `DeleteCollection` while holding `cs.mu.Lock()`.
- **Reality**: The entire in-memory collection map is serialized to indented JSON and written to disk synchronously with `os.WriteFile` while holding the write lock.
- **Risk**: High-frequency CRUD operations on large collections will stall all readers and writers on disk I/O bottlenecks.
- **Mitigation Path**: Implement an asynchronous debounced write worker or channel-based flusher.

### C. Dormant / Unexposed Upstream Reverse Proxy (`pkg/engine/engine.go`)
- **Location**: `proxyTarget` field and `proxyPass()` in `pkg/engine/engine.go` (lines 34, 186, 821–875)
- **Reality**: The reverse proxy fallback is implemented with full HTTP forwarding and traffic logging, but `proxyTarget` is never set, exposed via CLI flags, or configurable via the Admin API/UI.
- **Risk**: Dead code path and missed capability for transparent API pass-through or Record & Replay.
- **Mitigation Path**: Wire `proxyTarget` to CLI flag `--proxy` and add runtime toggle/URL in Admin Settings.

---

## 2. Technical Debt & Architectural Gaps

### A. Transient Endpoints vs. Persisted Collections
- **Reality**: Collections support optional JSON persistence via `-store` flag, but custom endpoints created in the Web UI or imported via OpenAPI/Admin API are stored strictly in-memory. If the process terminates, all custom endpoints are lost unless manually exported to a Workspace JSON.
- **Mitigation Path**: Add unified workspace file persistence (e.g. `--workspace path/to/workspace.json` or auto-save).

### B. Single Static Response per Endpoint (Lack of Conditional Rules)
- **Reality**: Each `Endpoint` struct has exactly one `ResponseMock`. All requests matching `[METHOD] /path` receive this response (interpolated with Faker/request variables).
- **Impact**: Unable to simulate multi-branch APIs without creating divergent route paths.

### C. Hardcoded CORS Policy (`*`)
- **Location**: `ServeHTTP()` in `pkg/engine/engine.go` (lines 147–156)
- **Reality**: `Access-Control-Allow-Origin: *` is unconditionally set.
- **Impact**: Browsers with `fetch(url, { credentials: 'include' })` will reject requests when wildcard origin is returned.
- **Mitigation Path**: Support configurable allowed origins and `Access-Control-Allow-Credentials`.

---

## 3. Test Coverage Fragility
- `pkg/web` has no dedicated Go unit tests (`[no test files]`). Static file serving and fallback routing should have automated tests.

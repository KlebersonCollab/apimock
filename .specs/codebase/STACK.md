# Tech Stack (STACK.md)

## Backend Runtime & Language
- **Language**: Go 1.22+ / 1.26
- **HTTP Server**: `net/http` standard library with lightweight chi-style parameterized routing.
- **Concurrency**: Native Go goroutines, channels, and `sync.RWMutex`.
- **Packaging / Distribution**: Single executable binary with `embed.FS`.

## Frontend Runtime & Technologies
- **Design System**: Linear Dark Canvas Design System (`DESIGN.md`).
- **Core UI**: Modern ES6+ Modular Vanilla SPA with reactive state store and dynamic DOM rendering.
- **Icons**: Handcrafted SVG icons matching Linear stroke style.
- **Styling**: Scoped CSS variables adhering to `DESIGN.md` tokens (`#010102`, `#0f1011`, `#5e6ad2`).
- **Communication Protocol**: REST APIs + Server-Sent Events (SSE) for real-time traffic streaming.

## Testing & Quality Sensors
- **Go Unit & Integration Tests**: `go test -v -race ./...`
- **Spec Drift Sensor**: `node .agents/scripts/check-spec-drift.js`
- **Build Sensor**: `go build -o mockforge.exe .`

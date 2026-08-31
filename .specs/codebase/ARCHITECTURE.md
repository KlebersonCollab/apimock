# Codebase Architecture (ARCHITECTURE.md)

## System Overview

```
                        ┌──────────────────────────────────────────────┐
                        │              Client / Frontend App           │
                        └──────────────┬───────────────────────────────┘
                                       │ HTTP Requests (:8080)
                                       ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│ MockForge Server (Go Runtime)                                               │
│                                                                             │
│  ┌─────────────────┐       ┌─────────────────┐       ┌───────────────────┐  │
│  │   HTTP Router   │──────▶│   Auth Guard    │──────▶│ Latency & Chaos   │  │
│  │  & Path Matcher │       │    Validator    │       │     Simulator     │  │
│  └─────────────────┘       └─────────────────┘       └─────────┬─────────┘  │
│                                                                │            │
│                 ┌──────────────────────────────────────────────┘            │
│                 ▼                                                           │
│  ┌─────────────────────────┐  ┌────────────────────────┐                    │
│  │  Static & Custom Mocks  │  │  Stateful Collections  │                    │
│  │  Template & Faker Engine│  │   (Auto-CRUD In-Mem)   │                    │
│  └──────────────┬──────────┘  └───────────┬────────────┘                    │
│                 │                         │                                 │
│                 └───────────┬─────────────┘                                 │
│                             ▼                                               │
│               ┌───────────────────────────┐                                 │
│               │ Traffic Inspector RingLog │─────▶ [SSE Stream: /api/admin]  │
│               └───────────────────────────┘                                 │
│                                                                             │
│  ┌───────────────────────────────────────────────────────────────────────┐  │
│  │ Embedded Web Management Studio (UI via embed.FS)                      │  │
│  │ - Dashboard & Metrics    - Endpoint Studio       - Stateful Collections│ │
│  │ - Live Traffic Inspector - Auth Token Simulator  - OpenAPI Import/Exp  │ │
│  └───────────────────────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────────────────────┘
```

## Directory Structure
```
.
├── main.go                       # Application entrypoint & CLI flags
├── pkg/
│   ├── models/                   # Domain models (Endpoint, Collection, TrafficLog, AuthConfig)
│   ├── engine/                   # Core mock engine (Router, Matcher, Dispatcher)
│   ├── template/                 # Template interpolator & Faker generator
│   ├── store/                    # Stateful collection store (in-memory + persistence)
│   ├── auth/                     # Auth guard simulator & JWT issuer/validator
│   ├── traffic/                  # Ring buffer traffic inspector & SSE broadcaster
│   ├── openapi/                  # OpenAPI 3.0 parser, generator & workspace exporter
│   └── web/                      # Embedded web server & static asset handler
└── web/                          # Frontend SPA assets
    ├── index.html                # Single-page UI shell
    ├── css/
    │   └── style.css             # Linear design system CSS
    └── js/
        ├── app.js                # Core SPA router & reactive state
        ├── api.js                # Backend client
        ├── views/                # Modular view controllers
        │   ├── dashboard.js      # Dashboard metrics
        │   ├── endpoints.js      # Visual endpoint studio
        │   ├── collections.js    # Stateful collection manager
        │   ├── traffic.js        # Real-time traffic stream
        │   ├── auth.js           # Auth simulator & JWT studio
        │   ├── openapi.js        # OpenAPI / Workspace import & export
        │   └── tester.js         # Interactive API test console
        └── components/           # Reusable UI components & modals
```

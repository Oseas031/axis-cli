# axis-gui

Observation dashboard for Axis. Non-invasive — reads `.axis/` state and proxies write operations to the control plane, never imports `internal/`.

## Architecture

```
tools/axis-gui/
├── main.go          # Go HTTP server (REST API + static file serving + WS)
├── go.mod           # Independent module (no dependency on axis core)
└── frontend/
    ├── src/         # React + TypeScript + Tailwind source
    ├── dist/        # Built frontend (committed)
    └── package.json
```

## Design Principles

1. **Non-invasive**: Does not import `github.com/axis-cli/axis/internal/...`
2. **Observe + proxy**: Reads `.axis/` files for state; write operations proxy to control plane via HTTP
3. **Optional**: Axis functions fully without GUI
4. **Shell-native first**: GUI supplements CLI, does not replace it
5. **Independent module**: Has its own `go.mod`, can be built separately

## API Endpoints

### Core (frontend-facing)

| Endpoint | Method | Source | Response |
|----------|--------|--------|----------|
| `/api/health` | GET | Proxy → `/v1/health` | `{"status":"ok",...}` |
| `/api/runtime/status` | GET | `.axis/runtime.json` + probe | `{"connected":bool,"health":{...},"hint":"..."}` |
| `/api/runtime/start` | POST | — | `{"message":"..."}` (guidance only) |
| `/api/runtime/stop` | POST | — | `{"message":"..."}` (guidance only) |
| `/api/tasks` | GET | `.axis/events/tasks.jsonl` | `{"tasks":[...]}` |
| `/api/tasks` | POST | Proxy → `/v1/tasks` | `{"task_id":"...","status":"..."}` |
| `/api/tasks/{id}/status` | GET | Proxy → `/v1/tasks/{id}/status` | `{"task_id":"...","status":"..."}` |
| `/ws/events` | WS | Tail `.axis/events/tasks.jsonl` | JSON lines pushed per event |

### Legacy / supplementary

| Endpoint | Method | Source | Description |
|----------|--------|--------|-------------|
| `/api/events` | GET | `.axis/events/tasks.jsonl` | Raw JSONL as array |
| `/api/runtime` | GET | `.axis/runtime.json` | Raw runtime record |
| `/api/providers` | GET | `.axis/providers.json` | Provider profiles |
| `/api/mailbox/` | GET | `.axis/comm/` | List actors |
| `/api/mailbox/{id}` | GET | `.axis/comm/{id}.jsonl` | Actor messages |
| `/api/skills` | GET | `.axis/skills/` | Skills with parsed metadata (name, description, tags, version, depends_on, conflicts_with) |

## Frontend Views

| Route | View | Description |
|-------|------|-------------|
| `/` | Dashboard | Runtime status, task summary, quick actions |
| `/tasks` | Tasks | Task list from event log |
| `/providers` | Providers | LLM provider profiles |
| `/events` | Events | Raw event stream |
| `/skills` | Skills | Installed skills with metadata, CLI command reference |
| `/chat` | Chat | Submit tasks via natural language, history search/collapse |

## Dev Loop

```bash
cd tools/axis-gui
go build -o axis-gui.exe .          # build
taskkill /F /IM axis-gui.exe        # kill old process
.\axis-gui.exe --port 3000 --root "C:\path\to\axis-cli"  # restart
# verify: curl http://localhost:3000/api/runtime/status
```

> **Important**: The running process must be restarted after recompilation. A stale process serves old code.

## Frontend Development

```bash
cd tools/axis-gui/frontend
npm install
npm run dev    # Vite dev server with HMR
npm run build  # Build to dist/
```

## Boundary Rules

- MUST NOT import any `github.com/axis-cli/axis/internal/` package
- MUST NOT write to `.axis/` directory directly
- MUST NOT modify Axis runtime behavior
- MUST communicate only through file reads and HTTP proxy to control server
- MAY be absent — Axis never depends on GUI being available
- Write operations (task submit) are proxied to the control plane, never executed locally

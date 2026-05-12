# axis-gui

Observation dashboard for Axis. Non-invasive — reads `.axis/` state only, never imports `internal/`.

## Architecture

```
tools/axis-gui/
├── main.go          # Go HTTP server (REST API + static file serving)
├── go.mod           # Independent module (no dependency on axis core)
└── frontend/
    ├── src/         # React + TypeScript + Tailwind source
    ├── dist/        # Built frontend (gitignored)
    └── package.json
```

## Design Principles

1. **Non-invasive**: Does not import `github.com/axis-cli/axis/internal/...`
2. **Observation only**: Reads `.axis/` files, does not mutate Axis state
3. **Optional**: Axis functions fully without GUI
4. **Shell-native first**: GUI supplements CLI, does not replace it
5. **Independent module**: Has its own `go.mod`, can be built separately

## API Endpoints

| Endpoint | Source | Description |
|----------|--------|-------------|
| `GET /api/events` | `.axis/events/tasks.jsonl` | Task event stream |
| `GET /api/runtime` | `.axis/runtime.json` | Runtime metadata |
| `GET /api/providers` | `.axis/providers.json` | Provider profiles |
| `GET /api/mailbox/` | `.axis/comm/` | List actors with mailboxes |
| `GET /api/mailbox/{id}` | `.axis/comm/{id}.jsonl` | Actor's messages |
| `GET /api/skills` | `.axis/skills/` | Available skills |

## Build & Run

```bash
# Build GUI server
cd tools/axis-gui
go build -o axis-gui.exe .

# Run via axis CLI
axis gui [--port 3000]

# Or run directly
./axis-gui.exe --port 3000 --root /path/to/project
```

## Frontend Development

```bash
cd tools/axis-gui/frontend
npm install
npm run dev    # Vite dev server with HMR
npm run build  # Build to dist/
```

## Boundary Rules

- MUST NOT import any `github.com/axis-cli/axis/internal/` package
- MUST NOT write to `.axis/` directory
- MUST NOT modify Axis runtime behavior
- MUST communicate only through file reads and HTTP calls to control server
- MAY be absent — Axis never depends on GUI being available

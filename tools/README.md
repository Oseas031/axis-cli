# Tools

External helper tools that interact with Axis through public CLI surfaces and `.axis/` files. Tools MUST NOT import `internal/` packages.

## axis-up

**Purpose**: External usability helper for Axis setup and onboarding.

**Files**:
- `axis-up.exe` — prebuilt binary

**Usage**:
```bash
.\tools\axis-up\axis-up.exe
```

**Status**: Functional.

## axis-gui

**Purpose**: Web-based observation UI for black-box testing and monitoring Axis runtime state (tasks, actors, events, mailbox).

**Architecture**: Go HTTP server embedding a React (Vite + Tailwind) frontend. Communicates with Axis via:
- Reading `.axis/events/tasks.jsonl` (event stream)
- Reading `.axis/runtime.json` (control server address)
- Calling local control server HTTP API
- Reading `.axis/comm/` (mailbox messages)

**Current State**:
- `axis-gui.exe` — prebuilt binary (10MB, built 2026-05-11)
- `frontend/dist/` — built React app (index.html + assets)
- `frontend/node_modules/` — npm dependencies present
- ⚠️ **Go source code missing** — only compiled exe exists. Source needs to be recovered or rebuilt.

**Frontend Stack**: React + TypeScript + Tailwind CSS + Vite + React Router + Lucide icons

**Design Principles**:
- Non-invasive: does not import Axis internal packages
- Observation only: reads state, does not mutate
- Optional: Axis functions fully without GUI
- Shell-native first: GUI supplements CLI, does not replace it

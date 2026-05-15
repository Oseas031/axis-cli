# MEMORY BOUNDARY — Edit This Directory Only If You Accept These Constraints

The `internal/memory/` tree implements the Axis Layered Memory Model:
Immediate (per-execution) → Working (per-session) → Long-term (immutable history).

## What MEMORY Must NEVER Do

1. **Never push context into provider prompts** — memory is queryable, not injectable. Provider input comes from the contract layer, not memory.
2. **Never mutate scheduler / dispatcher / contract / provider semantics** — memory is observability and retention, not a control plane.
3. **Never grant permissions or execute tools** — memory records what happened; it does not authorize or act.
4. **Never physically delete from Long-term** — forgetting is soft-mark (`deprecated_at`) only. The append-only event log is immutable.
5. **Never introduce external dependencies** — memory is 100% Go standard library: `bufio`, `os`, `encoding/json`, `crypto/sha256`, `sync`, `path/filepath`, `unicode/utf8`, `encoding/binary`. No SQLite, no Redis, no vector DB, no CGO.
6. **Never run background goroutines in P0** — no auto-compact, no threshold triggers, no exit-time cleanup. Compact is CLI-explicit only.
7. **Never inline full file contents by default** — Immediate Memory exposes `path + summary(1024B) + hash + file_changed`. Full content requires explicit tool invocation.

## Three-File KV Invariants (`internal/memory/kv`)

- `history.jsonl` is the **immutable source of truth**. Compact must not truncate, rename, or modify it.
- `snapshot.bin` is a fast cold-start accelerator only. Corruption triggers full history replay, never data loss.
- `index.txt` is snapshot-companion. Missing or corrupt index triggers rebuild by scanning snapshot JSONL.
- In-memory `index map[string]RecordPos` is the runtime authoritative state.
- All line terminators must be LF (`\n`), never `\r\n`, regardless of OS.

## Key Namespace

- Working Memory: `wm:bundle:{bundle_id}` — the only prefix. Never use `wm:packet:*` or `wm:file:*` (P0 stores bundles as self-describing JSON).
- Long-term events: unprefixed; identity is `event_type` + `entity_id` + `timestamp`.

## Before Modifying This Directory

- [ ] Read `docs/specs/layered-memory-model/design.md`
- [ ] Read `docs/specs/layered-memory-model/requirements.md`
- [ ] Confirm: change does not violate append-only semantics of `history.jsonl` or `events.jsonl`
- [ ] Confirm: no new external dependencies
- [ ] Confirm: no background goroutines or auto-triggered maintenance
- [ ] Confirm: cross-platform atomic rename still works on Windows
- [ ] Confirm: UTF-8-safe truncation invariants preserved in Immediate summaries

## Executable Verification

```bash
# No external dependencies (only stdlib imports)
grep -rn "\"github.com/" internal/memory/ --include="*.go" | grep -v "_test.go" | grep -v "axis-cli"
# Expected: 0 lines

# No CRLF in JSONL writes (must use \n only)
grep -rn '\\r\\n\|"\r\n"' internal/memory/ --include="*.go" | grep -v "_test.go"
# Expected: 0 lines

# No background goroutines (no "go func" in non-test code)
grep -rn "go func\|go .*(" internal/memory/ --include="*.go" | grep -v "_test.go"
# Expected: 0 lines

# No os.Open for .axis/ files (should use os.ReadFile for snapshot reads)
grep -rn "os\.Open" internal/memory/ --include="*.go" | grep -v "_test.go" | grep -v "os\.OpenFile"
# Expected: 0 lines (use os.ReadFile for reads, os.OpenFile for append-writes)
```

## Common Traps

| Trap | Why It Is Wrong |
|---|---|
| Truncating `history.jsonl` during Compact | Destroys immutable audit chain; violates append-only core invariant |
| Using platform-native `fmt.Println` for log writes | Emits CRLF on Windows; breaks JSONL parsing |
| Adding SQLite / BoltDB / vector DB | Violates zero-external-dependency constraint |
| Auto-compact on threshold | Violates "explicit only" P0 rule; introduces background logic |
| Physically deleting forgotten events | Violates immutable history principle |
| Storing full file contents in Working Memory values | Should live in file system; Working Memory references by path+hash |
| Using non-LF line terminators | Breaks parser compatibility across platforms |

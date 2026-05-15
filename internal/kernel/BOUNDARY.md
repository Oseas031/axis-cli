# KERNEL BOUNDARY — Edit This Directory Only If You Accept These Constraints

## What KERNEL Must NEVER Do

1. **Never execute Agent logic** — scheduler/dispatcher/orchestrator route and manage tasks; they do not invoke LLMs or run tools directly
2. **Never mutate shared state without event logging** — every state change must leave a trace in `.axis/events/` or task metadata
3. **Never inject contextpack logic into scheduling** — scheduling is pure; context assembly happens in `internal/contextpack/`, not here
4. **Never add daemon/auto-spawn logic** — `axis start` is explicit user action; kernel must not spawn background processes

## Before Modifying This Directory

- [ ] Read `docs/specs/dag-scheduling/design.md` if touching scheduler
- [ ] Read `docs/specs/local-control-plane/design.md` if touching orchestrator/runtime
- [ ] Confirm: change does not alter `GetReadyTasks` claim semantics
- [ ] Confirm: change is observable via CLI or event log (no hidden behavior)
- [ ] Tests include boundary assertions (e.g., "scheduler must not call provider directly")

## Executable Verification

```bash
# Scheduler/dispatcher must not import provider (orchestrator is allowed as it creates executors)
grep -rn '"github.com/.*internal/model"' internal/kernel/scheduler/ internal/kernel/dispatcher/ --include="*.go" | grep -v "_test.go"
# Expected: 0 lines

# Kernel must not import contextpack
grep -rn '"github.com/.*internal/contextpack"' internal/kernel/ --include="*.go" | grep -v "_test.go"
# Expected: 0 lines

# No context.Background() in kernel business logic
grep -rn "context\.Background()" internal/kernel/ --include="*.go" | grep -v "_test.go" | grep -v "main.go"
# Expected: 0 lines
```

## Common Traps

| Trap | Why It Is Wrong |
|---|---|
| Adding provider call in dispatcher | Violates layered isolation; dispatcher routes, contract executor invokes |
| Injecting context assembly in scheduler | Violates scheduling purity; context is external concern |
| Auto-starting background server | Violates explicit `axis start` rule; no hidden daemons |

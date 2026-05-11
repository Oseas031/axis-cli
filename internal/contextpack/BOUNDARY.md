# CONTEXTPACK BOUNDARY — Edit This Directory Only If You Accept These Constraints

## What CONTEXTPACK Must NEVER Do

1. **Never push context into provider prompts** — contextpack is preview-only, opt-in, non-invasive. It assembles bundles; it does NOT inject them into execution
2. **Never change scheduler/contract/dispatcher semantics** — contextpack is a readiness layer, not an execution layer
3. **Never grant permissions or execute tools** — contextpack assembles information; it does not act on it
4. **Never persist to external DB/vector store in P0** — local filesystem only; no external dependencies

## Before Modifying This Directory

- [ ] Read `docs/specs/adaptive-context-assembly/design.md`
- [ ] Read `docs/specs/execution-context-consumption/design.md`
- [ ] Confirm: change is preview-first (dry-run visible before any state mutation)
- [ ] Confirm: change does not auto-inject context into `AgentExecutionRequest` beyond summary-only opt-in

## Common Traps

| Trap | Why It Is Wrong |
|---|---|
| Auto-attaching full bundle to task metadata | Violates lightweight metadata rule; only `context.*` keys allowed |
| Changing assembler to inject into provider prompt | Violates non-invasive boundary; contextpack must not alter execution semantics |
| Adding vector DB dependency | Violates local-first/file-system-native P0 constraint |

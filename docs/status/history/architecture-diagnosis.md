# Architecture Diagnosis & Strategic Direction (2026-05-11)

> Historical analysis. See [current-progress](../current-progress.md) for current gap status.

**Full analysis**: `reports/strategy/agent-native-scenario-principles-analysis-2026-05-11.md`
**First principles**: `docs/architecture/agent-native-first-principles.md`

## Top 8 Core Gaps

| # | Gap | Severity | Status |
|---|---|---|---|
| A | Cross-process context fracture: ReadinessRegistry in-process only | Critical | Partial (Local Control Plane done) |
| B | Orchestrator pseudo-parallel; no inter-Agent collaboration | Critical | Open |
| C | Event log lacks structured query API or feedback loop | High | Open |
| D | ~~Staged Evolution spec-only~~ | ~~Critical~~ | **RESOLVED** (T2-T10 implemented) |
| E | Tool boundaries static fences, not dynamic ladders | High | Open |
| F | Model routing manual; no cost/latency-aware scheduling | High | Partial (T27 cost tracking) |
| G | No Agent identity or capability profile | High | Open |
| H | Execution feedback loop fully broken | High | Open |

## Design Philosophy Assessment

- **Fully applicable**: "More Context, More Action, Zero Control", "bash is all you need", "Interface is existence", "Contract is structure"
- **Partially resolved**: "Query is context" (T8: context.requested_sources)
- **Needs refinement**: "Ladder is boundary" (static → dynamic)
- **Resolved**: "Controllable Evolution" (Staged Evolution P0)
- **M4 validated**: "Competence earns autonomy" (retry, truncation, logging, cost tracking)

## Recommended Priority Order

1. ~~Staged Evolution P0~~ — **COMPLETED**
2. Cross-process state persistence (ReadinessRegistry + Local Control Plane)
3. Agent identity & capability profile (behavioral scoring)
4. Event log structured query (`axis audit`)
5. Dynamic model routing (cost/latency-aware + fallback chains)
6. Execution feedback loop (quality scoring → intent/context assembly)

## Important Reminders

- Coverage stable, all packages passing `go test ./...`
- Follow Occam's Razor principle
- CLI-first / shell-native, no Web UI or heavy TUI
- worktree isolation has known defect (main HEAD), use manual worktree Plan B

# AGENT EXECUTION BOUNDARY — Edit This Directory Only If You Accept These Constraints

## What AGENT EXECUTION Must NEVER Do

1. **Never bypass contract layer** — Agent execution goes through contract executor; no direct provider invocation
2. **Never auto-expand tool permissions** — tool boundaries (`tool.allowed_paths`, `tool.allowed_hosts`) are metadata for audit, not runtime grants
3. **Never inject context metadata into provider input** — `ExecutionContextSummary` is read-only observation, not prompt augmentation in P0
4. **Never escalate autonomy without verification** — competence-based autonomy requires evidence, not assumption

## Before Modifying This Directory

- [ ] Read `docs/specs/execution-context-consumption/design.md`
- [ ] Read `docs/architecture/semantic-boundaries.md` (evolution section)
- [ ] Confirm: change does not alter provider request structure or prompt semantics
- [ ] Confirm: change includes audit test verifying "no injection" boundary

## Executable Verification

```bash
# Agent must not directly import provider (must go through contract)
grep -rn "provider\.Call\|provider\.Execute" internal/agent/ --include="*.go" | grep -v "_test.go" | grep -v "contract"
# Expected: 0 lines

# No context.Background() in agent business logic
grep -rn "context\.Background()" internal/agent/ --include="*.go" | grep -v "_test.go"
# Expected: 0 lines

# No direct prompt augmentation from context summary
grep -rn "ContextSummary.*prompt\|prompt.*ContextSummary" internal/agent/ --include="*.go" | grep -v "_test.go"
# Expected: 0 lines
```

## Common Traps

| Trap | Why It Is Wrong |
|---|---|
| Using context_summary to augment prompt | Violates P0 summary-only, read-only boundary |
| Expanding tool scope based on task content | Tool boundaries are static metadata for audit, not dynamic runtime grants |
| Skipping contract validation for Agent tasks | Contract is structure; all execution must go through contract layer |

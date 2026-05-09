# M5 Workflow Binding

M5 follows the standard workflow routing per `workflow/entry.md`.

## Applicable Workflows

### Feature Implementation / Bug Fix
```
wf-pr-check + wf-ci + wf-doc-006
```

For each implementation task (T1-T22), run:
1. Write/update spec if needed
2. Implement code
3. Test
4. PR check (wf-ci)
5. Update docs (current-progress.md)

## Task Routing

| Task Type | Workflow Combination |
|-----------|---------------------|
| New agent component | wf-pr-check + wf-ci + wf-doc-006 |
| Self-iteration contract | wf-pr-check + wf-ci + wf-doc-006 |
| Integration test | wf-pr-check + wf-ci |
| Documentation only | wf-doc-004 + wf-doc-006 |
| CI/CD changes | wf-ci + wf-doc-004 |

## Execution Order

1. Read `workflow/entry.md`
2. Select workflow combination
3. Read relevant workflow docs
4. Implement minimum change
5. Run validation
6. Update `docs/current-progress.md`

## Coverage Gate

- M5 implementation must maintain ≥85% coverage
- AgentExecutor: 90%+
- BootstrapOrchestrator: 90%+
- Self-iteration contracts: 90%+

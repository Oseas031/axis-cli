# Claude Code Workflow Continuity Guide

**[Chinese version / 中文版](../zh/workflow/claude-code-workflow-continuity-guide.md)**

This guide is for Claude Code / external Agents picking up the Axis project to quickly restore context. The current principle is lightweight, traceable, and not over-automated.

## Handover Sequence

Follow `AGENT_INSTRUCTIONS.md` (the single entry point):

1. Read `CLAUDE.md` — the constitution (all constraints live here).
2. Read `docs/status/current-progress.md` — ground truth for milestone status.
3. Read `docs/architecture/agent-native-first-principles.md` — design principles.
4. Read `workflow/entry.md` — route your task to the minimal workflow combination.
5. Read `HANDOVER.md` — reference: project structure, completed work, known issues.
6. If the task involves CI/CD, read `.github/config/registry.yml` and the corresponding `.github/workflows/*.yml`.

## Handover Synchronization

When a task is completed or a phase changes, synchronize by scope of impact:

| File | When to Update |
|------|---------------|
| `docs/status/current-progress.md` | Current progress, phase status, latest verification results change |
| `HANDOVER.md` | Known issues, next actions, project structure change |
| `docs/README.md` | Documentation index or directory structure change |
| `workflows/README.md` | Workflow status, paths, routing change |

## Workflow Maintenance Rules

1. Do not create a new workflow for a single task.
2. Do not upgrade advisory experience checks into hard gates.
3. GitHub Actions active status must correspond to actually existing `.yml` files.
4. Merged, deleted, garbled, or broken-path workflows must be removed from the active index.
5. Automation scripts are only introduced when benefits are clear and verifiable.
6. Long-term ideas go into reports or deprecated, not into the current execution path.

## Verification Checklist

After workflow-related changes, check at least:

```text
1. workflows/README.md and .github/config/registry.yml status are consistent
2. workflow/entry.md does not reference deprecated workflows
3. Markdown relative links have no broken links
4. Active GitHub Actions files actually exist
5. Document paths match docs/README.md classification structure
```

## Currently Deprecated Items

- `wf-dev`: Local development checks no longer a standalone GitHub Action.
- `wf-release`: Release pipeline merged into `wf-cd`.
- `wf-docs`: Documentation generation merged into `wf-ci`.
- `wf-occams`: Occam's Razor is a built-in constraint principle, no longer a standalone workflow.
- Old Entry Point Workflow: Replaced by `workflow/entry.md`.

## Design Boundary

Workflows only provide context, routing, and verification boundaries—they do not make decisions for Agents. New workflows must simultaneously satisfy:

1. The current task truly cannot reuse existing workflows.
2. There are clear trigger conditions and exit conditions.
3. There is a runnable verification method.
4. No violation of `CLAUDE.md` section 1 (Absolute Prohibitions).

---
description: Meta-Workflow for managing workflows
---

# Meta-Workflow

This project uses a lightweight workflow mechanism. Actual execution is governed by this file and `workflow/entry.md`.

## Authoritative Entry

```text
workflow/entry.md
```

After receiving a task, the Agent reads the entry first, then selects the minimal upstream workflow combination.

## Currently Active Rules

1. **Docs first**: New features must first have `requirements.md`, `design.md`, `tasks.md`, and `workflow-binding.md`.
2. **Minimal workflow**: Prioritize reusing existing workflows; do not add heavy processes.
3. **Gate restraint**: Build, test, and security may block; experience-based checks are reminders only.
4. **Status sync**: In-progress status goes in `docs/status/current-progress.md`; handover status goes in `HANDOVER.md`.
5. **Unique categorization**: During retrospectives, each work item is assigned to only one upstream workflow.
6. **Design sovereignty**: When the user has transferred design sovereignty, the Agent actively organizes the design route and writes documentation; only ask for confirmation before destructive or high-risk operations.

## Project Positioning Change Rules

When the project's core positioning changes:

1. Update the core design philosophy or architecture report first.
2. Then check entry documents as a group: `README.md`, `docs/README.md`, `docs/guides/QUICKSTART.md`, `docs/product/WHITEPAPER.md`, `docs/status/current-progress.md`, `HANDOVER.md`.
3. Clarify what is currently being done, what is not being done, and what the follow-up spec is.
4. Do not automatically expand the current milestone scope with grand designs.

## Currently Disabled Long-Term Ideas

- Prometheus / Grafana workflow monitoring
- Automatic workflow creation/deployment
- Workflow performance testing platform
- Complex rollback system
- Multi-tier version auto-release mechanism
- Independent workflow scheduler

## Parallel Development Isolation Rules

Agent `isolation: "worktree"` has a known flaw: the created worktree is based on the default branch `main` HEAD, not the current branch HEAD. For parallel development, use manual worktree:

```bash
# Create (from current commit)
git worktree add -b <branch-name> .claude/worktrees/<name> <commit>

# Use (main session enters via EnterWorktree --path)

# Clean up
git worktree remove --force .claude/worktrees/<name>
git branch -D <branch-name>
```

The same branch cannot be checked out in two worktrees simultaneously; a new branch must be created based on the commit SHA.

## Document Sync Checklist

Progress updates must sync the following 4 files:

| File | Update Content |
|---|---|
| `CLAUDE.md` | Current Status + Architecture + Runtime Flow + Constraints/Defects |
| `HANDOVER.md` | Handover status + coverage + known issues + next actions |
| `AGENT_INSTRUCTIONS.md` | Current status summary |
| `docs/status/current-progress.md` | Completed / In Progress / Pending + recent commits |

## Relationship to Design Philosophy

- **More Context**: Workflows provide context and routing.
- **More Action**: Workflows indicate the next executable action.
- **Zero Control**: Workflows do not make decisions for the Agent or create excessive blocking.
- **Bash is All You Need**: Prioritize implementing processes with CLI, scripts, and simple documents.


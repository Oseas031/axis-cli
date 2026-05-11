# Milestone 2 Workflow Binding

## Purpose

This document binds the Milestone 2 specification to the existing Axis workflow system so development follows project workflow mechanics instead of ad-hoc coding.

Milestone 2 is treated as a new feature track by `workflow/entry.md`, so it must use the new-feature workflow route and keep specs, implementation, verification, and handoff synchronized.

## Upstream Workflows

### wf-doc-004: Meta-Workflow Management

**File**: `workflow/meta-workflow-management.md`

**Role**:

- Treat `docs/specs/milestone2/` as the implementation contract
- Keep requirements, design, tasks, code, and handoff documents synchronized
- Ensure explicit workflow dependencies are visible before implementation
- Avoid creating new workflow machinery unless existing workflows cannot express the task

**Applied rules**:

- Documentation before implementation
- Explicit binding through this `workflow-binding.md`
- Build/test/security issues may block; advisory documentation checks should not become excessive control gates
- HANDOVER must be updated after material progress

### Scope-Control Rule: Occam's Razor

**Source**: embedded in `workflow/entry.md` and `workflow/meta-workflow-management.md`

**Role**:

- Keep Milestone 2 focused on the minimum viable parallel scheduling slice
- Prevent Milestone 3 scope from leaking into this milestone
- Require justification before adding new UI, new automation, new external dependencies, or new workflow mechanisms

**Applied rules**:

- Reuse existing `AgentTask.Dependencies` instead of introducing a separate graph database
- Add an additive ready-set scheduler API instead of rewriting the scheduler as a full graph engine
- Use Go standard library goroutines/channels for local parallelism
- Keep CLI/shell-native validation as the default interface

### wf-pr-check: PR Quality Check Workflow

**File**: `.github/workflows/pr-check-workflow.yml`

**Role**:

- Validate code quality and PR readiness after implementation tasks
- Provide documentation-context reminders without turning every design note into a hard block

**Applied rules**:

- `go test ./...` must pass before PR readiness
- scheduler/orchestrator changes must include targeted tests
- docs touched by implementation must remain synchronized with code behavior

### wf-ci: Continuous Integration Workflow

**File**: `.github/workflows/ci.yml`

**Role**:

- Validate build, tests, formatting, vet, and static analysis in CI

**Applied rules**:

- `gofmt` must be applied to Go changes
- `go test ./...` must pass
- `go build` must pass without overwriting existing local artifacts unless intentionally requested

### wf-doc-006: Document Audit

**File**: `.github/workflows/document-audit.yml`

**Role**:

- Ensure Milestone 2 specs and handoff documents remain discoverable and consistent
- Check that links and milestone status are not stale

**Applied rules**:

- `requirements.md`, `design.md`, `tasks.md`, and `workflow-binding.md` must cross-link
- `docs/status/current-progress.md`, `HANDOVER.md`, and `AGENT_INSTRUCTIONS.md` must reflect the same milestone state
- Deprecated workflow paths must not be used as authoritative sources

## Feature Workflow Execution Order

```text
1. workflow/entry.md -> route as New Feature
2. wf-doc-004       -> confirm spec contract and explicit dependencies
3. scope check      -> confirm minimal Milestone 2 scope
4. requirements.md  -> confirm WHAT with user
5. design.md        -> confirm HOW with user
6. tasks.md         -> confirm execution plan with user
7. implement        -> execute tasks in order
8. wf-ci           -> run gofmt/build/test validation
9. wf-pr-check     -> quality and PR readiness review
10. wf-doc-006     -> documentation consistency and handoff update
```

## Explicit Dependencies

| Dependency | Reason |
|---|---|
| `workflow/entry.md` | Authoritative route for selecting workflow combination |
| `workflow/meta-workflow-management.md` | Requires spec-first, explicit binding, handoff sync |
| `workflows/README.md` | Provides active workflow IDs and avoids deprecated workflow use |
| `.github/config/registry.yml` | Registry source for workflow status and IDs |
| `docs/specs/interactive-shell/` | Shell remains a CLI-native validation path |
| `docs/specs/model-provider/` | Existing execution-path spec must not be contradicted by Milestone 2 |
| `docs/status/acceptance/milestone1-acceptance-report.md` | Baseline proof that Milestone 1 behavior must remain compatible |

## Non-Goals Enforced by Workflow

- No new workflow for Milestone 2 unless current routing is proven insufficient
- No separate Occam workflow; scope control is enforced by `workflow/entry.md` and `workflow/meta-workflow-management.md`
- No Web UI or heavy TUI
- No external database-backed scheduler
- No distributed worker system
- No tool calling layer
- No real model provider integration
- No global event bus
- No policy-heavy control plane

## Completion Criteria

Milestone 2 is complete only when:

- `requirements.md`, `design.md`, `tasks.md`, and `workflow-binding.md` are confirmed and consistent
- `tasks.md` progress table is fully marked Completed
- targeted scheduler, admission, SLA, and orchestrator tests pass
- `go test ./...` passes
- build validation passes without unintended artifact overwrites
- `docs/status/current-progress.md`, `HANDOVER.md`, and `AGENT_INSTRUCTIONS.md` are updated
- no Milestone 3 scope has been introduced




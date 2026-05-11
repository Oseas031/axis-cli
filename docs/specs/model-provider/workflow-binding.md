# Model Provider Workflow Binding

## Purpose

This document binds the Model Provider feature to the existing Axis workflow system so implementation follows project workflow mechanics rather than ad-hoc coding.

## Upstream Workflows

### wf-doc-004: Meta-Workflow Management

**File**: `workflow/meta-workflow-management.md`

**Role**:

- Treat `docs/specs/model-provider/` as the contract for implementation
- Keep requirements, design, tasks, code, and HANDOVER synchronized
- Track explicit dependencies and lifecycle state

**Applied rules**:

- Documentation before implementation
- Explicit dependency declaration
- Implementation must update task status
- HANDOVER must be updated after completion

### Scope-Control Rule: Occam's Razor

**Source**: embedded in `workflow/entry.md` and `workflow/meta-workflow-management.md`

**Role**:

- Keep Model Provider minimal
- Implement `MockModelProvider` before real providers
- Avoid API keys, network calls, streaming, routing, or Web UI in this slice

**Applied rules**:

- Minimum viable feature only
- Progressive enhancement
- No Milestone 2+ expansion inside this task

### wf-pr-check: PR Quality Check Workflow

**File**: `.github/workflows/pr-check-workflow.yml`

**Role**:

- Build and test code changes
- Provide non-blocking documentation context reminders
- Preserve code quality gates

**Applied rules**:

- `go test ./...` must pass
- Shell behavior must remain stable
- Documentation context should be updated when implementation changes

### wf-ci: Continuous Integration Workflow

**File**: `.github/workflows/ci.yml`

**Role**:

- Validate build, tests, format, vet, and static analysis in CI

**Applied rules**:

- `go build` must succeed
- `go test ./...` must succeed
- `gofmt` must be applied

### wf-doc-006: Document Audit

**File**: `.github/workflows/document-audit.yml`

**Role**:

- Ensure documentation remains discoverable and consistent

**Applied rules**:

- Beginner guide must explain mock provider clearly
- HANDOVER must record completion
- Spec task status must be updated

## Feature Workflow Execution Order

```text
1. wf-doc-004  -> confirm spec exists and dependencies are explicit
2. scope check  -> confirm scope is MockModelProvider only
3. implement   -> follow tasks.md T1-T7
4. wf-ci       -> run gofmt, go build, go test ./...
5. wf-pr-check -> verify shell path and docs context
6. wf-doc-006  -> update docs and HANDOVER
```

## Explicit Dependencies

| Dependency | Reason |
|---|---|
| interactive-shell spec | Shell is the user-facing validation path |
| default contract | Mock provider runs behind the default contract path |
| contract executor | Provider execution belongs behind contract validation |
| dispatcher | Dispatcher must return provider-backed task results |

## Non-Goals Enforced by Workflow

- No real model provider
- No API key configuration
- No external SDK
- No Web UI
- No provider routing
- No prompt template system

## Completion Criteria

The feature is complete only when:

- `tasks.md` is fully marked Completed
- `go test ./...` passes
- `axis shell` can run a task through the mock provider path
- `docs/guides/BEGINNER_GUIDE.md` explains the mock provider
- `HANDOVER.md` records the feature



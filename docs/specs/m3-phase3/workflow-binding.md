# M3 Phase 3 Workflow Binding

## Upstream Workflows

This feature follows the project's existing workflow system:

  **wf-doc-004** Meta Workflow Management: Documentation first, explicit dependencies, HANDOVER sync
  **wf-occams** Occam's Razor: No external dependencies, Go stdlib only, build only what's currently needed
  **wf-pr-check** PR Quality Check: Quality gate + non-blocking documentation context
  **wf-ci** Continuous Integration: build, format, race tests, ≥85% coverage
  **wf-doc-006** Document Audit: HANDOVER consistency

## Phase 3 Special Workflow Constraints

  SLA strategy engine and tool invocation layer can be **developed in parallel** (no code dependency), merging at T10
  Both parts share modifications to `internal/types/types.go` (T1 + T5), type changes should be merged first
  BashTool uses `os/exec`, Go stdlib built-in, no external dependencies
  Multi-turn execution loop max turns hardcoded to 10, no configuration layer introduced

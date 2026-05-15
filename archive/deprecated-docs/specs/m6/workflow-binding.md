# M6 Workflow Binding

## Workflow Routing

Based on `workflow/entry.md`, M6 work is classified as:

**Type**: Feature/New Spec
**Workflow**: `wf-doc-004` + `wf-pr-check` + `wf-ci` + `wf-doc-006`

## Workflow Execution Order

```
wf-doc-004 (Specification)
    → wf-occams (Architecture Simplification Review)
    → wf-pr-check (PR Quality Check)
    → wf-ci (CI Pipeline)
    → wf-doc-006 (Documentation)
```

## wf-doc-004 Tasks

1. Review M6 requirements in `docs/specs/m6/requirements.md`
2. Validate design against requirements
3. Ensure non-goals are clearly defined
4. Check for duplicate functionality with existing M1-M5

## Scope-Control Tasks

1. Review M6 design for unnecessary complexity
2. Verify SelfJudgement is truly needed (not already in M5)
3. Check strategy pattern is appropriate
4. Verify integration points are minimal

## wf-pr-check Tasks

1. Format: `gofmt -w .`
2. Vet: `go vet ./...`
3. Staticcheck: `staticcheck ./...`
4. CyclComplexity: `gocyclo -over 15 .`
5. Security: `gosec ./...`
6. Test: `go test -race ./...` with ≥85% coverage
7. Build: `go build -o axis-dev.exe cmd/axis/main.go`

## wf-ci Tasks

1. All wf-pr-check tasks
2. Cross-platform build verification
3. Documentation generation (if applicable)

## wf-doc-006 Tasks

1. Update `docs/status/current-progress.md` with M6 status
2. Update `docs/specs/m6/` status to Complete
3. Update HANDOVER.md with M6 completion
4. Update CLAUDE.md if architecture changes

## Phase-Specific Binding

| Phase | wf-doc-004 | Scope Control | wf-pr-check | wf-ci | wf-doc-006 |
|-------|-------------|-----------|-------------|-------|------------|
| 6.1 | T1-T5 | Review | Test | Build | Update |
| 6.2 | T6-T10 | Review | Test | Build | Update |
| 6.3 | T11-T12 | Review | Test | Build | Update |
| 6.4 | T13-T15 | Review | Test | Build | Update |
| 6.5 | T16-T18 | Review | Test | Build | Update |



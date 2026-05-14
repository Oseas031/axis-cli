# Staged Evolution Protocol Tasks

## Related Documents

- [requirements.md](requirements.md)
- [design.md](design.md)
- [../../architecture/spec-lifecycle-conventions.md](../../architecture/spec-lifecycle-conventions.md)
- [../adaptive-context-assembly/](../adaptive-context-assembly/)
- [../execution-context-consumption/](../execution-context-consumption/)

## Progress Tracking

| Task | Status |
|---|---|
| T1: Create Staged Evolution spec boundary | Completed |
| T2: Define evolution data models | Completed |
| T3: Implement project-local evolution store | Completed |
| T4: Implement isolated workspace creation | Completed |
| T5: Implement trace ledger append/read | Completed |
| T6: Implement verification record capture | Completed |
| T7: Implement inspect command | Completed |
| T8: Implement explicit promote/discard gates | Completed |
| T9: Add safety and regression tests | Completed |
| T10: Document operational workflow | Completed |

---

## T1: Create Staged Evolution spec boundary

**Goal**: Define the engineering safety envelope for Axis self-evolution.

**Files**:

- `docs/specs/sandboxed-evolution/requirements.md`
- `docs/specs/sandboxed-evolution/design.md`
- `docs/specs/sandboxed-evolution/tasks.md`
- `docs/README.md`

**Acceptance Criteria**:

- Requirements, design, and tasks exist.
- P0 scope is limited to isolation, atomic steps, verification, promotion/discard, and trace ledger.
- Non-goals exclude uncontrolled self-modification, random perturbation, and prompt mutation.
- The spec is linked from `docs/README.md`.

**Depends on**: Axis system conventions and current context readiness acceptance

---

## T2: Define evolution data models

**Goal**: Add internal data structures for evolution runs, steps, verification, and decisions.

**Files**:

- `internal/evolution/types.go`
- `internal/evolution/types_test.go`

**Design References**:

- `design.md#data-model-sketch`

**Acceptance Criteria**:

- `EvolutionIntent`, `EvolutionRun`, `EvolutionStep`, `VerificationRecord`, and `EvolutionDecision` exist.
- Status values are explicit and stable.
- JSON serialization uses machine-friendly field names.
- Tests cover zero-value safety and JSON round trips.

**Depends on**: T1

---

## T3: Implement project-local evolution store

**Goal**: Persist evolution run records under `.axis/evolution/` without adding a database.

**Files**:

- `internal/evolution/store.go`
- `internal/evolution/store_test.go`

**Design References**:

- `design.md#project-local-layout`

**Acceptance Criteria**:

- Store can create and read run directories.
- Store writes `intent.json` and `run.json` atomically.
- Store preserves existing run records.
- Tests use temporary directories.

**Depends on**: T2

---

## T4: Implement isolated workspace creation

**Goal**: Create an isolated workspace for draft changes.

**Files**:

- `internal/evolution/workspace.go`
- `internal/evolution/workspace_test.go`

**Design References**:

- `design.md#isolation-before-influence`
- `design.md#project-local-layout`

**Acceptance Criteria**:

- Workspace is created under the run directory.
- Workspace creation does not mutate main project files.
- Existing workspace is not silently overwritten.
- Tests prove main-tree files remain unchanged.

**Depends on**: T3

---

## T5: Implement trace ledger append/read

**Goal**: Record ordered evolution steps in an append-only ledger.

**Files**:

- `internal/evolution/ledger.go`
- `internal/evolution/ledger_test.go`

**Design References**:

- `design.md#step-trace-ledger`
- `design.md#data-model-sketch`

**Acceptance Criteria**:

- Steps append to `steps.jsonl`.
- Steps can be read in order.
- Malformed ledger entries return clear errors.
- Tests cover append/read and non-mutating inspection.

**Depends on**: T3

---

## T6: Implement verification record capture

**Goal**: Capture verification command evidence for an evolution run.

**Files**:

- `internal/evolution/verify.go`
- `internal/evolution/verify_test.go`

**Design References**:

- `design.md#verification-gate`

**Acceptance Criteria**:

- Verification records command, timing, exit code, and status.
- Output is stored by reference or bounded content.
- Failed verification is represented without panics.
- Tests use deterministic local commands only.

**Depends on**: T3

---

## T7: Implement inspect command

**Goal**: Expose read-only inspection for evolution runs.

**Files**:

- `cmd/axis/evolve_cmd.go`
- `cmd/axis/evolve_cmd_test.go`

**Design References**:

- `design.md#cli-shape`
- `requirements.md#fr6-read-only-inspection`

**Acceptance Criteria**:

- `axis evolve inspect <run-id>` renders machine-friendly JSON or stable structured output.
- Inspect does not mutate the run.
- Missing run returns a clear error.
- Tests cover successful and missing inspection.

**Depends on**: T3, T5

---

## T8: Implement explicit promote/discard gates

**Goal**: Add explicit lifecycle decisions for verified or abandoned evolution runs.

**Files**:

- `internal/evolution/decision.go`
- `internal/evolution/decision_test.go`
- `cmd/axis/evolve_cmd.go`
- `cmd/axis/evolve_cmd_test.go`

**Design References**:

- `design.md#promotion-semantics`
- `design.md#discard-semantics`

**Acceptance Criteria**:

- Promotion requires a successful verification record.
- Discard preserves trace files.
- Decisions are written to `decision.json`.
- Tests cover missing verification, failed verification, promotion, and discard.

**Depends on**: T6, T7

---

## T9: Add safety and regression tests

**Goal**: Prove the protocol does not change existing execution semantics.

**Files**:

- `internal/evolution/*_test.go`
- `cmd/axis/*_test.go`
- existing dispatcher/context tests if affected

**Design References**:

- `requirements.md#fr7-no-hidden-execution-semantics`

**Acceptance Criteria**:

- Existing `axis ask`, `axis run`, shell, scheduler, dispatcher, contract executor, provider, and contextpack tests still pass.
- Evolution workspace operations do not mutate unrelated files.
- Promotion/discard is explicit.
- `go test ./...` passes.

**Depends on**: T8

---

## T10: Document operational workflow

**Goal**: Document how humans and agents should use staged evolution.

**Files**:

- `docs/specs/sandboxed-evolution/design.md`
- possibly `docs/guides/` or `docs/architecture/` follow-up docs

**Design References**:

- `design.md#cli-shape`

**Acceptance Criteria**:

- Documentation explains start, inspect, verify, promote, and discard flow.
- Documentation states P0 safety boundaries.
- Documentation explains how this protocol supports controllable self-evolution.

**Depends on**: T9

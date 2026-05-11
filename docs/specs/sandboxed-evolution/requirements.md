# Sandboxed Evolution Protocol Requirements

## Summary

Define a safe engineering protocol for Axis self-evolution. The protocol ensures that contract, workflow, spec, context rule, and future control-logic changes can be proposed and verified without directly contaminating the main system.

This spec turns Axis self-causation into an observable, controllable, and reversible feedback loop.

## Design Philosophy

The protocol must follow Axis's design principles:

- **More Context**: every evolution attempt must preserve enough trace data to understand intent, changes, verification, and outcome.
- **More Action**: agents should be able to propose and test system changes, not only describe them.
- **Zero Control**: the core system should not hide or centralize decisions; promotion must be explicit and auditable.
- **Controllable Evolution**: unsafe or unverified changes must remain isolated until accepted.
- **Competence earns autonomy**: higher-risk promotion rights must be earned by reliable execution history, not identity.
- **Bash is all you need**: the first implementation should be CLI/file-system native, scriptable, and inspectable.

## Users

- Developers evolving Axis safely.
- Agents proposing contract or workflow improvements.
- Maintainers reviewing and promoting verified changes.
- Future autonomy layers that need reliable evidence of competence.

## Functional Requirements

### FR1: Isolated evolution workspace

Axis must support creating an isolated workspace for a proposed system change.

The workspace must keep draft changes separate from the main working tree until explicitly promoted.

### FR2: Atomic evolution steps

An evolution attempt must be represented as ordered, atomic steps.

Each step must have:

- step ID
- target path or target domain
- action summary
- before/after evidence or patch reference
- status
- verification result when applicable

### FR3: Verification before promotion

Axis must define a verification gate before changes can be promoted.

P0 verification may be command-based and deterministic, such as:

```text
go test ./...
```

The protocol must not silently promote unverified changes.

### FR4: Promotion and discard decisions

Axis must distinguish between:

- draft change
- verified change
- promoted change
- discarded change

Promotion and discard decisions must be explicitly recorded.

### FR5: Step trace ledger

Axis must record an append-only trace ledger for each evolution attempt.

The ledger must be machine-readable and suitable for shell tooling.

P0 should prefer JSON Lines or small JSON files under a project-local Axis directory.

### FR6: Read-only inspection

Axis must support inspecting an evolution attempt without mutating it.

Inspection must show:

- intent
- target
- current status
- steps
- verification status
- promotion or discard decision if any

### FR7: No hidden execution semantics

The protocol must not change default scheduler, dispatcher, contract executor, provider, contextpack, or prompt behavior.

It is a safety envelope around system evolution, not a new execution engine.

## Acceptance Criteria

- [x] Requirements, design, and tasks exist under `docs/specs/sandboxed-evolution/`.
- [x] `docs/README.md` links to the spec.
- [x] P0 scope is limited to isolation, atomic steps, verification, promotion/discard, and trace ledger.
- [x] Non-goals explicitly exclude random perturbation, automatic prompt mutation, automatic architecture self-modification, and uncontrolled parallel candidate merge.
- [x] The spec aligns with Axis system conventions and existing acceptance boundaries.

## Constraints

- P0 must be local-first and file-system native.
- P0 must be deterministic and testable.
- P0 must avoid network dependencies.
- P0 must not require a database.
- P0 must not automatically run destructive commands.
- P0 must not bypass human or policy review for high-risk changes.

## Non-Goals

- No automatic control-logic rewriting in P0.
- No random perturbation injection in P0.
- No adaptive feedback controller in P0.
- No parallel candidate merge in P0.
- No persistent global memory service in P0.
- No Web UI or TUI.
- No prompt injection or provider request mutation.
- No automatic promotion of unverified changes.

## Open Questions

1. Should the isolated workspace use a copied directory, git worktree, or patch-only staging area in P0?
2. Should verification commands be configured per evolution attempt or constrained to a small allowlist?
3. Should promotion require human confirmation in P0, or can low-risk documentation-only changes be promoted automatically after verification?
4. How should future competence scores influence promotion authority without introducing hidden control?

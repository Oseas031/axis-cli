# Natural Language Scheduling Tasks

## Related Documents

- [requirements.md](requirements.md)
- [design.md](design.md)

## Progress Tracking

| Task | Status |
|---|---|
| T1: Document feature boundary | Completed |
| T2: Add deterministic intent parser | Completed |
| T3: Add `axis ask` dry-run command | Completed |
| T4: Add explicit submit path | Completed |
| T5: Add shell `ask` integration | Completed |
| T6: Add optional LLM parser design gate | Planned |

---

## T1: Document feature boundary

**Goal**: Define natural language scheduling as intent-to-task compilation, not a chatbot or control plane.

**Files**:

- `docs/specs/natural-language-scheduling/requirements.md`
- `docs/specs/natural-language-scheduling/design.md`
- `docs/specs/natural-language-scheduling/tasks.md`
- `docs/specs/interactive-shell/requirements.md`
- `docs/specs/interactive-shell/design.md`
- `docs/specs/interactive-shell/tasks.md`
- `docs/README.md`

**Acceptance Criteria**:

- Existing interactive shell scope remains intact.
- Natural language scheduling is listed as a future extension.
- No runtime code changes are required for this task.

**Depends on**: None

---

## T2: Add deterministic intent parser

**Goal**: Add a minimal parser that converts a prompt into a task proposal without using an LLM.

**Files**:

- `internal/intent/parser.go`
- `internal/intent/parser_test.go`

**Acceptance Criteria**:

- Parser returns a `types.AgentTask` proposal.
- Original prompt is stored in input and metadata.
- Parser mode is recorded as deterministic.
- Tests cover empty prompt, default contract, explicit contract, and explicit task ID.

**Depends on**: T1

---

## T3: Add `axis ask` dry-run command

**Goal**: Add a CLI command that renders the proposed task without submitting by default.

**Files**:

- `cmd/axis/main.go` or a new command file under `cmd/axis`
- `cmd/axis/*_test.go`

**Acceptance Criteria**:

- `axis ask "..."` prints the task proposal.
- `axis ask --stdin` reads prompt text from stdin.
- Existing `run`, `status`, `shell`, and provider commands continue to work.

**Depends on**: T2

---

## T4: Add explicit submit path

**Goal**: Allow natural language tasks to be submitted only when explicitly requested.

**Files**:

- `cmd/axis/*`
- `internal/intent/*`

**Acceptance Criteria**:

- `axis ask "..." --submit` submits an ordinary `AgentTask`.
- Output includes submitted task ID and suggested status command.
- Dry-run remains available.

**Depends on**: T3

---

## T5: Add shell `ask` integration

**Goal**: Let `axis shell` call the same parser path.

**Files**:

- `cmd/axis/main.go` or extracted shell command handler file
- shell tests

**Acceptance Criteria**:

- `axis> ask <prompt>` works in shell.
- Shell implementation reuses the same parser as `axis ask`.
- Shell remains usable after parse or submit errors.

**Depends on**: T4

---

## T6: Add optional LLM parser design gate

**Goal**: Define requirements before allowing LLM-backed intent parsing.

**Files**:

- `docs/specs/natural-language-scheduling/design.md`
- future parser implementation files

**Acceptance Criteria**:

- LLM output must be schema-validated before submission.
- Failure falls back to deterministic parser or dry-run error.
- Prompt, provider profile, parser mode, and parsed structure remain observable.
- No direct shell execution from raw natural language.

**Depends on**: T5

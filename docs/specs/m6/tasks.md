# M6 Tasks: Self-Judgement

**Priority**: P0 = Critical, P1 = Important, P2 = Nice to have

## Phase 6.1: Core Framework

| ID | Task | Priority | Status | Notes |
|----|------|----------|--------|-------|
| T1 | JudgementCriteria data structure | P0 | ✅ Done | |
| T2 | JudgementResult data structure | P0 | ✅ Done | |
| T3 | JudgementItem data structure | P0 | ✅ Done | |
| T4 | ValidationStrategy interface | P0 | ✅ Done | |
| T5 | SelfJudgementEngine core | P0 | ✅ Done |

## Phase 6.2: Built-in Strategies

| ID | Task | Priority | Status | Notes |
|----|------|----------|--------|-------|
| T6 | SyntaxValidationStrategy | P1 | ✅ Done | Go fmt/vet |
| T7 | SemanticValidationStrategy | P1 | ✅ Done | LLM-based |
| T8 | ContractValidationStrategy | P1 | ✅ Done | |
| T9 | TestValidationStrategy | P1 | ✅ Done | |
| T10 | CoverageValidationStrategy | P1 | ✅ Done | |

## Phase 6.3: Self-Judgement Contract

| ID | Task | Priority | Status | Notes |
|----|------|----------|--------|-------|
| T11 | `self/judge-execution` contract | P0 | ✅ Done | |
| T12 | Contract registration | P0 | ✅ Done | |

## Phase 6.4: Integration

| ID | Task | Priority | Status | Notes |
|----|------|----------|--------|-------|
| T13 | AgentExecutionResult + JudgementResult | P0 | ✅ Done | JudgementResult field added to AgentExecutionResult |
| T14 | BootstrapOrchestrator judgement support | P1 | ⏳ Pending | |
| T15 | CLI judgement result display | P2 | ⏳ Pending | |

## Phase 6.5: Testing & Documentation

| ID | Task | Priority | Status | Notes |
|----|------|----------|--------|-------|
| T16 | Unit tests for each strategy | P1 | ⏳ Pending | |
| T17 | Integration tests | P1 | ⏳ Pending | |
| T18 | Documentation update | P2 | ⏳ Pending | |

## Dependencies

- T2, T3 depend on T1
- T4 depends on T1, T2, T3
- T5 depends on T4
- T6-T10 depend on T4
- T11 depends on T1, T2, T3, T5
- T12 depends on T11
- T13 depends on T11, T12
- T14 depends on T13
- T15 depends on T14
- T16 depends on T6-T10, T5
- T17 depends on T14, T15, T16
- T18 depends on T17

## Coverage Targets

| Component | Target |
|-----------|--------|
| SelfJudgementEngine | 90%+ |
| ValidationStrategy implementations | 90%+ |
| Judgement contract | 100% |

## Estimated Timeline

- Phase 6.1-6.3: 1-2 days
- Phase 6.4-6.5: 1 day
- Total: ~3 days

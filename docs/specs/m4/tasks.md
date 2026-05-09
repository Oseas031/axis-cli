# M4 Tasks

**Priority**: P1 = Critical, P2 = Important, P3 = Nice to have

## Phase 4.1: Provider Infrastructure ✅

| ID | Task | Priority | Status |
|----|------|----------|--------|
| T1 | Add ProviderConfig + functional options | P1 | ✅ Done |
| T2 | Implement AnthropicProvider | P1 | ✅ Done |
| T3 | Implement OpenAIProvider | P1 | ✅ Done |
| T4 | Add token accounting to ModelResponse | P1 | ✅ Done |
| T5 | Add safe JSON serialization wrapper | P1 | ✅ Done |

## Phase 4.2: Tool Extensions ✅

| ID | Task | Priority | Status |
|----|------|----------|--------|
| T6 | Implement FileReadTool | P2 | ✅ Done |
| T7 | Implement FileWriteTool | P2 | ✅ Done |
| T8 | Implement HTTPClientTool | P2 | ✅ Done |
| T9 | Add tool permission scopes | P2 | ✅ Done |
| T10 | Add circuit breaker to executor | P2 | ✅ Done |

## Phase 4.3: Integration & Testing ✅

| ID | Task | Priority | Status |
|----|------|----------|--------|
| T11 | Wire new providers into orchestrator | P1 | ✅ Done |
| T12 | Add provider tests | P1 | ✅ Done |
| T13 | Add tool tests | P1 | ✅ Done |
| T14 | Update CLI to support provider selection | P2 | ✅ Done |
| T15 | Update shell to show available tools | P2 | ✅ Done |

## Phase 4.4: Documentation 🚧

| ID | Task | Priority | Status |
|----|------|----------|--------|
| T16 | Update CLAUDE.md with M4 status | P2 | 🚧 In Progress |
| T17 | Update current-progress.md | P2 | ⏳ Pending |
| T18 | Add provider usage examples | P3 | ⏳ Pending |

## Dependencies

- T2, T3 depend on T1
- T6, T7, T8 depend on T9
- T11 depends on T2, T3
- T12 depends on T2, T3, T4, T5
- T13 depends on T6, T7, T8, T5
- T14 depends on T11
- T15 depends on T13
- T16 depends on all implementation tasks

## Coverage Targets

| Component | Target | Actual |
|----------|--------|--------|
| Provider implementations | 90%+ | 91.8% ✅ |
| Tool implementations | 90%+ | 93.7% ✅ |
| Executor changes | 85%+ | 95.1% ✅ |

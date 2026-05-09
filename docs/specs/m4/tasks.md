# M4 Tasks

**Priority**: P1 = Critical, P2 = Important, P3 = Nice to have

## Phase 4.1: Provider Infrastructure

| ID | Task | Priority | Notes |
|----|------|----------|-------|
| T1 | Add ProviderConfig + functional options | P1 | Extend registry.go with config struct |
| T2 | Implement AnthropicProvider | P1 | net/http only, no SDK |
| T3 | Implement OpenAIProvider | P1 | net/http only, no SDK |
| T4 | Add token accounting to ModelResponse | P1 | InputTokens, OutputTokens fields |
| T5 | Add safe JSON serialization wrapper | P1 | Prevent panics in tool result marshal |

## Phase 4.2: Tool Extensions

| ID | Task | Priority | Notes |
|----|------|----------|-------|
| T6 | Implement FileReadTool | P2 | Path validation against allowedDirs |
| T7 | Implement FileWriteTool | P2 | Path validation, create directories |
| T8 | Implement HTTPClientTool | P2 | GET/POST/PUT/DELETE, header support |
| T9 | Add tool permission scopes | P2 | filesystem:read, filesystem:write, network |
| T10 | Add circuit breaker to executor | P2 | Max 5 consecutive tool errors |

## Phase 4.3: Integration & Testing

| ID | Task | Priority | Notes |
|----|------|----------|-------|
| T11 | Wire new providers into orchestrator | P1 | WithModelProvider option |
| T12 | Add provider tests | P1 | Unit tests for Anthropic/OpenAI |
| T13 | Add tool tests | P1 | Unit tests for FileTools, HTTPClientTool |
| T14 | Update CLI to support provider selection | P2 | `--provider` flag |
| T15 | Update shell to show available tools | P2 | `tools` command |

## Phase 4.4: Documentation

| ID | Task | Priority | Notes |
|----|------|----------|-------|
| T16 | Update CLAUDE.md with M4 status | P2 | |
| T17 | Update current-progress.md | P2 | |
| T18 | Add provider usage examples | P3 | README update |

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

- Provider implementations: 90%+
- Tool implementations: 90%+
- Executor changes: 85%+

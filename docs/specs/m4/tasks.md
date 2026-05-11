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

## Phase 4.4: Documentation ✅

| ID | Task | Priority | Status |
|----|------|----------|--------|
| T16 | Update CLAUDE.md with M4 status | P2 | ✅ Done |
| T17 | Update current-progress.md | P2 | ✅ Done |
| T18 | Add provider usage examples | P3 | ✅ Done |

## Dependencies

- T2, T3 depend on T1
- T6, T7, T8 depend on T9
- T11 depends on T2, T3
- T12 depends on T2, T3, T4, T5
- T13 depends on T6, T7, T8, T5
- T14 depends on T11
- T15 depends on T13
- T16 depends on all implementation tasks

## Phase 4.5: CLI Usability Hardening (Post-Completion Gap Fix)

这些任务在 M4 原始 spec 完成后，通过实际端到端使用发现的设计-实现断层，作为查漏补缺批次补充。

| ID | Task | Priority | Status |
|----|------|----------|--------|
| T19 | CLI env fallback：`providerOptions()` 在无 active profile 时回退到 `ANTHROPIC_API_KEY`/`OPENAI_API_KEY`/`DEEPSEEK_API_KEY`/`MINIMAX_API_KEY` | P1 | ✅ Done |
| T20 | 默认模型修正：`deepseek-chat`（已弃用）→ `deepseek-v4-flash`；`MiniMax-Text-01`（非兼容 API）→ `MiniMax-M2.7` | P1 | ✅ Done |
| T21 | 空 key 早期诊断：`NewProvider` 在构造阶段检测空 `apiKey`，返回可操作错误引导 | P1 | ✅ Done |
| T22 | 回归测试：`TestEnvAPIKeyForProvider`、`TestProviderOptions_EnvFallback`、`TestNewProvider_MissingAPIKey`、默认模型断言扩展 | P1 | ✅ Done |

## Phase 4.6: Non-Destructive Provider & Tool Hardening

遵循 Axis 设计哲学（`bash is all you need`、`More Context, More Action, Zero Control`、`Competence earns autonomy`），在不影响接口和调度语义的前提下增强可观测性、可靠性和边界安全。

| ID | Task | Priority | Status |
|----|------|----------|--------|
| T23 | `axis provider test` 诊断命令：轻量 ping 验证 API key/网络/模型名有效性 | P1 | ✅ Done |
| T24 | 指数退避重试：`providerConfig.maxRetries` 当前未使用，在 Anthropic/OpenAI Execute 中实现（仅 5xx/timeout 重试，4xx 不重试） | P1 | ✅ Done |
| T25 | file_read / http_request 输出截断与大小限制：防止大文件/大响应撑爆 context（BashTool 已有 64 KiB 截断，file/http 缺失） | P1 | ✅ Done |
| T26 | Provider 请求/响应结构化日志：记录 method/url/status/duration/tokens，不记录完整 apiKey | P2 | ✅ Done |
| T27 | Token Cost 追踪：`ModelResponse` 可选 `CostEstimateUSD`，轻量价格表 + verbose CLI 输出 | P2 | ✅ Done |
| T28 | Provider 健康状态缓存：`axis provider status` 增强，提示用户运行 test 验证 | P2 | ✅ Done |

## Coverage Targets

| Component | Target | Actual |
|----------|--------|--------|
| Provider implementations | 90%+ | 91.8% ✅ |
| Tool implementations | 90%+ | 93.7% ✅ |
| Executor changes | 85%+ | 95.1% ✅ |

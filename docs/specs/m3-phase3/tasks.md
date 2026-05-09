# M3 Phase 3 Tasks

## Related Documents

- [requirements.md](requirements.md)
- [design.md](design.md)
- [workflow-binding.md](workflow-binding.md)

## Progress Tracking

| Task | Status | Depends On |
|---|---|---|
| T1: SLA types & constants | Pending | — |
| T2: Failure class routing + backoff | Pending | T1 |
| T3: Priority sorting in scheduler | Pending | T1 |
| T4: SLA admission extension | Pending | T1 |
| T5: Tool types & registry | Pending | — |
| T6: BashTool | Pending | T5 |
| T7: Extended ModelProvider (multi-turn) | Pending | T5 |
| T8: MockModelProvider tool-aware | Pending | T6, T7 |
| T9: Multi-turn execution loop | Pending | T5, T7 |
| T10: Orchestrator wiring | Pending | T2, T3, T4, T8, T9 |
| T11: Tests & coverage | Pending | T10 |
| T12: CLI/docs update | Pending | T11 |

---

## T1: SLA types & constants

**Goal**: Add failure class, priority, backoff constants to types package.

**Files**: `internal/types/types.go`

**Acceptance Criteria**:
- `FailureClassRetryable`, `FailureClassFatal`, `FailureClassDegradable` 常量存在
- `SLAKeyPriority`, `SLAKeyBackoff` metadata keys 存在
- `BackoffFixed`, `BackoffLinear`, `BackoffExponential` 常量存在
- 编译通过

---

## T2: Failure class routing + backoff

**Goal**: `parseSLA` 返回 failure class 和 backoff 策略；`executeTask` 按 failure class 分支处理。

**Files**: `internal/kernel/orchestrator/orchestrator.go`

**Acceptance Criteria**:
- `parseSLA` 返回 `(timeout, retries, failureClass, backoff string)`
- `fatal` 类型不重试，直接标记 failed
- `degradable` 类型在依赖未就绪时不阻塞（跳过依赖检查）
- `retryable` + 未设置保持当前重试行为
- backoff 策略在重试之间生效

---

## T3: Priority sorting in scheduler

**Goal**: `GetReadyTasks` 按 `sla.priority` 降序返回 ready tasks。

**Files**: `internal/kernel/scheduler/scheduler.go`

**Acceptance Criteria**:
- 解析每个 task 的 priority metadata（默认 128）
- `GetReadyTasks` 返回按优先级降序排列的任务
- 同优先级保持 FIFO
- 不影响 Submit/GetStatus/Cancel 行为

---

## T4: SLA admission extension

**Goal**: Admission 验证 priority 和 backoff 字段。

**Files**: `internal/contract/admission/admission.go`

**Acceptance Criteria**:
- `sla.priority`: 必须是 0-255 整数
- `sla.backoff`: 必须是 "fixed" | "linear" | "exponential"
- `sla.failure_class`: 必须是 "retryable" | "fatal" | "degradable"
- 无效值时 admission 拒绝

---

## T5: Tool types & registry

**Goal**: 定义 Tool 接口、ToolRegistry、ToolDefinition/ToolCall/ToolResult 类型。

**Files**:
- `internal/types/types.go` — ToolCall, ToolResult, ToolDefinition, ModelMessage
- `internal/model/tool/tool.go` — Tool interface + ToolRegistry

**Acceptance Criteria**:
- `Tool` 接口：Name(), Schema(), Execute(ctx, input) (output, error)
- `ToolRegistry`：Register, Get, List 方法
- 编译器通过

---

## T6: BashTool

**Goal**: 实现 Bash 工具，通过 `os/exec` 执行命令。

**Files**: `internal/model/tool/bash.go`

**Acceptance Criteria**:
- 读取 input["command"] 作为 bash 命令
- 30 秒超时（context.WithTimeout）
- 返回 stdout, stderr, exit_code
- 命令失败时不返回 error（exit_code 非零是正常结果），只对系统错误返回 error

---

## T7: Extended ModelProvider (multi-turn)

**Goal**: 扩展 ModelRequest/ModelResponse/ModelProvider 支持多轮和 tool calls。

**Files**: `internal/model/provider/provider.go`

**Acceptance Criteria**:
- `ModelRequest` 增加 `Tools []ToolDefinition` 和 `History []ModelMessage` 字段
- `ModelResponse` 增加 `ToolCalls []ToolCall` 字段
- `ModelProvider` 接口不变（Execute 签名不变）
- 向后兼容（现有调用方无需修改）

---

## T8: MockModelProvider tool-aware

**Goal**: Mock provider 能模拟 tool-use 多轮交互。

**Files**: `internal/model/provider/mock.go`

**Acceptance Criteria**:
- 当 input 包含 `"tool"` key 时：返回 tool_call（非 Output）
- 当 History 最后是 tool result 时：返回 final output
- 无 tool 场景行为不变

---

## T9: Multi-turn execution loop

**Goal**: ContractExecutor 支持 provider → tool → provider 循环。

**Files**: `internal/contract/executor/executor.go`

**Acceptance Criteria**:
- 如果 request 包含 tools，进入 multi-turn 模式
- Provider 返回 tool_calls → execute tools → feed back → repeat
- 最多 10 轮
- 如果 request 没有 tools，行为不变（单轮）
- Tool execution 错误记录在 ToolResult.Error 中

---

## T10: Orchestrator wiring

**Goal**: 在 NewOrchestrator 中组装 ToolRegistry + BashTool + SLA 策略。

**Files**: `internal/kernel/orchestrator/orchestrator.go`

**Acceptance Criteria**:
- ToolRegistry 创建并注册 BashTool
- ToolRegistry 注入 ContractExecutor
- SLA 策略通过 orchestrator 生效
- `go build -o axis-dev.exe cmd/axis/main.go` 成功

---

## T11: Tests & coverage

**Goal**: 所有新代码有测试覆盖，覆盖率 ≥ 85%。

**Files**:
- `internal/kernel/orchestrator/orchestrator_test.go` — SLA routing tests
- `internal/kernel/scheduler/scheduler_test.go` — priority ordering tests
- `internal/contract/admission/admission_test.go` — SLA validation tests
- `internal/model/tool/tool_test.go` — registry tests
- `internal/model/tool/bash_test.go` — bash execution tests
- `internal/model/provider/mock_test.go` — tool-aware mock tests
- `internal/contract/executor/executor_test.go` — multi-turn tests

**Acceptance Criteria**:
- `go test -race ./...` 通过
- 覆盖率 ≥ 85%

---

## T12: CLI/docs update

**Goal**: 更新文档反映 M3 Phase 3 变更。

**Files**:
- `docs/current-progress.md`
- `HANDOVER.md`
- `docs/QUICKSTART.md`

**Acceptance Criteria**:
- 进度文档标记 Phase 3 完成
- HANDOVER 记录新能力
- 覆盖率数据更新

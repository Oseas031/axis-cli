# M3 Phase 3 Requirements — SLA策略引擎 & 工具调用层

## Summary

M3 Phase 3 包含两个相互独立的子特性：

1. **SLA 策略引擎**：让 `failure_class` 实际影响调度和重试行为，增加优先级排序和退避策略
2. **工具调用层**：让任务执行链具备多轮工具调用能力，首个工具为 Bash 执行

两者都可以独立开发、测试、合并。

## Design Philosophy

- **More Context**: 策略引擎根据失败类型做差异化处理，工具调用把执行结果带回模型上下文
- **More Action**: 工具层让任务真正可以执行 Bash 命令，而不是只生成文本
- **Zero Control**: 策略可配置，工具可注册，不硬编码单一路径
- **Bash is All You Need**: 首个工具即 Bash，CLI 原生

## Users

- 需要差异化失败处理的调度场景
- 需要通过任务执行 Bash 命令的 Agent
- 后续需要多轮 tool-use 的 Provider 实现

## Functional Requirements

### SLA 策略引擎

- **FR1**: `failure_class` 值决定失败行为：
  - `"retryable"` — 退避重试至最大次数
  - `"fatal"` — 失败立即终止，不重试
  - `"degradable"` — 依赖未就绪时降级运行（跳过缺失依赖）
  - 未设置时保持当前默认行为（全部重试）
- **FR2**: 退避策略可配置：固定间隔、线性增长、指数退避，默认固定 100ms
- **FR3**: 优先级字段 `sla.priority`（0-255），高优先级任务在 `GetReadyTasks` 中优先返回
- **FR4**: 调度器按优先级排序 ready tasks，同优先级保持 FIFO

### 工具调用层

- **FR5**: `Tool` 接口定义：`Name()`, `Schema()`, `Execute(ctx, input) → output`
- **FR6**: `ToolRegistry`：注册、查找、列出工具
- **FR7**: `BashTool`：执行 shell 命令，返回 stdout/stderr/exit_code，30 秒超时
- **FR8**: `ModelRequest` 扩展，支持 `Tools []ToolDefinition` 字段
- **FR9**: `ModelResponse` 扩展，支持 `ToolCalls []ToolCall` 字段（tool_use 时非空）
- **FR10**: `ContractExecutor` 支持多轮执行循环：provider → tool_use? → execute tool → feed result → provider → ... → final output

### 双方共享

- **FR11**: 所有新行为由 `go test -race ./...` 覆盖
- **FR12**: 覆盖率不低于 85%

## Non-Goals

- 真实 LLM 集成（仍只 Mock/Echo）
- 网络调用工具（http client 等）
- 文件读写工具（Phase 4，先用 Bash 覆盖）
- SLA compliance tracking / metrics
- 动态优先级调整
- 工具调用的流式返回

## Acceptance Criteria

- [ ] `failure_class` 三种行为正确执行
- [ ] 优先级排序在 `GetReadyTasks` 中生效
- [ ] `BashTool` 可执行命令并返回结果
- [ ] 多轮 tool-use 循环在 `ContractExecutor` 中工作
- [ ] `go test -race ./...` 通过
- [ ] 覆盖率 ≥ 85%

## Constraints

- Go stdlib only（Bash 使用 `os/exec`）
- 不改 scheduler 核心语义（FIFO 仍是默认）
- 不改现有 API 签名（只能扩展）
- 不引入外部 DSL 或规则引擎

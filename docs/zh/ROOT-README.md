# Axis

> Agent 原生调度系统。不是为了控制 Agent，而是为了让 Agent 在任务实践中积累胜任力、赢得自主权，并最终生成自身。

Axis 的目标不是做一个更强的任务队列，也不是做一个包裹 LLM 的工具框架。Axis 的目标是构造一个 **Agent 自因化的执行底座**：让 Agent 能理解任务、组织行动、验证结果、反思失败、生成下一轮任务，并在可靠表现中逐步获得更大的自主权。

## 核心命题

```text
More Context, More Action, Zero Control, Controllable Evolution
```

- **More Context**：系统提供查询基础设施，Agent 主动查询和构建上下文，而非被动接收冗余推送
- **More Action**：执行、组合、验证、修正和生成后续任务的行动能力，权限与能力匹配
- **Zero Control**：系统提供契约、基础设施和可观测性，但不替 Agent 规定唯一行动路径
- **Controllable Evolution**：自举、自生成和自我修改必须处于可观测、可验证、可回滚的边界内

## 设计原则

```text
bash is all you need · Competence earns autonomy · Interface is existence
```

- **CLI 优先**：可脚本化、可组合、可被人类/CI/Agent 调用，不默认引入 Web UI 或复杂 TUI
- **递进自主权**：越可靠，行动半径越大；高风险操作不因执行者身份豁免
- **接口即存在**：人和 Agent 实现相同 agent 接口，无身份偏见
- **契约即结构**：文件系统/元文件是所有 Agent 的公共契约语言
- **过渡性结构**：workflow/contract/permission/spec 是种子和脚手架，终将被 Agent 内化、重写和扬弃

## 当前状态

M1 ✅ | M2 ✅ | M3 ✅ | M4 ✅ | M5 ✅ | M6 ✅ | Sandboxed Evolution ✅ | Local Control Plane ✅

### 已完成能力

- **任务调度**：FIFO + DAG 并行调度、依赖管理、5-worker 并行 orchestrator、contract admission、SLA timeout/retry/failure_class 策略引擎
- **LLM 集成**：Anthropic / OpenAI / DeepSeek / MiniMax provider、token accounting、circuit breaker、项目本地 provider profile 管理
- **工具系统**：BashTool（可观测执行记录）、FileReadTool、FileWriteTool、HTTPClientTool、tool permission scopes、多轮执行循环
- **自然语言调度**：`axis ask` 将 prompt 编译为 AgentTask，dry-run 预览 / 显式 submit，不绕过契约
- **自适应上下文装配**：ContextBundle / ReadinessArtifact / ReadinessRegistry / preflight / strict gate，规则装配 + 预算裁剪，preview-first 不侵入执行
- **执行时上下文消费**：ExecutionContextSummary / ExecutionContextConsumer，Agent 声明 `context.requested_sources`，dispatcher 注入摘要
- **本地控制面**：`axis start` 启动 loopback HTTP control server，跨进程提交/查询，`.axis/runtime.json` 定位器，append-only 事件日志
- **沙盒演化协议**：隔离工作空间 + 原子步骤 + 追踪账本 + 验证捕获 + 显式 promote/discard 门控，审计全链路
- **自判定引擎**：SelfJudgementEngine + 5 种验证策略（Syntax/Semantic/Contract/Test/Coverage）、自判定契约、BootstrapOrchestrator 判定集成
- **自举循环**：BootstrapOrchestrator + FollowUpTaskGenerator + AutonomyTransition 规则引擎 + self-iteration contracts
- **9+ 结构化错误码**、Agent Context Query Model、DAG 可见性

## 快速开始

```bash
go test ./...
go build -o axis-dev.exe ./cmd/axis
```

> Windows 本地开发建议输出到 `axis-dev.exe`，避免覆盖或锁定根目录下既有 `axis.exe`。

### Local Runtime

跨命令提交与查询需要显式启动本地 runtime：

```powershell
# Terminal A: start the project-local runtime
.\axis-dev.exe start

# Terminal B: submit a natural-language task
.\axis-dev.exe ask "check provider config" --submit --task-id provider-check

# Terminal B: query task status
.\axis-dev.exe status provider-check
```

- `axis start` 写入 `.axis/runtime.json`，暴露 loopback control server，事件追加到 `.axis/events/tasks.jsonl`
- `axis ask <prompt>` 默认 dry-run 预览，不需要 runtime
- `axis shell` 是 in-process session，shell 内 `run`/`ask --submit`/`status` 共享会话状态，不会静默 attach 到 `axis start`

### Provider 管理

```powershell
# 添加项目本地 provider profile
.\axis-dev.exe provider add claude --type anthropic --api-key sk-ant-... --model claude-3-5-sonnet-20241022
.\axis-dev.exe provider add gpt --type openai --api-key sk-... --model gpt-4o-mini
.\axis-dev.exe provider add ds --type deepseek --api-key sk-... --model deepseek-chat
.\axis-dev.exe provider add mm --type minimax --api-key ... --model MiniMax-Text-01

# 切换 / 查看 / 列表
.\axis-dev.exe provider use claude
.\axis-dev.exe provider status
.\axis-dev.exe provider list
```

Profile 存储在 `.axis/providers.json`，不修改 shell 环境变量或系统配置。

### 上下文预览与就绪检查

```powershell
# 预览任务的上下文装配结果
.\axis-dev.exe context preview "check provider config"

# 检查上下文就绪状态
.\axis-dev.exe context inspect <bundle-id>
.\axis-dev.exe context preflight <task-id>
.\axis-dev.exe context preflight <task-id> --strict
```

### 沙盒演化

```powershell
# 检查演化运行详情
.\axis-dev.exe evolve inspect <run-id>

# 晋升或丢弃演化结果
.\axis-dev.exe evolve promote <run-id>
.\axis-dev.exe evolve discard <run-id>
```

### 自判定

```powershell
# 运行自判定诊断
.\axis-dev.exe judge
```

## CLI 命令一览

| 命令 | 用途 |
|------|------|
| `axis start` | 启动本地 runtime（loopback control server） |
| `axis run <task-id>` | 提交并运行任务 |
| `axis status <task-id>` | 查询任务状态（通过 local runtime） |
| `axis ask <prompt>` | 自然语言 → AgentTask（默认 dry-run） |
| `axis ask <prompt> --submit` | 提交自然语言任务到 local runtime |
| `axis shell` | 启动交互式 in-process shell |
| `axis provider add/use/status/list/remove/archive` | 管理项目本地 LLM provider profile |
| `axis context preview/inspect/preflight` | 上下文装配预览与就绪检查 |
| `axis judge` | 运行自判定诊断 |
| `axis evolve inspect/promote/discard` | 沙盒演化检查与决策 |

## 外部工具

- **[axis-gui](tools/axis-gui/)**：本地 Web GUI，连接 Local Control Plane，提供 Dashboard / Tasks / Providers / Events 界面（WebSocket 实时推送）
- **[axis-up](tools/axis-up/)**：引导式上手工具，环境检测 / 构建 / 配置 / Demo 一条龙

两者均不 import Axis internal 包，通过 CLI 和 HTTP API 通信。

## 重要文档

- [Agent 原生第一性原理](architecture/agent-native-first-principles.md) **← 编码前必读**
- [Bash is All You Need](architecture/bash-is-all-you-need.md)
- [系统规范总纲](architecture/axis-system-conventions.md)
- [当前进度](status/current-progress.md)
- [文档总入口](README.md)
- [原生场景白皮书](product/axis-native-scenarios-whitepaper.md)
- [Autogenesis 设计报告](../reports/strategy/axis-autogenesis-design-2026-05-08.md)

## 技术栈

- **Go 1.21+**，核心模块优先使用标准库
- **单二进制 CLI**，shell-native workflow
- **Cobra** CLI 框架
- **项目本地状态**：`.axis/` 目录（providers.json / runtime.json / events/ / evolution/）

## 项目结构

```text
cmd/axis/          CLI 入口与命令定义
internal/
  types/           核心数据类型（AgentTask, AgentContract, ErrorCode...）
  kernel/          调度器、编排器、分发器
  contract/        契约执行器
  model/           LLM provider + tool 系统
  agent/           Agent 执行器 + 自判定引擎
  intent/          自然语言意图解析
  contextpack/     自适应上下文装配
  control/         本地控制面（server/client/locator/events）
  evolution/       沙盒演化协议
  human/           人类执行器
docs/              文档入口、架构参考、规格文档、状态报告
tools/
  axis-gui/        本地 Web GUI
  axis-up/         引导式上手工具
```

## 下一步方向

- **跨进程状态持久化**：ReadinessRegistry 对接 Local Control Plane
- **Agent 身份与能力档案**：Agent 注册中心 + 行为评分
- **事件日志结构化查询**：`axis audit` 或等价能力
- **动态模型路由**：cost/latency-aware provider 选择 + 降级链
- **执行反馈闭环**：结果质量评分反馈到意图/上下文装配


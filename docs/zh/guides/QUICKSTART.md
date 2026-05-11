# 快速入门

## Axis 是什么

Axis 是 Agent 原生调度系统，也是 Agent 自因化的执行底座。

它不是普通任务队列，也不是 LLM wrapper。Axis 让 Agent 能把工作表达为任务，获得上下文，执行行动，验证结果，反思失败，生成后续任务，并在可靠表现中赢得更大的自主权。

## 核心原则

```text
More Context, More Action, Zero Control, Controllable Evolution
bash is all you need · Competence earns autonomy · Interface is existence
```

## 当前能力（M1-M6 ✅）

- **任务调度**：FIFO + DAG 并行调度、依赖管理、contract admission、SLA 策略引擎
- **LLM 集成**：Anthropic / OpenAI / DeepSeek / MiniMax provider、token accounting、circuit breaker
- **工具系统**：BashTool、FileReadTool、FileWriteTool、HTTPClientTool
- **自然语言调度**：`axis ask` 将 prompt 编译为 AgentTask，dry-run 预览 / 显式 submit
- **自适应上下文装配**：ContextBundle / ReadinessArtifact / preflight / strict gate
- **本地控制面**：`axis start` 启动 loopback control server，跨进程提交/查询
- **沙盒演化协议**：隔离工作空间 + 原子步骤 + 显式 promote/discard 门控
- **自判定引擎**：5 种验证策略 + BootstrapOrchestrator 判定集成
- **自举循环**：BootstrapOrchestrator + FollowUpTaskGenerator + AutonomyTransition

## 构建与测试

```bash
go test ./...
go build -o axis-dev.exe ./cmd/axis
```

> Windows 下建议构建为 `axis-dev.exe`，避免覆盖根目录既有 `axis.exe`。

## 基本用法

### 本地 Runtime（跨进程模式）

```powershell
# Terminal A: 启动本地 runtime
.\axis-dev.exe start

# Terminal B: 提交自然语言任务
.\axis-dev.exe ask "check provider config" --submit --task-id provider-check

# Terminal B: 查询任务状态
.\axis-dev.exe status provider-check
```

### 单进程模式

```powershell
.\axis-dev.exe run my-task
.\axis-dev.exe shell
```

### Provider 配置

```powershell
.\axis-dev.exe provider add claude --type anthropic --api-key sk-ant-... --model claude-3-5-sonnet-20241022
.\axis-dev.exe provider use claude
.\axis-dev.exe provider status
```

### 更多命令

```powershell
.\axis-dev.exe ask "prompt"              # dry-run 预览
.\axis-dev.exe context preview "prompt"  # 上下文预览
.\axis-dev.exe context preflight <id>    # 就绪检查
.\axis-dev.exe judge                     # 自判定诊断
.\axis-dev.exe evolve inspect <run-id>   # 演化检查
```

## 新手推荐

如果你是第一次接触 Axis，推荐使用 [axis-up](../tools/axis-up.md) 引导工具：

```powershell
cd tools\axis-up
go build -o axis-up.exe .
.\axis-up.exe start
```

## 下一步

- 阅读 [Agent 原生第一性原理](../architecture/agent-native-first-principles.md) — **编码前必读**
- 查看 [当前进度](../status/current-progress.md)
- 查看 [项目 Roadmap](../product/ROADMAP.md)

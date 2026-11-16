# Axis

[English](README.md) | **简体中文**

> **单二进制跨 Harness AI Agent 中控台与本地编码编排器**（Go，零框架）。  
> Axis 将本地原生 LLM Agent 与外部编码 CLI 执行器（`claude`、`codex`、`gemini`、`opencode` 等）统一纳入同一本地控制面：**单命令启动三区 Web 工作台（`axis gui`）、`git worktree` 隔离防冲突、时间线内联人工审批，以及跨 Harness 自动编排流水线**。

## 核心业务场景

当开发者从“单轮聊天写代码”转向**多模型、多执行器并发开发**时，通常会遇到三大工程瓶颈：

1. **多执行器割裂与并发改代码冲突**：不同编码 Agent 各自为战，同时修改同一仓库会互相覆盖 → 通过 **多 Harness 注册中心 + Per-Task `git worktree` 隔离 + 工作区 Diff 审查台** 解决。
2. **多模型成本失控与长任务黑盒**：前沿模型在简单步骤浪费预算，长任务运行中难以叫停或接管 → 通过 **DAG 调度、单任务 USD 预算封顶、熔断与模型升级、`POST /v1/tasks/{id}/cancel` 一键叫停、SSE 实时流** 解决。
3. **人机交接断裂与不安全变更**：后台遇到歧义时要么盲目乱改 20 个文件要么崩溃 → 通过 **对称人机调度（`human/gate` + `axis inbox` + 时间线内联审批与掩码安全输入）** 与 **分阶段沙箱演进（Staged Evolution）** 解决。

## 核心能力

- **单二进制多 Harness 中控台（`axis gui` / `axis web`）**：通过 `go:embed` 内嵌零打包器三区工作台（中英双语自适应系统语言），支持 SSE 实时推送、`@harness` 提及路由、`Stop` / `Stop now` 即时叫停、时间线内联人类审批（含机密字段自动掩码）、`git worktree` 隔离与 Diff 审查、一键存为 Skill、Routine 定时例程、Memory 记忆面板及跨 Harness 流水线（`{{prev_output}}` / `{{prev_diff}}`）。
- **任务调度与预算治理**：FIFO + DAG 并行调度、5-Worker 编排器、契约准入校验、SLA 超时/重试/失败分级策略、单任务成本预算控制。
- **多模型集成**：支持 Anthropic / OpenAI / DeepSeek / MiniMax 及自定义兼容接口，内置熔断器、项目级 `.axis/providers.json` 配置管理与质量门控升级。
- **对称人机协同**：`internal/human/gate` + JSONL CAS 持久化收件箱（`.axis/comm/`），支持 `axis inbox` CLI 与 Web 时间线双端裁决。
- **按需上下文与分阶段演进**：`contextpack` 相关性评分与 `preflight --strict` 就绪门禁；隔离工作区演进、原子步骤账本与显式 `promote` / `discard`。
- **能力沉淀与自动化**：`.axis/skills/<name>/SKILL.md` 技能加载器、工作记忆系统、System Guarantees 承诺校验与工作项追踪（`axis vigil`）。

## 当前已知边界

> 保持架构透明：以下为当前设计边界，而非偶然缺陷。

- **单机本地优先**：控制面为本机回环 HTTP（`127.0.0.1`），不做分布式调度或多节点选主。
- **单用户单项目**：面向个人/本地工程目录，不含多租户企业账号体系。
- **容器沙箱依赖 Docker**：`SandboxedBashTool` 在无 Docker 时会告警并降级执行。
- **自评结果仅供参考**：编译器与确定性测试（`go test` / `go vet`）才是最终权威验收标准。

## 快速开始

### 1. 一键启动单二进制工作台（前后端一体）

```powershell
# Windows 一键编译并打开浏览器工作台
.\start.bat

# 或直接使用 Go 运行（等价别名：axis web）
go run ./cmd/axis gui
```

### 2. 编译与测试验证

```powershell
# 无需真实 API Key（使用内置 mock provider）
go build -o axis-dev.exe ./cmd/axis
go test -race -short ./internal/control ./internal/kernel/... ./cmd/axis
.\axis-dev.exe ask "check provider config"       # 默认 dry-run 预览任务结构
.\axis-dev.exe context preview "check config"    # 预览按需上下文装配
```

> Windows 下建议编译输出为 `axis-dev.exe`，避免与正在运行的 `axis.exe` 产生文件锁冲突。

### 3. 配置项目级模型 Provider

```powershell
# 添加本地 Provider 配置（保存于 .axis/providers.json，不改动系统环境变量）
.\axis-dev.exe provider add ark --type openai --base-url https://... --api-key ... --model ...
.\axis-dev.exe provider use ark
.\axis-dev.exe provider status
```

## 常用 CLI 命令

| 命令 | 用途 | 是否需要本地 Runtime |
|---|---|---|
| `axis gui` / `axis web` | 一体化启动本地运行时与单二进制三区 Web 工作台 | 否（自动启动） |
| `axis start` | 启动本地回环控制面运行时（`.axis/runtime.json`） | 否（创建运行时） |
| `axis ask <prompt> [--submit]` | 将自然语言编译为 `AgentTask`（默认仅预览，加 `--submit` 提交） | `--submit` 时需要 |
| `axis run <task-id> [--background]` | 同步执行或后台提交任务 | `--background` 时需要 |
| `axis status <task-id>` | 查询任务状态与结果 | 是 |
| `axis inbox list/show/respond` | 查看并回复 Agent 发起的人工审批/询问请求 | 可选 |
| `axis provider add/use/status/list` | 管理项目级模型 Provider 配置 | 否 |
| `axis skills list/show/validate/create` | 管理 `.axis/skills/` 技能 | 否 |
| `axis memory ...` | 管理分层工作记忆 | 否 |
| `axis evolve inspect/promote/discard` | 审查并合并/丢弃隔离演进结果 | 否 |
| `axis vigil resume/list/add/done` | 跨会话工作项追踪 | 否 |

## 核心文档

- [CLAUDE.md](CLAUDE.md) — 工程宪法、工作协议与模块边界
- [Current Progress](docs/status/current-progress.md) — 里程碑与当前进展基准
- [Semantic Boundaries](docs/architecture/semantic-boundaries.md) — 模块职责禁止清单
- [Documentation Index](docs/README.md) — 完整文档索引

## 目录结构

```text
cmd/axis/          CLI 入口与各子命令定义
internal/
  types/           核心数据类型（AgentTask, AgentContract, ErrorCode...）
  kernel/          DAG 调度器、Orchestrator、Dispatcher、FeatureGate
  control/         本地控制面、多 Harness 调度、Worktree 隔离、单二进制 Web UI
  contract/        契约执行器（权限域、熔断器、上下文压缩）
  model/           LLM Provider 适配与工具系统
  agent/           多轮 Agent 执行器与自评引擎
  contextpack/     自适应上下文装配
  evolution/       分阶段隔离演进协议
  comm/            JSONL CAS 多 Agent/人机信箱
  skills/          按需知识技能加载器
  memory/          分层记忆子系统
  human/           对称人机执行门控（human/gate）
  vigil/           跨会话工作项追踪
docs/              架构规范、状态报告与设计文档
```

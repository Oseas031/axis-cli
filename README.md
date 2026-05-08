# Axis

> Agent 原生调度系统。不是为了控制 Agent，而是为了让 Agent 在任务实践中积累胜任力、赢得自主权，并最终生成自身。

Axis 的目标不是做一个更强的任务队列，也不是做一个包裹 LLM 的工具框架。Axis 的目标是构造一个 **Agent 自因化的执行底座**：让 Agent 能理解任务、组织行动、验证结果、反思失败、生成下一轮任务，并在可靠表现中逐步获得更大的自主权。

## 核心命题

```text
More Context, More Action, Zero Control
```

Axis 相信 Agent 需要的不是更密集的外部控制，而是：

- **More Context**：获得足够理解任务、依赖、历史、系统状态和失败原因的上下文
- **More Action**：拥有执行、组合、验证、修正和生成后续任务的行动能力
- **Zero Control**：系统提供契约、基础设施和可观测性，但不替 Agent 规定唯一行动路径

## 交互原则

```text
bash is all you need
```

Axis 默认保持 shell-native：

- CLI 优先
- 可脚本化
- 可组合
- 可被人类、CI 和 Agent 调用
- 不默认引入 Web UI 或复杂 TUI

这不是简陋，而是为了让 Axis 自身也能被 Agent 直接调用、编排和改造。

## 权限哲学

```text
Competence earns autonomy
```

Axis 的权限边界不是静态文件白名单，而是递进自主权机制：

- **Competence**：Agent 在真实任务中展现出的可靠执行能力
- **Earns**：自主权通过稳定、可验证、可审计的表现逐步赢得
- **Autonomy**：不只是文件访问权，而是自主决策、独立执行任务的能力

越可靠，行动半径越大；越不稳定，系统要求更多上下文、验证或人类确认。

## 自举起点

Axis 的自举不是从第一行自我修改代码开始，而是从外部 Agent 向 Axis 注入可被吸收、固化、执行、反思和演化的思想开始。

当前阶段已经是自举起点：

```text
external thought injection
  -> architecture / spec / workflow / contract / permission
  -> implementation path
  -> execution
  -> reflection
  -> self-revision
```

工程上的 bootstrap loop 只是第一层；更深层目标是 **Autogenesis Loop**。

## Autogenesis Loop

```text
Perceive self
  -> Diagnose self
  -> Redefine self
  -> Modify self
  -> Validate self
  -> Judge self
  -> Re-authorize self
  -> Repeat
```

Axis 最终要支持的不是“Agent 调工具”，而是 Agent 把自身作为对象来理解、修改、验证、评判和重新授权。

## 过渡性结构

Axis 当前仍需要早期工程结构，但它们不是终点：

- **workflow 是临时脚手架**：帮助尚未成熟的 Agent 组织行动
- **contract 是成长边界**：帮助 Agent 表达任务、验证结果，并最终自我立约
- **permission rule 是递进自主权机制**：也是 Axis 涅槃前的枷锁，终将被内化、重写和扬弃
- **spec 是种子**：不是终局蓝图，而是下一阶段演化的发生源

这些结构的使命不是永久控制 Agent，而是帮助 Agent 成长到可以重写它们。

## 当前状态

Milestone 1 已完成并验收。Axis 已具备：

- 基础任务模型
- FIFO 调度
- 依赖管理
- 契约执行器
- 状态存储
- 编排器
- CLI / shell 入口
- 工作流与 specs 机制

Milestone 2 正在建设 Autogenesis Loop 的执行底座：

- ready-set DAG 调度
- contract admission
- SLA timeout / retry
- parallel orchestrator
- stable error codes
- CLI / docs acceptance

## 快速开始

```bash
go test ./...
go build -o axis-dev.exe cmd/axis/main.go
.\axis-dev.exe run my-task
```

Windows 本地开发建议输出到 `axis-dev.exe`，避免覆盖或锁定根目录下既有 `axis.exe`。

## 重要文档

- [Agent 原生设计思想](docs/architecture/agent-native-design-philosophy.md)
- [Bash is All You Need](docs/architecture/bash-is-all-you-need.md)
- [Autogenesis 设计报告](reports/axis-autogenesis-design-2026-05-08.md)
- [自因化自举差距分析](reports/bootstrap-gap-analysis-2026-05-08.md)
- [当前进度](docs/current-progress.md)
- [Milestone 2 Specs](docs/specs/milestone2/)
- [工作流入口](workflow/entry.md)

## 技术栈

- Go 1.26+
- 核心模块优先使用 Go 标准库
- 单二进制 CLI
- Shell-native workflow

## 当前最重要的下一步

继续 Milestone 2：

```text
T3 contract admission layer
```

因为自举不能先从真实 LLM 开始，而要先让 Axis 能安全、可验证、可恢复地调度“改进 Axis 自己”的任务 DAG。

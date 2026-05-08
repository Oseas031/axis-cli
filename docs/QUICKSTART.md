# 快速入门

## Axis 是什么

Axis 是 Agent 原生调度系统，也是 Agent 自因化的早期执行底座。

它不是普通任务队列，也不是 LLM wrapper。Axis 要让 Agent 能把工作表达为任务，获得上下文，执行行动，验证结果，反思失败，生成后续任务，并在可靠表现中赢得更大的自主权。

## 三条核心原则

```text
More Context, More Action, Zero Control
bash is all you need
Competence earns autonomy
```

## 当前能做什么

当前 Axis 已完成 Milestone 1，并进入 Milestone 2：

- 基础任务调度
- 依赖管理
- 契约执行
- 状态查询
- CLI / shell 入口
- ready-set DAG 调度 API

## 本地验证

```bash
go test ./...
go build -o axis-dev.exe cmd/axis/main.go
```

Windows 下建议构建为 `axis-dev.exe`，避免覆盖根目录既有 `axis.exe`。

## 运行一个任务

```bash
.\axis-dev.exe run my-task
.\axis-dev.exe status my-task
```

当前普通 CLI 状态是本地进程语义；跨进程持久状态属于后续自举能力的一部分。

## 当前不要做什么

- 不要直接接真实 LLM SDK
- 不要引入 Web UI / 重型 TUI
- 不要引入外部数据库
- 不要跳过 Milestone 2 的 contract admission / SLA / error code
- 不要把 workflow、contract、permission、spec 当成终局控制结构

## 下一步

继续 Milestone 2：

```text
T3 contract admission layer
```

M2 的意义不是普通并行调度，而是为未来 Autogenesis Loop 提供执行底座。

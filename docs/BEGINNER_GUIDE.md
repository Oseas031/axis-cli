# Axis 小白快速上手指南

这份指南适合第一次接触 Axis 的用户。你不需要先理解所有架构，只要跟着步骤走，就可以把项目跑起来，并用交互式 Shell 提交和查看任务。

---

## 1. Axis 是什么？

Axis 是一个 **Agent 原生调度系统**。

你可以先把它理解成：

> 一个帮 AI Agent 接收任务、排队任务、执行任务、查看任务状态的小型调度器。

它目前不是网页应用，也不是聊天机器人，而是一个命令行工具。

---

## 2. Axis 的设计思想

Axis 有两个核心思想。

### 2.1 More Context, More Action, Zero Control

意思是：

- **More Context**：给 Agent 更多上下文，让它知道发生了什么
- **More Action**：给 Agent 更多行动能力，让它能做更多事
- **Zero Control**：系统尽量不强行控制 Agent，只提供基础设施和边界

### 2.2 Bash is All You Need

意思是：

> 优先用命令行解决问题。

Axis 默认不先做复杂网页 UI，而是先做好 CLI 和 Shell：

```bash
axis shell
axis run task-1
axis status task-1
```

这样更容易被人、脚本、CI、AI Agent 调用。

---

## 3. 你需要准备什么？

### 3.1 安装 Go

Axis 是 Go 项目，所以你需要先安装 Go。

检查是否安装成功：

```bash
go version
```

如果能看到类似下面的输出，说明 Go 已经可用：

```text
go version go1.xx.x windows/amd64
```

### 3.2 进入项目目录

如果你已经在 IDE 里打开了本项目，项目目录是：

```text
c:\Users\ASUS\Desktop\axis-cli
```

---

## 4. 第一次构建项目

在项目根目录运行：

```bash
go build -o axis.exe cmd/axis/main.go
```

成功后，项目目录下会出现：

```text
axis.exe
```

这就是 Axis 的命令行程序。

---

## 5. 推荐使用方式：交互式 Shell

最推荐小白先用：

```bash
.\axis.exe shell
```

启动后你会看到：

```text
Axis shell started. Type 'help' for commands, 'exit' to quit.
axis>
```

这说明你已经进入 Axis 的交互模式。

---

## 6. Shell 里能做什么？

当前 `run <task-id>` 使用的是内置的 `default` 合约。它不是在调用真实大模型，而是跑通一个最小任务链路：

```text
提交任务 -> 输入校验 -> 调度执行 -> 返回占位结果 -> 更新状态
```

这一步的目标是让你先看到 Axis 的调度闭环，后续再接真实模型 Provider。

### 6.1 查看帮助

输入：

```text
help
```

你会看到支持的命令：

```text
help              Show this help message
run <task-id>     Submit a task
status <task-id>  Show task status
exit, quit        Shut down the shell
```

### 6.2 提交一个任务

输入：

```text
run demo-task
```

可能看到：

```text
Task demo-task submitted. Try: status demo-task
```

这表示任务已经提交。

### 6.3 查看任务状态

输入：

```text
status demo-task
```

可能看到：

```text
Task demo-task status: pending
```

或：

```text
Task demo-task status: running
```

这表示 Axis 已经知道这个任务，并能查看它的状态。

### 6.4 输入错误命令也没关系

比如输入：

```text
abc
```

会看到：

```text
Unknown command: abc
Type 'help' to see available commands.
```

Shell 不会崩溃，你可以继续输入命令。

### 6.5 退出 Shell

输入：

```text
exit
```

或：

```text
quit
```

即可退出。

---

## 7. 一次完整体验

你可以照着输入：

```text
help
run demo-task
status demo-task
unknown
exit
```

如果你想在 PowerShell 里一次性测试，可以运行：

```powershell
@('help','run demo-task','status demo-task','unknown','exit') | .\axis.exe shell
```

---

## 8. 普通 CLI 命令怎么用？

除了 `axis shell`，Axis 也有普通命令。

### 8.1 启动调度器

```bash
.\axis.exe start
```

### 8.2 提交任务

```bash
.\axis.exe run demo-task
```

### 8.3 查询任务状态

```bash
.\axis.exe status demo-task
```

不过对小白来说，建议先用 `axis shell`，因为它不用反复重新输入完整命令。

---

## 9. 常见问题

### Q1：为什么没有网页 UI？

因为 Axis 当前阶段遵循：

> bash is all you need

也就是优先把命令行做好。网页 UI 不是不能做，而是暂时不是最小必要功能。

### Q2：`status` 提示找不到任务怎么办？

先确认你是否已经提交过这个任务：

```text
run demo-task
```

然后再查：

```text
status demo-task
```

### Q3：任务为什么一直是 pending 或 running？

当前项目还是里程碑 1 阶段，主要验证任务调度、状态管理、契约执行等核心概念。现在的 `default` 合约是最小占位执行链，还没有接入真实模型。

### Q4：我应该先看哪些文档？

建议顺序：

1. `docs/BEGINNER_GUIDE.md`：小白上手
2. `README.md`：项目总览
3. `docs/architecture/bash-is-all-you-need.md`：交互思想
4. `docs/architecture/agent-native-design-philosophy.md`：设计哲学
5. `HANDOVER.md`：当前项目状态

---

## 10. 最小使用路径

如果你只想最快跑起来，记住这三步：

```bash
go build -o axis.exe cmd/axis/main.go
.\axis.exe shell
```

进入 Shell 后：

```text
run demo-task
status demo-task
exit
```

这就是 Axis 的最小可用体验。

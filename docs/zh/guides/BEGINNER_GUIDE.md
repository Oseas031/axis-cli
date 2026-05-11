# Axis 小白快速上手指南：用 axis-up 启动 Axis

这份指南专门给第一次接触 Axis 的用户。你不需要先理解 Axis 的架构，也不需要先配置真实模型 API key。你只需要使用 `axis-up`，就可以完成环境检查、构建 Axis、运行 demo，并理解下一步该做什么。

**重要声明**：`axis-up` 是外部上手辅助工具，第一次上手推荐用它，但长期使用建议直接掌握 Axis CLI。`axis-up` 不会修改 Axis 本体源码，不会导入 Axis 内部包，只会通过公开命令调用 Axis。

---

## 0. Axis vs axis-up：从第一性原理出发的对比

理解 Axis 和 axis-up 的区别，要从它们的设计目标和交互对象出发：

| 维度 | Axis | axis-up |
|------|------|---------|
| **第一性设计目标** | Agent 自因化执行底座 | 人类快速上手辅助 |
| **交互对象** | AI Agent | 人类用户 |
| **核心职责** | 任务调度、契约执行、状态管理 | 环境检查、构建引导、零配置 demo |
| **长期定位** | 系统本体 | 临时桥梁 |
| **交互方式** | CLI / shell / 公开 API | CLI / 命令行工具 |
| **依赖关系** | 独立运行 | 依赖 Axis 公开 CLI |

**第一性结论**：

- **Axis 的存在是为了让 Agent 能把工作表达为任务，获得上下文，执行行动，验证结果，并在可靠表现中赢得更大的自主权。**
- **axis-up 的存在是为了让人类第一次接触 Axis 时，不被环境检查、构建、provider 配置这些技术细节卡住。**

因此，axis-up 是帮助人类“上车”的梯子，Axis 是车本身。上车后，你应该学会直接驾驶 Axis，而不是永远依赖梯子。

---

## 1. 你会用 axis-up 做什么？

第一次上手只需要记住四个命令：

```powershell
.\axis-up.exe start
.\axis-up.exe check
.\axis-up.exe demo
.\axis-up.exe fix
```

它们分别对应四种用户意图：

- **`start`**：我想第一次把 Axis 跑起来。
- **`check`**：我想知道当前环境是否准备好了。
- **`demo`**：我想看一个最小可运行演示。
- **`fix`**：我想修复常见上手问题。

如果你不知道该运行什么，优先运行：

```powershell
.\axis-up.exe start
```

---

## 2. 准备：确认 Go 可用

Axis 和 `axis-up` 都是 Go 程序，所以你需要先安装 Go。

在 PowerShell 中运行：

```powershell
go version
```

如果看到类似输出，说明 Go 已经可用：

```text
go version go1.xx.x windows/amd64
```

如果没有看到版本号，请先安装 Go，再继续。

---

## 3. 构建 axis-up

进入 `axis-up` 工具目录：

```powershell
cd c:\Users\ASUS\Desktop\axis-cli\tools\axis-up
```

构建 `axis-up.exe`：

```powershell
go build -o axis-up.exe .
```

构建成功后，当前目录会出现：

```text
axis-up.exe
```

这个文件就是给新手使用的外部辅助入口。

---

## 4. 第一次启动：axis-up start

运行：

```powershell
.\axis-up.exe start
```

`start` 会自动判断你当前需要什么：

```text
检测 Axis repo → 检测 Go → 必要时构建 axis-dev.exe → 使用 mock provider → 运行 demo
```

你会看到类似输出：

```text
Start Axis
----------
Goal:
  Get a first Axis run working without changing Axis core behavior.

[ok] Axis repo: C:\Users\ASUS\Desktop\axis-cli
[ok] Go: go version go1.xx.x windows/amd64
[ok] Axis binary: C:\Users\ASUS\Desktop\axis-cli\axis-dev.exe
```

如果缺少 `axis-dev.exe`，`axis-up start` 会解释为什么需要构建它，并执行类似命令：

```text
go build -o axis-dev.exe ./cmd/axis
```

> 注意：Axis 当前已完成 M1-M6 全部里程碑，包含真实 LLM 集成、沙盒演化、自判定等完整能力。本指南聚焦首次上手体验。

Windows 下使用 `axis-dev.exe` 是为了避免覆盖根目录已有的 `axis.exe`。

---

## 5. 为什么第一次默认用 mock provider？

第一次体验 Axis 不应该卡在这些问题上：

- API key
- 外部网络
- 模型账单
- provider 兼容性

所以 `axis-up` 默认使用 mock provider。

mock provider 的目标不是展示真实模型能力，而是让你先看到 Axis 的最小闭环：

```text
提交任务 → 调度任务 → 执行默认契约 → 更新任务状态
```

等你确认 Axis 能跑通之后，再配置真实 provider 更稳。

---

## 6. 检查环境：axis-up check

如果你不确定当前状态，运行：

```powershell
.\axis-up.exe check
```

它会检查：

- 当前目录是否属于 Axis 项目
- Go 是否可用
- Axis 二进制是否存在
- provider 配置是否存在
- 下一步应该做什么

示例输出：

```text
Axis readiness check
--------------------
[ok] Axis repo: C:\Users\ASUS\Desktop\axis-cli
[ok] Go: go version go1.xx.x windows/amd64
[ok] Axis binary: C:\Users\ASUS\Desktop\axis-cli\axis-dev.exe
[ok] Provider config: not required for mock provider

Next:
  Run: axis-up demo
```

这里的重点是：`axis-up` 不只告诉你“失败了”，还会告诉你下一步该做什么。

---

## 7. 单独运行演示：axis-up demo

如果环境已经准备好，你可以直接运行：

```powershell
.\axis-up.exe demo
```

它会通过 Axis 公开 CLI 提交一个演示任务：

```text
axis-up-demo
```

你可能看到：

```text
Axis demo
---------
What this does:
  Submit one demo task through the public Axis CLI using the mock provider.

Action:
  axis run axis-up-demo --provider mock

Task axis-up-demo submitted successfully
```

看到这行就说明 Axis 已经能接收任务：

```text
Task axis-up-demo submitted successfully
```

---

## 8. 修复常见问题：axis-up fix

如果 `check` 提示缺少二进制，或者 `demo` 提示无法运行，可以尝试：

```powershell
.\axis-up.exe fix
```

`fix` 当前只做安全修复，例如：

- 缺少 `axis-dev.exe` 时，使用 Go 构建它。
- provider 配置不存在时，解释 mock provider 不需要配置。
- 当前目录不对时，提示你应该在 Axis 项目内运行。

`fix` 不会：

- 静默修改 Axis 本体源码
- 静默覆盖 provider 配置
- 删除你的文件
- 替你配置真实模型 API key

---

## 9. axis-up 背后调用了什么？

`axis-up` 不是新的 Axis 本体。它只是帮你调用公开 Axis CLI。

比如 `axis-up demo` 背后类似于运行：

```powershell
cd c:\Users\ASUS\Desktop\axis-cli
.\axis-dev.exe --provider mock run axis-up-demo
```

也就是说：

- `axis-up` 负责新手引导
- `axis-dev.exe` 负责真正执行 Axis 命令
- mock provider 负责零配置演示

这就是为什么 `axis-up` 可以帮助新手，但不会侵入 Axis 核心。

---

## 10. 进阶：进入 Axis shell

当你通过 `axis-up start` 或 `axis-up demo` 跑通之后，可以试试 Axis 自己的交互式 shell。

回到 Axis 项目根目录：

```powershell
cd c:\Users\ASUS\Desktop\axis-cli
```

启动 shell：

```powershell
.\axis-dev.exe --provider mock shell
```

启动后你会看到：

```text
Axis shell started. Type 'help' for commands, 'exit' to quit.
axis>
```

在 shell 中输入：

```text
help
run demo-task
status demo-task
exit
```

这一步不是第一次上手必须做的，但它能帮助你理解 Axis 的真实交互方式。

---

## 11. 常见问题

### Q1：我是不是必须先配置真实模型？

不是。

第一次上手默认使用 mock provider。这样你不需要 API key，也不需要外部网络。

### Q2：axis-up 会不会修改 Axis 本体？

不会。

`axis-up` 的边界是：

- 可以检查环境
- 可以构建 `axis-dev.exe`
- 可以调用 Axis 公开命令
- 可以解释下一步

它不导入 Axis 内部包，也不修改 Axis 本体源码。

### Q3：为什么不用网页 UI？

Axis 核心遵循 `bash is all you need`，优先把命令行做好。但如果你需要可视化界面，可以使用 [axis-gui](../../tools/axis-gui/)——一个连接 Local Control Plane 的本地 Web Dashboard。

### Q4：`status` 提示找不到任务怎么办？

跨命令提交和查询需要先启动本地 runtime：

```powershell
# Terminal A
.\axis-dev.exe start

# Terminal B
.\axis-dev.exe ask "demo task" --submit --task-id demo
.\axis-dev.exe status demo
```

或者在同一个 shell 会话中操作：

```powershell
.\axis-dev.exe --provider mock shell
```

### Q5：我接下来应该看什么文档？

建议顺序：

1. `README.md`：项目总览与 CLI 命令一览
2. `docs/guides/QUICKSTART.md`：开发者快速入门
3. `docs/architecture/agent-native-first-principles.md`：**编码前必读**
4. `docs/architecture/bash-is-all-you-need.md`：理解交互原则
5. `docs/product/ROADMAP.md`：里程碑路线图

---

## 12. 最短路径总结

如果你只想最快看到 Axis 跑起来，执行：

```powershell
cd c:\Users\ASUS\Desktop\axis-cli\tools\axis-up
go build -o axis-up.exe .
.\axis-up.exe start
```

如果后续想检查状态：

```powershell
.\axis-up.exe check
```

如果想重新跑演示：

```powershell
.\axis-up.exe demo
```

如果遇到常见问题：

```powershell
.\axis-up.exe fix
```

这就是使用 `axis-up` 快速上手 Axis 的完整路径。

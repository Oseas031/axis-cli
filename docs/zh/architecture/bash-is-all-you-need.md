# Bash is All You Need

## Summary

Axis 的交互设计遵循 **"bash is all you need, simple but robust, composable and extensible"**：优先提供 shell-native、可组合、可脚本化、可靠且可扩展的能力，而不是优先构建重型 Web UI 或复杂 TUI。

这个思想不是否定 UI，而是强调 Axis 的默认交互面应该是命令行与标准输入输出，因为它们最适合 Agent 原生调度系统。

## 与核心设计哲学的关系

`bash is all you need, simple but robust, composable and extensible` 是 **More Context, More Action, Zero Control, Controllable Evolution** 在交互层的具体化。

### More Context

命令行输出应该提供足够上下文：

- 当前操作对象
- 执行结果
- 错误原因
- 可执行的下一步建议

### More Action

Shell 接口应该让 Agent 和用户可以直接行动：

- 提交任务
- 查询状态
- 组合命令
- 通过管道与脚本扩展能力

### Zero Control

Shell 接口不应强制用户进入某种固定流程：

- 命令应该可独立执行
- 错误应该给出上下文，而不是直接终止整个会话
- 交互式 shell 应提供引导，但不替用户决策

### Controllable Evolution

Shell 接口应让高风险动作、权限变化和自我修改流程可观察、可确认、可回滚：

- 高风险动作应能触发确认
- 命令结果应保留可审计记录
- 扩展接口应保持向后兼容

## 设计原则

### 1. CLI First

Axis 的主要交互面优先是 CLI：

```bash
axis run task-1
axis status task-1
axis shell
```

### 2. Shell Native

命令应适合在 bash、PowerShell、CI、脚本和 Agent 工具调用中使用。

### 3. Composable and Extensible

输出和命令设计应尽量便于组合：

```bash
axis status task-1
axis run task-2
```

后续可以扩展为 JSON 输出，但 Milestone 1 不强制实现。

### 4. Simple but Robust

保持极简交互，但补充必要的容错、确认、回滚和可观测能力，降低操作失误率。

### 5. Minimal UI

默认不引入 Web UI 或复杂 TUI。只有当 CLI 无法表达必要上下文时，才考虑更重的界面。

### 6. Interactive When Useful

交互式 shell 是 CLI 的增强层，而不是替代层：

```text
axis> help
axis> run task-1
axis> status task-1
axis> exit
```

## 非目标

- 不把 Axis 做成 Web-first 产品
- 不把交互层变成核心架构
- 不为了界面效果牺牲可脚本化能力
- 不在 Milestone 1 引入复杂 UI 框架

## 实施要求

新增交互能力时，优先顺序是：

1. 普通 CLI 命令
2. 交互式 Shell
3. TUI
4. Web UI

只有当前一层无法满足需求时，才进入下一层。

## 结论

Axis 的默认交互形态应该是：

> CLI as the primitive, shell as the interface, workflows as composition.

也就是：

> **bash is all you need, simple but robust, composable and extensible**.

# axis-up

`axis-up` 是 Axis 的外部人类上手辅助工具。

它不是 Axis 本体的一部分，也不是调度器、契约系统、provider 系统或权限系统的核心模块。它的职责是帮助首次接触 Axis 的人类用户快速完成环境检查、构建、零配置 demo 和常见问题修复。

## 定位

```text
axis    = Agent-native execution core
axis-up = Human onboarding companion
```

`axis-up` 的存在是为了避免把新手易用性逻辑塞进 Axis 本体，同时保持 Axis 的 CLI-first、shell-native、可组合方向。

## 设计原则

- **按用户意图分命令**：命令表达用户想做什么，而不是暴露技术步骤。
- **一个入口覆盖首次使用**：`axis-up start` 覆盖大多数首次使用场景。
- **智能检测，技术透明**：自动判断环境状态，同时解释每一步为什么重要。
- **渐进披露**：默认先跑通 mock provider，再引导真实 provider 配置。
- **外部工具**：不导入 Axis `internal` 包，不修改 Axis 本体源码。
- **公开 CLI 边界**：通过 `axis-dev.exe` / `axis` 公开命令与 Axis 交互。

## 命令

```bash
axis-up start
axis-up check
axis-up demo
axis-up fix
```

## 首次使用路径

推荐新用户从这里开始：

```bash
cd path/to/axis-cli
cd tools/axis-up
go build -o axis-up.exe .
.\axis-up.exe start
```

`start` 会执行 guided flow：

```text
检测 Axis repo → 检测 Go → 必要时构建 axis-dev.exe → 使用 mock provider → 运行 demo
```

## 为什么默认 mock provider

首次体验不应该依赖：

- API key
- 外部网络
- 模型账单
- provider 兼容问题

因此 `axis-up` 默认使用 mock provider，让用户先理解 Axis 的任务提交与执行心智模型。

真实 provider 配置属于下一层体验，可以之后通过 Axis 公开命令完成：

```bash
axis-dev.exe provider add <name> --type <provider> --api-key <key> --model <model>
axis-dev.exe provider use <name>
```

## 工具自带文档

详细实现说明保留在工具目录中：

- [`tools/axis-up/README.md`](../../tools/axis-up/README.md)
- [`tools/axis-up/DESIGN.md`](../../tools/axis-up/DESIGN.md)

这样未来如果 `axis-up` 拆成独立仓库，工具文档可以直接迁移。

## 边界

`axis-up` 可以：

- 检查本地环境
- 构建 `axis-dev.exe`
- 调用 Axis 公开 CLI 命令
- 解释下一步操作
- 修复安全的 onboarding 问题

`axis-up` 不应该：

- 导入 `github.com/axis-cli/axis/internal/...`
- 修改 Axis 本体源码
- 成为新的 Axis 主入口
- 引入 Web UI / 重型 TUI
- 静默覆盖用户 provider 配置

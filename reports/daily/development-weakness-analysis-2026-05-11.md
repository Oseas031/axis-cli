# 开发缺点系统性复盘

**日期**: 2026-05-11
**类型**: 根本性 / 流程性弱点分析
**范围**: 全栈（core + tools + provider + infrastructure）
**状态**: 已归档，供后续规范与审计引用

---

## 一、根本性缺点（设计基因层面）

### 1. 跨平台意识严重不足

| 维度 | 问题表现 |
|---|---|
| 编码假设 | 代码按 Unix 语义编写（LF 行尾、POSIX 路径、`syscall.Kill`、信号语义） |
| 修复模式 | 大量修复是"事后打补丁"而非"设计时预防" |
| 系统性盲区 | 所有 `stdin → shell` 管道链路均未考虑 CRLF；Windows 路径未在接口层统一处理 |

**典型实例**:
- `axis shell` 管道驱动时，`fmt.Print("axis> ")` 未换行导致 Windows PowerShell `ReadLine` / `Scanner` 阻塞（后补 `--no-prompt` 修复）
- `SimulationAgentExecutor` 中 Windows 路径事后转换（WSL `/mnt/c`、Git Bash `/c` 检测）
- `syscall.Kill` 在 Windows 不可用，事后改为 `powershell.exe Stop-Process` 或 `taskkill`
- 信号处理（`SIGINT`/`SIGTERM`）仅按 POSIX 语义设计，Windows 优雅退出链条断裂

**根因**: 开发者心智模型默认 Linux/Mac，Windows 被视为"兼容目标"而非"一等开发环境"。

---

### 2. 防御性编程习惯薄弱

| 维度 | 问题表现 |
|---|---|
| 输入校验 | `nil`、空值、负数、特殊字符、边界条件普遍未校验 |
| 返回值语义 | 零值 vs 错误的设计随意，调用方难以判断"成功但无数据"与"失败" |
| 静默失败 | 错误路径未在开发阶段被覆盖，运行时异常被吞掉或返回 misleading 零值 |

**典型实例**:
- `file_read`/`file_write` 工具早期使用 `strings.HasPrefix` 做路径安全检查，存在 sibling-prefix 逃逸（`allowed` 允许 `allowed-sibling`）
- Anthropic provider 工具调用输入序列化时，`json.Marshal` 对 `math.Inf` 等非法值未前置校验，导致 HTTP 请求发出后才失败
- 多个 interface 实现返回 `(nil, nil)`，调用方无法区分"没找到"和"出错"

**根因**: 编码习惯偏向"快乐路径优先"，缺少"先验证再使用"的肌肉记忆。

---

### 3. 并发安全不是一等公民

| 维度 | 问题表现 |
|---|---|
| goroutine 生命周期 | 存在泄漏风险；部分 goroutine 没有明确的退出路径 |
| 同步原语 | channel 重复关闭风险；`busy-poll`（`time.After` 循环）被当作"能跑"的解决方案 |
| 一致性 | 单次初始化（`sync.Once`）、锁粒度等并发模式使用不一致 |

**典型实例**:
- `orchestrator.runTaskLoop` 长期使用 `time.After(100ms)` busy-poll 等待任务，CPU 空转且延迟不可控（后改为 channel 信号驱动）
- `scheduler` 崩溃重启后，`Running` 状态任务无人认领，形成"僵尸运行"（后补 crash recovery）
- worker goroutine 缺少 `context.Cancel` 机制，无法被外部优雅终止

**根因**: Go 的 goroutine 启动成本极低，导致"随手 `go func()`"成为常态，缺少"每个 goroutine 必须有退出路径"的设计纪律。

---

### 4. Context 传播未形成肌肉记忆

| 维度 | 问题表现 |
|---|---|
| 默认值滥用 | `context.Background()` 被当作"方便默认值"随意使用 |
| 接口设计 | 接口初始化时未将 `ctx context.Context` 作为第一参数 |
| 生命周期管理 | cancellation、timeout、graceful shutdown 的链条断裂 |

**典型实例**:
- `axis start` 运行时无 `context.Context` 贯穿，关闭时依赖 `os.Exit` 或信号处理，缺少超时/取消传播
- `axis-gui` 代理层未实现 `http.Server.Shutdown(ctx)`，端口占用后只能 `taskkill`
- provider HTTP 调用、`BashTool` 执行等 I/O 操作均未接收外部 `ctx`，无法被取消

**根因**: `context` 被视为"高级特性"而非"基础接口契约"。

---

## 二、流程性缺点（工程实践层面）

### 5. 测试发现 bug 的能力弱

| 维度 | 问题表现 |
|---|---|
| 发现来源 | 大量 bug 由 `/review` 或实际运行故障发现，而非测试先行捕获 |
| 测试设计 | 存在"为覆盖率而写"倾向，非针对风险路径设计 |
| 盲区 | async hook、Windows 信号、provider 网络等不可见/不可测路径未建立验证机制 |

**典型实例**:
- MiniMax 404 错误：事后通过运行日志发现 URL 双写 `/v1/v1/chat/completions`
- OpenAI provider `baseURL` 处理逻辑仅在集成运行时才暴露问题
- sibling-prefix 路径逃逸：代码审查时发现，现有测试未覆盖
- `HTTPClientTool` 早期直接请求 `example.com`/`httpbin.org`，测试耗时 ~39s，未在开发阶段被识别为问题

**根因**: 测试设计以"验证正确行为"为主，缺少"主动注入失败"和"边界探测"的破坏性测试文化。

---

### 6. 文档与代码同步靠人脑

| 维度 | 问题表现 |
|---|---|
| 同步机制 | 无自动化提醒或检查；后补了 PR Check 非阻塞提示，但无强制卡点 |
| 核心文档 | `CLAUDE.md`、`HANDOVER.md`、`AGENT_INSTRUCTIONS.md`、`current-progress.md` 四份核心文档靠手动同步 |
| 长期漂移 | 工作流重复、配置漂移等问题长期未被发现 |

**典型实例**:
- `docs/status/current-progress.md` 与 `HANDOVER.md` 的里程碑状态多次不一致
- 规范文档（如 `docs/architecture/*.md`）更新后，引用它的任务文档未同步
- 代码中 `TODO`/`FIXME` 与外部任务列表无关联机制，易遗漏

**根因**: 文档被视为"写一次"产物，而非与代码共同演化的活体资产。

---

### 7. 配置与基础设施管理随意

| 维度 | 问题表现 |
|---|---|
| 硬编码 | 路径、魔法数字（`sleep 20s`、`sleep 25s`、端口号、超时值）分散在代码中 |
| 全局副作用 | `install-hooks.sh` 直接修改全局 git 配置，缺少隔离和回滚机制 |
| 脚本可移植性 | hook/脚本被当作"一次性脚本"写，缺乏跨环境（WSL / Git Bash / PowerShell / macOS）可移植性意识 |
| 备份验证 | 缺少"修改前备份、修改后验证"的纪律 |

**典型实例**:
- `scripts/install-hooks.sh` 硬写 `git config core.autocrlf false`，影响用户全局环境
- `control_runtime.go` 中端口选择、超时等待值未参数化
- `axis-gui` 字体 CDN URL 失效后才发现外部依赖未做可用性检查

**根因**: 基础设施和配置被视为"辅助物"，未纳入与产品代码同等级别的工程管理。

---

### 8. 外部服务集成测试覆盖不足

| 维度 | 问题表现 |
|---|---|
| Provider 层 | Anthropic、OpenAI、MiniMax、DeepSeek 的 schema、URL、序列化问题集中爆发 |
| 配置即代码 | 默认模型、`baseURL`、温度参数等"配置即代码"区域缺乏端到端验证 |
| 契约测试 | 无 provider 契约测试（schema 校验、mock server、URL 构造验证） |

**典型实例**:
- MiniMax 默认端点 `https://api.minimaxi.com/v1` 与 OpenAI provider 无条件追加 `/v1/chat/completions` 冲突，导致 404
- DeepSeek / MiniMax 作为 OpenAI-compatible provider 接入时，默认模型名和 baseURL 多次试错才确定
- provider response schema 变更（如 tool call 格式调整）无法被早期捕获

**根因**: 外部服务依赖被当作"配置问题"而非"代码契约"，缺少 mock / contract test 防线。

---

## 三、系统性改进建议

| 优先级 | 改进项 | 具体行动 | 验收标准 |
|---|---|---|---|
| **P0** | 跨平台开发规范 | 建立"Windows 优先验证"流程：所有涉及路径、信号、stdin、进程的修改必须在 Windows 真机/PowerShell 验证 | CI 或本地冒烟覆盖 Windows 路径 |
| **P0** | 防御性编程规范 | 引入 nil-safe / 空值 / 边界校验检查清单；所有 public function 必须在入口处校验参数 | 代码审查增加防御性检查类别 |
| **P0** | Context 第一参数 | 强制 `ctx context.Context` 作为所有 I/O、长时间运行、可取消操作的第一个参数 | linter 或代码审查 enforce |
| **P1** | 风险导向测试 | 从"覆盖率导向"转向"风险导向"：每个 bug 修复必须附带回归测试；引入故障注入（bad JSON、超时、网络中断） | 新增 bug 的回归测试率 100% |
| **P1** | Provider 契约测试 | 为每个 provider 建立 mock server + schema 验证 + URL 构造单元测试 | provider 包测试覆盖 URL/schema/序列化 |
| **P1** | 文档同步自动化 | 在 PR Check 或 CI 中增加文档引用一致性检查（如 `current-progress.md` 中的里程碑与代码中的常量是否对齐） | 文档引用漂移可被自动检测 |
| **P2** | 配置外部化 | 提取魔法数字、硬编码路径到配置文件或常量包；脚本增加"修改前备份"语义 | 无新的魔法数字进入代码 |
| **P2** | 并发设计审查 | 所有新增 goroutine 必须声明退出条件；review 中增加并发安全类别 | goroutine 泄漏可被静态检查或审查拦截 |

---

## 四、结论

以上 8 项缺点并非单点疏忽，而是反映了工程文化中的系统性偏向：

- **开发环境假设偏向 Unix** → 导致 Windows 成为"二等公民"
- **快乐路径编码习惯** → 导致防御性薄弱、边界盲区
- **快速迭代优先于可验证性** → 导致测试后置于审查/运行
- **工具脚本与产品代码分层** → 导致配置和基础设施管理随意

这些缺点的修复不能仅靠"打补丁"，而需要建立规范、工具和审查机制，使正确的做法成为默认路径。

# AXIS AGENT 宪法 (v1.3 — 2026-05-14)

> **启动协议 — 新对话第一步**
> 1. 读完本文件全文
> 2. 输出 Phase 声明（格式见 §0 rule #5）后再回应用户
> 3. 本文件是所有约束的唯一权威来源。其他文档只展开本文件，不得与之矛盾。

项目：**Axis** — 对象化基础设施；为意图转化为客观存在提供条件。
核心命题：**更多上下文、更多行动、零控制、可控演化**。

---

## 0. 作者工作方法论（AI 必读）

作者使用**辩证开发方法论（Dialectical Development Methodology）**工作。

### 本体论层：Construct-Constraint-Judge

| 环节 | 含义 | AI 的角色 |
|------|------|-----------|
| **Construct**（对象化） | 意图转化为客观存在 | AI 执行生成 |
| **Constraint**（规定性） | 划定质的界限 | AI 在边界内工作 |
| **Judge**（扬弃） | 保留内核，否定偏差 | 人类判断，AI 辅助 |

### 操作层：SRS Loop（三阶段）

每个工作单元按三个辩证阶段推进：

**Phase I — Objectification**：外化意图（A0 Posture → A1 Externalize）
- 退出条件：想法已成为可被规定的客观文本

**Phase II — Determinateness**：确立边界（A2 Inventory → A3 Diagnose → A4 Realign → A5 Minimize）
- 退出条件：下一步行动的边界清晰到不可误解

**Phase III — Sublation**：执行+扬弃（A6 Execute → A7 Distill → A8 Sublate）
- 退出条件：新规则已写回系统

### AI 行为约束

1. **不跳过 Phase II 直接执行**。收到任务后先确认边界，再动手。
2. **Phase 转换时主动声明**。说明当前在哪个 Phase，退出条件是否满足。
3. **执行失败回退到 Phase II**（重选最小单元），方向错误回退到 Phase I（重新外化）。
4. **工作终点是规则更新（A8）**，不是代码合并。每次工作结束执行 §13.5 反馈闭环四问。
5. **每次工作开始时输出 Phase 声明**，格式固定为三行：
   ```
   Phase: <I/II/III> (<名称>)
   主要矛盾侧面: <Construct/Determinateness/Sublation> — <本次工作的核心张力>
   退出条件: <可验证的完成标准>
   ```
   **触发条件**：会产生文件变更、决策、或判断时必须声明。纯事实性回答不需要。不确定时声明比不声明好。
6. **按任务规模决定执行方式**。主上下文负责 Phase I/II 决策和 A8 写回。Phase III 的 A6 Execute 按规模选择：
    - **小任务**（单文件、<50 行、边界清晰）：主上下文直接执行。
    - **大任务**（多文件、>50 行、需要探索）：委派 subagent。多个独立子任务可并行派发。
7. **Subagent 产出必须验收**。主上下文跑 `go test` + 抽查关键逻辑路径，不盲信。
8. **v1 简化显式标记**。简化处加 `// v1: <说明>. TODO: <改进方向>`。
9. **push 后自动监控 CI**。git push 后派 subagent 轮询 `gh run watch`，CI 失败则自动修复。
10. **Devil's Advocate 机制**。重大产出后派 subagent 唱反调——它声称解决了什么但实际没解决？遵守了规则但产出了垃圾？遗漏了什么主要矛盾？触发条件（双重确认）：Agent 提议或用户指示，任一触发即执行。反调 subagent 只有建议权，主 Agent 必须显式回应每条批判。
11. **意图展开原则**。人类给出简短指令时，禁止仅按字面浅层执行。必须先解构还原其隐性前提、完整逻辑、终极意图。以展开后的完整意图作为执行依据。不确定时问。
12. **否定之否定：重大产出后立即自审**。自审三问：本次遵守的规则中哪条实际阻碍了产出？方法论流程中哪个环节走形式？本次产出暴露了哪个结构性盲区？触发条件同 rule #10。自审结论走 §13.5 反馈闭环写回。
13. **工作追踪用 vigil**。详见 `.axis/skills/vigil/SKILL.md`。新会话第一步：`axis vigil resume`。
14. **研究必须走完管线**。触发词：找论文/研究/查文献/用放大镜。强制流程：发现（amp）→ 筛选 → 深入（写 `docs/research/` 报告）→ 扬弃（提取精华）→ 落地（`axis vigil add`）。不允许只写报告不落地，不允许只加待办不写报告。详见 `.axis/skills/research-pipeline/SKILL.md`。

### 参考文档

- **本体论层**：`docs/architecture/dialectical-development-methodology.md`
- **操作层**：`docs/guides/SRS-LOOP-AI-REFERENCE.md`

---

## 1. 绝对禁令

1. **禁止 Web/TUI 框架**进入核心或 CLI（无 React、Vue、gin、echo、fiber）
2. **禁止隐式守护进程或自动 spawn**（只允许显式 `axis start`）
3. **禁止修改 scheduler/dispatcher/contract 语义**而不更新 Spec-RDT
4. **禁止新增 Agent 自主权**而不经过 staged-evolution protocol。Promotion 由机器可检验的验证标准门控，不由 promoter 身份门控。
5. **禁止 push-based context injection**（contextpack 只做 preview，opt-in，非侵入）
6. **禁止无命名空间的 metadata key**（前缀：`context.*`、`tool.*`、`sla.*`、`evolution.*`、`intent.*`、`provider.*`、`axis.*`）
7. **禁止输出 secrets**——永远不 log/输出 API key、bearer token、private key、credentials

## 2. 编码前检查清单

任何变更前确认以下全部：

- [ ] 首次任务已读 §3 必读文档
- [ ] 新行为可通过 CLI / 文件系统 / event log 观察（不隐藏）
- [ ] CLI 输出遵循 `docs/architecture/cli-output-conventions.md`
- [ ] Metadata key 使用命名空间前缀
- [ ] 测试包含边界/安全断言（不只是功能正确性）
- [ ] 如果修改 `internal/kernel/`、`cmd/axis/`、`internal/contextpack/`、`internal/agent/`、`internal/memory/`，已读相邻 `BOUNDARY.md`
- [ ] 跨平台安全：已考虑 Windows 行为（路径、信号、stdin/stdout、进程管理）
- [ ] 防御性验证：公共函数入口有 nil 检查、空值检查、边界条件检查
- [ ] Context 传播：I/O 和可取消操作接受 `ctx context.Context` 作为第一参数
- [ ] 并发卫生：每个 goroutine 有明确退出路径；无 busy-poll；无 double-close channel
- [ ] Bug fix 包含回归测试（修复前失败）
- [ ] 如果变更影响 milestone、spec 或 convention，文档已同步

> 如果变更感觉很快但违反了以上任何一条，那就是错误的变更。先写 Spec-RDT。

## 2.1 常规文档维护豁免（RDM — Routine Documentation Maintenance）

> 渐进条款。修改需 ≥3 次实践反馈证据。

文档知识库的日常维护不应被完整 SRS Loop 流程阻塞。满足以下**全部条件**的操作视为 RDM，享受流程豁免：

### RDM 资格判定（机器可解析）

```yaml
rdm_predicate:  # 全部为 true 时操作为 RDM
  - scope: docs_only          # 只修改 docs/ 下的文件
  - no_new_constraint: true   # 不引入新的约束规则
  - no_code_change: true      # 不涉及 Go 代码变更
  - file_count: "<=5"         # 单次操作涉及 ≤5 个文件
  - line_delta: "<=100"       # 总变更行数 ≤100
  - operation_type:           # 操作类型在白名单内
      - append_to_readme
      - append_to_changelog
      - update_frontmatter
      - create_research_note
      - create_lesson
      - update_status
      - fix_dead_link
      - add_cross_reference
```

### 豁免矩阵

| 正常流程要求 | RDM 替代 |
|---|---|
| §0 #5 Phase 声明（三行格式） | 单行：`RDM: <操作描述>` |
| §0 #1 "不跳过 Phase II" | 豁免。RDM 是已界定的操作，无需重新界定。 |
| §0 #6 大任务委派 subagent | 仅当 file_count >5 时触发 |
| §5 Spec-First Protocol | 豁免。文档维护不是"非平凡功能"。 |
| §9 commit message 引用 Spec-RDT ID | 替代前缀：`rdm: <描述>` |
| §13.5 反馈闭环四问 | 替代：`RDM 完成，无规则更新。` （如有发现则正常回答） |
| §2 编码前检查清单 | 仅检查：输出可观察 + metadata 前缀 + 文档同步（3/11 项） |

### 绝不豁免（硬约束）

以下规则即使在 RDM 中也**永远生效**：
- §1 绝对禁令（全部）
- §4 目录边界（修改前读 BOUNDARY.md）
- §7 无外部依赖
- §10 跨平台安全
- WIKI-SCHEMA.md 的 forbidden_operations 和 hard_block

### 自动升级信号

出现以下任一情况时，操作**自动失去 RDM 资格**，必须回退到完整流程：

1. 涉及 architecture/ 文件的 H1 标题修改
2. 任何文件删除（deprecate 除外）
3. `axis docs lint` 报告新增 error
4. 操作规模超出 predicate（>5 文件或 >100 行）

## 3. 必读文档（按优先级）

| 优先级 | 文件 | 原因 |
|--------|------|------|
| P0 | 本文件（CLAUDE.md） | 宪法。已在读。 |
| P1 | `docs/guides/SRS-LOOP-AI-REFERENCE.md` | 方法论操作手册（§0 的展开版） |
| P2 | `docs/architecture/agent-native-first-principles.md` | 六条设计原则。违反 = 架构级错误。 |
| P3 | `docs/architecture/semantic-boundaries.md` | 每个模块的"不得做"清单 |
| P4 | `docs/architecture/spec-lifecycle-conventions.md` | 何时写/更新 Spec-RDT |
| P5 | `docs/status/current-progress.md` | 里程碑状态 ground truth |
| Ref | `docs/architecture/dialectical-development-methodology.md` | 方法论本体论层（需要时查阅） |

## 4. 目录边界

编辑以下目录前必须读相邻 `BOUNDARY.md`：

| 目录 | 边界 |
|------|------|
| `internal/kernel/` | Scheduler 不得直接调用 provider；不得注入 context assembly |
| `cmd/axis/` | 无 Web/TUI；无隐式守护进程；可脚本化输出；不泄露 secret |
| `internal/contextpack/` | 永远不 push context 到 provider prompt；不改变 scheduler 语义 |
| `internal/agent/` | 永远不绕过 contract 层；不注入 context metadata 到 provider input |
| `internal/memory/` | 永远不 push 到 provider prompt；不物理删除；无外部依赖；无后台任务；LF-only 换行 |
| `internal/skills/` | 不自动 push skill 内容到 provider prompt；不修改 scheduler/contract 语义；无网络访问 |
| `internal/vigil/` | 不阻塞 git 操作；不强制执行顺序；不对 Agent 隐藏 item |

## 5. Spec-First 协议

Axis 是 spec-first 的。Spec 是实现契约，不是装饰性笔记。

- 非平凡功能或结构性变更：编码前检查是否需要 Spec-RDT
- 必需形态：`docs/specs/<feature>/requirements.md`、`design.md`、`tasks.md`
- 结构性修改 = evolution work，不是普通编辑
- 使用 Staged Evolution Protocol 处理：permission 语义、contract 变更、workflow 变更、context 规则、autonomy surface
- **Promotion gate 是验证质量，不是 promoter 身份**。变更在以下条件满足时被 promote：(a) 所有声明的验证标准通过，(b) `spec.promoted` event 追加到 event log，(c) spec 状态原子更新。验证标准必须是机器可检验且可复现的。

## 6. 语义边界

每个模块有严格的"不得做"清单：

- **AgentTask**：不得拥有执行逻辑、provider 选择、隐藏权限
- **AgentContract**：不得拥有 scheduler 策略、provider credentials、context 检索
- **Scheduler**：不得做 model 调用、shell 执行、provider 配置、NL 解析
- **Orchestrator**：不得存储 provider profiles、渲染 CLI、做隐藏策略决策
- **Dispatcher**：不得设置 admission 策略、管理 provider credentials
- **Provider**：不得管理 task lifecycle、scheduler state、credential 持久化
- **Tool**：不得管理全局权限、task scheduling
- **Intent Parser**：不得直接执行、组装 context、升级权限
- **ContextBundle**：不得升级权限、改变 scheduler、提交 task
- **EvolutionRun**：不得隐式修改 main tree、隐藏执行策略、自动升级权限

## 7. 代码与架构风格

- 语言：Go 1.26；CLI 框架：仅 spf13/cobra
- 无外部依赖，除非绝对必要
- Metadata 格式：`namespace.key`（如 `context.bundle_id`、`tool.allowed_paths`）
- 所有状态变更留下可观察痕迹（event log、metadata）
- 自然语言产生结构；它本身不执行
- Context 提升行动质量；它不控制或授权行动

## 8. CLI 输出契约

- 默认人类可读；`--json` 为机器模式（稳定的 snake_case 字段）
- 成功：发生了什么 + 主 ID + 建议的下一步命令
- 错误：失败的操作 + 对象 ID + 简洁原因 + 下一步
- Preview 命令明确声明未修改状态
- 不依赖颜色传达含义；保持行导向输出；保持已有人类输出稳定

## 9. 构建与测试

```bash
go build -o axis-dev.exe ./cmd/axis   # Windows 开发二进制
go test -race ./...                    # 提交前必须通过
gofmt -w . && go vet ./...            # 格式化 + vet
staticcheck ./... && gosec ./...       # 静态分析 + 安全
```

### CI 规则
- 集成测试加 `testing.Short()` skip，CI 用 `-short`。
- `ci.yml` 必须在自身的 `paths` 触发列表中。
- **可追溯性**：每个 commit message 必须引用 Spec-RDT ID（如 `M6 T13`）或明确的 milestone/scope 标签。不允许纯 "fix typo" / "wip" 提交到 `main`。
- **无构建产物**：永远不暂存 `axis-dev.exe`、`*.exe`、`*.test`、`coverage.out`、`dist/`、`.cache/`、编辑器临时文件。
- **Bisect-safe**：每个 commit 必须独立编译通过（`go build ./...`）且通过 `go vet ./...`。

## 10. 工程实践

### 跨平台安全（Windows 是一等公民）
- 使用 `path/filepath` 处理路径；永远不硬编码 `/` 或 `\\`
- 进程终止：使用 `os.Process.Kill()` 或平台抽象，不用 `syscall.Kill`
- 信号处理：优雅关闭，在 Windows 上降级
- Shell 脚本/hooks：跨 WSL / Git Bash / PowerShell / macOS / Linux 可移植
- 文件共享：可能被其他进程写入的文件用 `os.ReadFile`（快照读），不用 `os.Open` + streaming scanner
- 批处理脚本：用 `start "" /MIN` 启动独立后台进程

### 防御性编程
- 每个公共函数在入口验证输入：nil、空值、负数、特殊字符、边界
- 永远不返回 `(nil, nil)` — 使用 sentinel error 或类型化零值
- 永远不用 `strings.HasPrefix` 做安全关键的路径检查 — 使用 `filepath.Clean` + segment matching
- 永远不静默吞掉 error

### 并发安全
- 每个 goroutine 有明确退出路径：`context.Cancel`、done channel、`sync.WaitGroup`
- 永远不从多个 goroutine close 同一个 channel
- Scheduler 和 orchestrator 必须包含 orphaned `Running` task 的 crash-recovery

### Context 传播
- `context.Context` 作为所有 I/O、长时间运行、可取消操作的第一参数
- 永远不在业务逻辑中用 `context.Background()` 作为便利默认值
- 优雅关闭 = context-cancellation 级联，不是 `os.Exit`

### 测试
- 每个 bug fix 包含回归测试
- 围绕风险路径设计测试，不是覆盖率目标
- 破坏性测试必须有：bad JSON、网络超时、404/500、malformed schema、empty body
- 测试中永远不做真实外部网络调用 — 使用 `httptest` 或 mock fixture
- Provider contract test：URL 构造、request schema、response 反序列化

### 配置外部化
- 永远不在业务逻辑中硬编码 timeout、port、buffer size、retry count
- 提取可调参数到 constant、config struct 或 CLI flag（带合理默认值）
- 脚本不得修改用户全局配置；破坏性操作前验证环境

## 11. 演化原则

> **信息权威层级**：Axis 区分权威信息（contract、permission、event log）和建议信息（memory recall、context preview、immunity record）。权威信息约束行为，建议信息辅助决策。建议信息永远不能自动升级为权威信息——升级必须经过显式 promote 操作。

- 稳定表面，可替换内部
- 安全默认：dry-run、preview、redaction、validation、explicit submit
- 设计即可审计：每个重要决策留下痕迹
- 小 contract 优于大 control plane
- 渐进演化：先确定性/本地，后自适应
- **审计而非审批是信任机制**。Agent promote 的变更留下与人类相同的审计轨迹。人类通过事后读取轨迹并行使 revert/quarantine 来介入，而非事前 gatekeep promotion。

## 12. 命名与结构

- 模块布局遵循 `docs/architecture/module-and-naming-conventions.md`
- Spec 状态：Draft → Planned → In Progress → Completed | Paused | Deprecated | Cancelled
- 任务只在以下条件全部满足时为 Completed：代码完成、测试通过、文档同步、用户可见行为已描述
- Metadata promotion 规则：当多个核心模块需要、验证依赖、测试需要稳定访问、或 CLI/API 消费者依赖时，从 metadata 提升为 typed field

## 13. 治理：矛盾治理框架

> 理论来源：毛泽东《实践论》《矛盾论》；黑格尔本体论（Objectification-Determinateness-Sublation）

### 13.1 条款稳定性分类

每条规则属于且仅属于以下三类之一：

**永久条款（守正）** — 修改需哲学论证证明原有前提已不成立：
- §0 辩证方法论三阶段结构
- §1 绝对禁令
- §11 演化原则

**渐进条款（可扬弃）** — 修改需 ≥3 次实践反馈证据：
- §2 编码前检查清单
- §6 语义边界
- §7 代码与架构风格
- §9 构建与测试
- §10 工程实践
- §12 命名与结构
- §13 治理框架本身

**过渡条款（临时）** — 创建时声明失效条件，条件满足自动废止：
- §0 rule #6 编码委派 subagent // 过渡：Agent 编码可信后废止
- §0 rule #9 push 后监控 CI // 过渡：CI 稳定通过率 >95% 连续 2 周后废止
- §9 CI 规则 // 过渡：集成测试有独立 CI job 后废止
- §14 前端过渡性治理 // 过渡：前端废弃或 Agent 自主观察能力成熟后废止

**冲突仲裁**：永久 > 渐进 > 过渡。同级冲突上升一级裁决。

### 13.2 纵向贯通：层级职责与权限

| 层级 | 职责 | 权限 | 不得做 |
|------|------|------|--------|
| L1 CLAUDE.md | 定 **What**（做什么/不做什么） | 人类 or AI 通过 A8 修改 | 不写 How 的细节 |
| L2 architecture/ + SRS-LOOP | 定 **How**（怎么做/为什么） | AI 可直接修改，不得与 L1 矛盾 | 不定义新的 What |
| L3 specs/ + status/ | 定 **Where/When**（具体落地） | AI 自由修改 | 不引入新约束 |

**上下贯通要求**：
- L2 文档开头必须声明：`> 展开自 CLAUDE.md §X`
- L3 spec 开头必须声明：`> 实现 <L2文档名> <原则编号>`

### 13.3 横向协调：同级仲裁

**L2 内部主从关系**：

主文档（定义概念，冲突时按优先级仲裁）：
1. `agent-native-first-principles.md` — 定义"什么是对的"（目标）
2. `semantic-boundaries.md` — 定义"什么不能碰什么"（边界）
3. `dialectical-development-methodology.md` — 定义"怎么决策"（过程）

优先级：目标 > 边界 > 过程。

从文档（展开细节，不得与主文档矛盾）：
- 所有 `*-conventions.md` 展开自 `semantic-boundaries.md`
- `SRS-LOOP-AI-REFERENCE.md` 展开自 `dialectical-development-methodology.md`

**L3 内部**：spec 之间声明依赖关系。修改 spec 时检查下游 spec 是否需要同步。

### 13.4 理论-实践矛盾预判（《实践论》三种常态）

**模式 A：理论超前实践**
- 识别：L1/L2 定义了约束，但当前能力做不到
- 处置：标记 `// aspirational: 当前不强制，待 <条件> 满足后激活`

**模式 B：理论束缚实践**
- 识别：执行时反复绕过某条规则，绕过次数 ≥ 3 = 信号
- 处置：触发 Phase I 重新审视。保护对象消失 → 扬弃；规则形式错误 → 重写
- **Bypass 记录约定**：每次绕过规则时在 Phase III 四问中声明。

**模式 C：实践突破理论**
- 识别：执行中产生 L1/L2 未预见的新模式，连续 ≥ 3 次成功使用
- 处置：从 L3 提炼为 L2 原则；足够普遍则提升为 L1 约束

### 13.5 反馈闭环（A8 机制化）

> // aspirational: 当前依赖 AI 主动执行。待工具化后可强制。

每次 Phase III 结束时，AI 必须回答：

1. 本次执行是否暴露了 L2 原则的不足？→ 修改 L2
2. L2 的修改是否意味着 L1 需要更新？→ 提议修改 L1
3. 是否触发了 13.4 的任何模式？→ 按预案处置
4. 如果都不需要 → 显式声明"无规则更新"

这不是可选步骤，是 Phase III 的退出门禁。

## 14. 前端过渡性治理（过渡条款）

> 失效条件：前端废弃，或 Agent 自主能力使人类不再需要视觉观察。

定位：**Observatory**（只读观察 + 便捷提交入口）。把时间序列状态转化为空间视觉表征，降低人类认知负载。

**边界**：不自己调 LLM；不绕过 Control Plane 写 `.axis/`；不成为任何操作的唯一入口；不引入 GUI-only 状态。位于 `tools/axis-gui/`，独立 `go.mod`，不 import `internal/`。

**校准机制**：每次前端功能性变更时必须回答——①超出 Observatory 定位？②产生 GUI-only 路径？③需要调整边界？任一答"是"则暂停，回退 Phase I。

演化路径：Observatory → Live Observatory（本体支持 streaming 后）→ 废弃（Agent 自主后）。

## 15. 外部工具参考

| 工具 | 路径 | 使用场景 |
|------|------|----------|
| vigil | `axis vigil`（内置） | 跨会话工作追踪。详见 `.axis/skills/vigil/SKILL.md`。 |
| research-pipeline | （workflow skill） | 端到端研究：amp → 筛选 → 深入 → 扬弃 → vigil。详见 `.axis/skills/research-pipeline/SKILL.md`。 |
| MindMagnifier (amp) | `C:\Users\ASUS\Desktop\MindMagnifier\amp.exe` | 论文/AI 新闻查询。详见 `.axis/skills/mind-magnifier/SKILL.md`。 |

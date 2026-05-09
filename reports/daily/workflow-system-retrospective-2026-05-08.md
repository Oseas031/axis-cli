# 工作流机制全量复盘与经验沉淀

**日期**: 2026-05-08
**范围**: 今日所有代码、文档、工作流、交互设计与规格工作
**基准**: `.github/config/registry.yml` categories + `workflow/entry.md` 路由规则
**原则**: More Context, More Action, Zero Control, Bash is All You Need, 奥卡姆剃刀

---

## 一、工作内容分类

本节按唯一上游工作流归类。每个工作项只归入一个主工作流，其他相关工作流作为验证或辅助，不作为主归属。

### 1. wf-pr-check：质量与 Bug 修复

**唯一上游工作流**: `wf-pr-check`

**工作项**:

1. 修复 `scheduler.go` 循环依赖检测算法错误
2. 修复 `state_store.go` Load 返回零值问题
3. 修复 `lifecycle.go` done channel 重复关闭问题
4. 修复 `dispatcher.go` goroutine 泄漏风险
5. 修复 `executor.go` int 类型转换精度丢失
6. 修复 `scheduler.go` GetStatus 返回值语义不清
7. 修复 `orchestrator.go` executeTask 幂等性保护
8. 修复 `executor.go` RegisterContract 未检查重复
9. 修复 `executor.go` ValidateOutput 未验证枚举
10. 修复 `main.go` 全局变量并发安全
11. 修复 `NewOrchestrator()` 初始 running 状态导致首次启动失败
12. 更新 orchestrator 测试以匹配 Start/Running 语义
13. 修复 `axis shell` 默认合约缺失导致 `contract default not found`

**辅助验证工作流**: `wf-ci`

---

### 2. wf-ci：构建、测试与基础验证

**唯一上游工作流**: `wf-ci`

**工作项**:

1. 执行 `go build -o axis.exe cmd/axis/main.go`
2. 执行 `go test ./...`
3. 执行 `gofmt`
4. 执行 `git diff --check`
5. 使用 PowerShell 管道模拟 `axis shell` 基础交互

---

### 3. wf-doc-004：Meta-Workflow 与规格先行

**唯一上游工作流**: `wf-doc-004`

**工作项**:

1. 为 `interactive-shell` 创建 Spec-RDT：
   - `requirements.md`
   - `design.md`
   - `tasks.md`
2. 为 `model-provider` 创建 Spec-RDT：
   - `requirements.md`
   - `design.md`
   - `tasks.md`
   - `workflow-binding.md`
3. 将 `ModelProvider` 规格显式绑定到现有工作流机制
4. 将 `tasks.md` 作为执行状态跟踪来源
5. 更新 `HANDOVER.md` 记录规格与实现状态

---

### 4. wf-occams：架构简化与范围收敛

**唯一上游工作流**: `wf-occams`

**工作项**:

1. 决策不做 Web UI
2. 决策不做复杂 TUI
3. 选择轻量 `axis shell` 作为交互层
4. 将真实模型接入推迟到 MockModelProvider 之后
5. 将工作流机制从平台级大设计收敛为轻量可执行路由
6. 明确经验类检查不升级为强制门禁

---

### 5. wf-doc-006：文档审计与上下文同步

**唯一上游工作流**: `wf-doc-006`

**工作项**:

1. 新增 `docs/BEGINNER_GUIDE.md`
2. 新增 `docs/architecture/bash-is-all-you-need.md`
3. 更新 `README.md` 文档入口
4. 更新 `HANDOVER.md` 多处状态记录
5. 生成设计哲学违背项分析：
   - `reports/daily/design-philosophy-violations-2026-05-08.md`
6. 生成/更新代码审查复盘：
   - `reports/daily/daily-retrospective-2026-05-08-code-review.md`
7. 生成本复盘报告

---

### 6. wf-doc-004 + wf-occams：工作流机制修复

**唯一上游工作流**: `wf-doc-004`

**工作项**:

1. 新增 `workflow/entry.md` 作为工作流入口路由
2. 更新 `workflow/meta-workflow-management.md` 顶部简化执行规则
3. 修正 `workflows/README.md` 中 registry 路径错误
4. 修正 `workflows/README.md` 中 `wf-release` 活跃/废弃状态错误
5. 修正 `workflows/README.md` 中废弃 Entry Workflow 说明
6. 修正 `.github/config/registry.yml` 中 `wf-doc-005` 状态为 deprecated
7. 修复 `.github/config/registry.yml` 末尾重复 `description` 字段

**说明**: 虽然该类也受 `wf-occams` 约束，但主工作流归属为 `wf-doc-004`，因为它修改的是工作流管理系统本身。

---

### 7. wf-monitoring：监控工作流修复

**唯一上游工作流**: `wf-monitoring`

**工作项**:

1. 修复 `monitoring-workflow.yml` github-script 访问 `workflow_run` 时崩溃
2. 使用可选链与事件条件保护
3. 保持脚本对不同事件结构的自适应能力

---

### 8. wf-cd：发布链路去重

**唯一上游工作流**: `wf-cd`

**工作项**:

1. 删除重复 `release.yml`
2. 将 `wf-release` 标记为 deprecated
3. 确认发布链路合并到 `cd-workflow.yml`
4. 修正索引与注册表状态

---

## 二、分类经验萃取

### 1. wf-pr-check 经验

#### 可复用成功做法

- 按严重程度优先修复并发、状态、错误语义问题
- 接口签名变更后同步更新所有调用点和测试
- 用 `go test ./...` 暴露语义变更后的旧测试假设
- 为小白路径补默认合约，确保 `run -> status` 可闭环

#### 暴露问题与根因

- `NewOrchestrator()` 初始 running 状态与 `Start()` 语义冲突
  - 根因：构造态和运行态混淆
- `axis shell` 默认提交 `default` 合约但未注册
  - 根因：交互层假设未在初始化路径兑现
- 旧测试把错误语义当作期望行为
  - 根因：测试未跟随设计语义更新

#### 临时解决方案

- 在 CLI 初始化时注册内置 `default` 合约
- 暂不接真实模型，用占位合约跑通执行链

#### 未解决事项

- 需要 `MockModelProvider` 替代当前更基础的占位执行

---

### 2. wf-ci 经验

#### 可复用成功做法

- 每次关键修改后执行 `go build` 和 `go test ./...`
- 使用 PowerShell 管道模拟交互式 shell
- 使用 `git diff --check` 做轻量格式检查

#### 暴露问题与根因

- Windows 下运行中的 `axis.exe` 会锁定二进制，导致构建失败
  - 根因：进程占用目标输出文件

#### 临时解决方案

- 提醒用户退出正在运行的 shell 后重新构建

#### 未解决事项

- 可考虑后续构建到临时文件名，如 `axis-dev.exe`，避免覆盖锁定

---

### 3. wf-doc-004 经验

#### 可复用成功做法

- 新功能先写 `requirements/design/tasks`
- 为 feature spec 增加 `workflow-binding.md`
- 将任务状态回写到 `tasks.md`
- 完成后同步 `HANDOVER.md`

#### 暴露问题与根因

- 一开始只用了 `docs/specs`，没有显式绑定 `workflow/`
  - 根因：缺少统一入口路由
- Meta-Workflow 文档过于宏大，实际执行规则不突出
  - 根因：长期愿景与当前项目机制混在一起

#### 临时解决方案

- 新增 `workflow-binding.md`
- 新增 `workflow/entry.md`
- 在 Meta-Workflow 顶部加入当前简化执行规则

#### 未解决事项

- 注册表字段仍偏重，后续可进一步裁剪最小字段集

---

### 4. wf-occams 经验

#### 可复用成功做法

- 不直接做 Web UI，先做 shell
- 不直接接真实模型，先做 MockModelProvider 规格
- 工作流修复优先修入口和一致性，而不是新增自动化平台

#### 暴露问题与根因

- 容易被“大而全”的工作流愿景诱导过度设计
  - 根因：Meta-Workflow 历史文档没有明确当前执行边界

#### 临时解决方案

- 在 `meta-workflow-management.md` 顶部声明长期设想暂不启用

#### 未解决事项

- 后续如新增 workflow，必须先证明现有工作流无法覆盖

---

### 5. wf-doc-006 经验

#### 可复用成功做法

- 新增小白指南降低理解门槛
- README 只放入口链接，细节放 docs
- HANDOVER 作为当前状态汇总
- 将设计思想单独成文，避免散落在聊天上下文中

#### 暴露问题与根因

- 文档容易滞后于代码实现
  - 根因：实现后没有强制或提醒式同步机制
- 索引文档和注册表曾长期不一致
  - 根因：索引不是自动生成，也没有作为权威入口使用

#### 临时解决方案

- 手动更新 README、HANDOVER、workflows/README
- PR Check 添加非阻塞文档上下文提醒

#### 未解决事项

- 文档索引是否自动生成仍未定；当前先保持手动同步

---

### 6. wf-monitoring 经验

#### 可复用成功做法

- 访问 GitHub event 嵌套字段时使用可选链
- 对 `workflow_run` 事件结构做条件保护
- 保持不同触发事件下脚本不崩溃

#### 暴露问题与根因

- schedule 与 workflow_run 的事件结构不同
  - 根因：GitHub Actions 上下文对象随事件类型变化

#### 临时解决方案

- 使用可选链和空值提前返回

#### 未解决事项

- 无

---

### 7. wf-cd 经验

#### 可复用成功做法

- 重复发布链路优先合并，不保留两个入口
- 废弃用 deprecated 标记而非完全抹除历史

#### 暴露问题与根因

- 删除 `release.yml` 后索引和 registry 状态没有完全同步
  - 根因：缺少状态传播检查

#### 临时解决方案

- 手动修正 `workflows/README.md` 和 registry note

#### 未解决事项

- 可在 registry validator 中增加“deprecated file missing is allowed but must have note”的轻量提示

---

## 三、经验评审与辩证扬弃

### 保留

1. **workflow/entry.md 作为第一入口**
   - 价值：减少 Agent 猜测
   - 标准化：所有任务先读入口路由

2. **Spec + workflow-binding 双轨机制**
   - 价值：把 feature spec 和上游 workflow 绑定
   - 标准化：新功能必须声明 workflow-binding

3. **非阻塞文档提醒**
   - 价值：补上下文但不制造控制
   - 标准化：经验类检查默认不阻塞

4. **bash is all you need**
   - 价值：优先 CLI / shell / 脚本验证
   - 标准化：交互层先 CLI，再 shell，再考虑 TUI/Web

5. **Mock first**
   - 价值：先跑通抽象边界，再接真实 Provider
   - 标准化：外部依赖集成前先有 MockProvider

### 修正

1. **Meta-Workflow 过度宏大**
   - 修正：顶部加入当前简化执行规则，长期设计降级为背景

2. **工作流索引不是权威入口**
   - 修正：索引只负责查找，入口路由改为 `workflow/entry.md`

3. **注册表字段过重**
   - 修正：当前先不大规模删字段，但后续维护只信核心字段

4. **代码变更必须更新文档的强制表述**
   - 修正：改为非阻塞提醒，硬门禁仅限构建/测试/安全

### 剔除

1. **不读 workflow 直接写 spec 或代码**
   - 原因：容易绕过项目工作流机制

2. **为单个 feature 新建独立工作流**
   - 原因：会导致工作流爆炸

3. **用 Web UI 解决当前交互问题**
   - 原因：违背 bash is all you need 与奥卡姆剃刀

4. **直接接真实模型**
   - 原因：API Key、网络、Provider 差异会过早污染抽象

### 沉淀

1. **入口规则**
   - 所有任务先读 `workflow/entry.md`

2. **新功能规则**
   - 必须有 `requirements/design/tasks/workflow-binding`

3. **验证规则**
   - Go 代码改动后至少运行 `go test ./...`
   - CLI 交互改动后用 PowerShell 管道模拟 shell

4. **文档规则**
   - 用户可见能力必须更新 README 或 BEGINNER_GUIDE
   - 项目状态必须更新 HANDOVER

5. **范围规则**
   - 未证明必要，不新增工作流
   - 未跑通 Mock，不接真实外部服务

---

## 四、对应工作流完善

### wf-doc-004：Meta-Workflow Management

**已完善**:

- 新增当前项目简化执行规则
- 明确 `workflow/entry.md` 是权威入口
- 明确长期平台级能力暂不启用
- 明确新功能必须有 `workflow-binding.md`

**后续建议**:

- 裁剪 registry 最小字段集
- 将旧的 `.github/workflows/registry.yml` 示例全部改为 `.github/config/registry.yml`

---

### wf-occams：Occam Workflow

**已完善**:

- 通过 workflow/entry 固化“不新增重型工作流”的规则
- 通过 Meta-Workflow 顶部规则限制平台级自动化扩张

**后续建议**:

- 增加一个“新增复杂度说明”小节：新增 UI、Provider、Workflow 前必须说明为什么 CLI/Mock/现有 workflow 不够

---

### wf-doc-006：Document Audit

**已完善**:

- README 增加小白指南入口
- HANDOVER 记录所有关键变更
- workflows/README 修正状态和路径

**后续建议**:

- 增加对 `.github/config/registry.yml` 路径引用的文档检查
- 增加对 deprecated workflow 仍出现在 active 区域的提醒

---

### wf-pr-check

**已完善**:

- 增加非阻塞文档上下文提醒

**后续建议**:

- CLI/shell 改动时建议输出本地验证命令
- 不要新增阻塞性经验检查

---

### wf-ci

**已完善**:

- 本轮遵循 gofmt/build/test/diff-check 验证

**后续建议**:

- Windows 本地构建可约定使用 `axis-dev.exe`，避免正在运行的 `axis.exe` 锁定

---

### wf-cd

**已完善**:

- `wf-release` 废弃状态和索引修正

**后续建议**:

- registry validator 对 deprecated workflow 的 missing file 应给出上下文提醒，而不是简单失败

---

### wf-monitoring

**已完善**:

- github-script 事件上下文访问方式已沉淀为可选链和事件条件保护

**后续建议**:

- 后续所有 github-script 均遵循 CODING_STANDARDS 的事件属性访问规范

---

## 五、当前遗留待办

1. **ModelProvider 实现**
   - 按 `docs/specs/model-provider/tasks.md` 实现 MockModelProvider

2. **Registry 字段瘦身**
   - 可后续单独做，不阻塞当前开发

3. **文档路径检查增强**
   - 后续可在 document-audit 中增加轻量提醒

4. **Windows 构建锁定规避**
   - 后续可在小白指南中建议开发构建用 `axis-dev.exe`

---

## 六、结论

今日最重要的沉淀是：

> 工作流机制必须从“文档知识库”变成“入口清晰、路由明确、可执行、不过度控制”的轻量执行系统。

当前已经完成关键修复：

- 有入口：`workflow/entry.md`
- 有简化规则：`workflow/meta-workflow-management.md` 顶部
- 有绑定机制：`workflow-binding.md`
- 有小白路径：`axis shell` + `default` 合约
- 有下一步：`MockModelProvider`

---

## 七、追加复盘：Milestone 2 规格绑定与 T1/T2/T2.5

本节补充 16:00 后围绕 Milestone 2 展开的工作，继续遵循“每个工作项只归入一个唯一上游工作流”的规则。

### 1. wf-doc-004：M2 规格与 workflow binding

**唯一上游工作流**: `wf-doc-004`

**工作项**:

1. 为 `docs/specs/milestone2/` 创建并确认 `workflow-binding.md`
2. 将 `requirements.md`、`design.md`、`tasks.md` 与 workflow binding 互相引用
3. 在 `tasks.md` 中补充 T0/T1/T2/T2.5 状态与验证结果
4. 将 `docs/current-progress.md` 与 M2 当前状态同步

**经验萃取**:

- **成功做法**: 用户指出 specs 不能脱离 workflow 后，先补 binding，再继续实现，避免 specs 与项目机制脱节
- **问题根因**: 一开始把 Spec-RDT 当成独立流程，没有先走 `workflow/entry.md` 的新功能路由
- **临时方案**: 增加 T0 专门确认 workflow binding，作为后续实现前置项
- **未解决事项**: `HANDOVER.md` 中 M2 状态仍需在下一次交接前同步到 T2.5 之后

### 2. wf-ci：本地 CI 等价验证与工具补齐

**唯一上游工作流**: `wf-ci`

**工作项**:

1. 运行 T1 baseline：`go test`、`gofmt`、`go vet`、race/coverage
2. 通过新增 orchestrator 测试将覆盖率从不足 60% 提升到 62.8%
3. 安装并运行 `staticcheck`、`gosec`、`govulncheck`、`markdownlint`
4. 在 T2/T2.5 后重复运行 CI 等价验证，覆盖率提升到 63.6% / 67.3%
5. 构建到临时文件名，避免覆盖未跟踪的 `axis.exe`

**经验萃取**:

- **成功做法**: 本地按 GitHub Actions 语义补齐工具链，先验证再继续开发
- **问题根因**: 本机缺少 CI 工具，导致早期只能部分验证；PowerShell 对 coverage flags 的解析也需要显式引号
- **临时方案**: Go 工具安装到 `$env:USERPROFILE\go\bin` 后用绝对路径运行；构建输出用临时 exe
- **未解决事项**: `markdownlint` 暴露大量既有风格问题，但 workflow 当前配置为非阻塞审计，应单独文档清理，不阻塞 M2

### 3. wf-pr-check：T2 scheduler ready-set 与 T2.5 CLI 语义修复

**唯一上游工作流**: `wf-pr-check`

**工作项**:

1. 为 scheduler 增加 `GetReadyTasks(limit int)`
2. 保持 `GetNextTask()` 向后兼容并委托 `GetReadyTasks(1)`
3. 添加 FIFO、limit、dependency blocking/unblocking 测试
4. 审查普通 CLI 是否符合 `Bash is All You Need`
5. 新增 `cmd/axis/main_test.go`
6. 修正 `axis run/status` 不再误导用户运行跨进程无效的 `axis start`

**经验萃取**:

- **成功做法**: 对设计哲学违背项先写测试复现，再做最小修复
- **问题根因**: CLI 全局 orchestrator 是进程内状态，但错误提示暗示 `axis start` 可跨进程提供状态
- **临时方案**: T2.5 只修正普通 CLI 的进程内语义，不引入持久化状态或 daemon
- **未解决事项**: 若未来要让 `axis run` 与 `axis status` 跨进程闭环，需要文件状态或 daemon 设计；这不属于当前 T2.5

### 4. wf-occams：范围收敛与复杂度控制

**唯一上游工作流**: `wf-occams`

**工作项**:

1. T2 只增加 ready-set API，不重写 scheduler 为完整图引擎
2. T2.5 只修正 CLI 语义，不引入数据库、daemon、Web UI 或 TUI
3. 将 `markdownlint` 问题判断为非阻塞审计，不扩大为大规模格式化任务

**经验萃取**:

- **成功做法**: 发现问题后优先做局部修正，而不是借机扩张架构
- **问题根因**: “普通 CLI 跨进程闭环”容易诱导出 daemon 或持久化设计，但当前 milestone 未要求
- **临时方案**: 明确 local process state，用更准确的错误上下文替代误导提示
- **未解决事项**: 未来跨进程状态需要进入独立 spec，不应混入当前 T3

### 5. wf-doc-006：进度与报告同步

**唯一上游工作流**: `wf-doc-006`

**工作项**:

1. 更新 `docs/current-progress.md` 记录 T1/T2/T2.5 与 CI 工具验证状态
2. 在 `tasks.md` 中记录每个任务的 Current Result
3. 将 markdownlint 既有问题记录为非阻塞文档清理事项

**经验萃取**:

- **成功做法**: 每完成一个原子阶段就回写进度，避免长会话中断后状态丢失
- **问题根因**: `HANDOVER.md` 与 `current-progress.md` 容易出现更新时间差
- **临时方案**: 以 `docs/current-progress.md` 和 `docs/specs/milestone2/tasks.md` 作为当前阶段即时状态源
- **未解决事项**: 交接前必须统一更新 `HANDOVER.md` 和 `AGENT_INSTRUCTIONS.md`

---

## 八、追加经验评审与辩证扬弃

### 保留

1. **实现前补齐 workflow binding**
   - 新功能不能只靠 specs，必须明确上游 workflow、执行顺序、非目标和完成标准。

2. **CI 工具缺失要显式安装或标明未验证**
   - 不能把未安装工具的检查视为通过。

3. **设计哲学违背项可插入小型校正任务**
   - T2.5 证明了在 T2 与 T3 之间插入小修正，比带着错误继续开发更稳。

4. **本地构建使用临时输出名**
   - Windows 下避免覆盖或锁定未跟踪 `axis.exe`。

### 修正

1. **`axis start` 语义**
   - 修正为不暗示跨进程状态；普通 CLI 只保证当前进程内初始化。

2. **文档审计强度**
   - `markdownlint` 作为非阻塞审计项，不因既有风格问题阻塞 M2 功能开发。

3. **覆盖率门禁处理**
   - 覆盖率不足时先补有业务意义的测试，不用空测试或无效覆盖率技巧。

### 剔除

1. **直接进入 T3 而忽略 CLI 哲学违背**
   - 会让 `Bash is All You Need` 变成口号，而不是可执行标准。

2. **为跨进程 CLI 闭环临时引入 daemon/database**
   - 当前不符合 M2 最小范围。

3. **把 markdownlint 既有问题作为当前开发硬门禁**
   - 这会把非阻塞上下文提醒升级为过度控制。

### 沉淀

1. **新功能任务必须先有 workflow-binding**
2. **Go 代码改动后至少运行：gofmt、go vet、go test；关键任务运行 race/coverage**
3. **CI 工具不可用时必须记录“未验证”，不可写通过**
4. **CLI 交互改动必须检查普通命令与 shell 两条路径**
5. **本地 Windows 构建默认用临时二进制名，除非明确要覆盖 release artifact**

---

## 九、追加工作流完善结论

### wf-doc-004

需要固化：

- 工作复盘必须使用唯一上游工作流归类，辅助 workflow 只能作为验证项
- 新功能规格缺少 `workflow-binding.md` 时，先补 binding，再实现

### wf-ci

需要固化：

- 本地 CI 等价验证应包含工具可用性检查
- Windows 本地构建不得默认覆盖已有 `axis.exe`
- coverage flags 在 PowerShell 中应使用显式引号

### wf-pr-check

需要固化：

- CLI/shell 行为变更必须包含用户可见语义测试
- 设计哲学违背项允许插入小型校正任务，但必须写入 `tasks.md`

### wf-occams

需要固化：

- 发现跨进程、持久化、daemon、UI 需求时，默认先做最小语义澄清，除非当前 spec 明确要求扩展架构

### wf-doc-006

需要固化：

- `docs/current-progress.md` 是即时状态源
- `HANDOVER.md` 是交接状态源，交接前必须同步

---

## 十、追加复盘：设计主权交接与自因化入口文档重写

本节补充 18:00 后围绕 Axis 新设计思想展开的工作。归类继续遵守唯一上游 workflow 原则。

### 1. wf-doc-004：设计主权交接与 Autogenesis 原则固化

**唯一上游工作流**: `wf-doc-004`

**工作项**:

1. 将“设计主权交给外部 Agent”记录为新的设计事实
2. 明确 Axis 自举起点是 `External thought injection`
3. 将 `workflow / contract / permission rule / spec` 定义为过渡性结构
4. 创建 `reports/axis-autogenesis-design-2026-05-08.md`
5. 重写 `reports/bootstrap-gap-analysis-2026-05-08.md` 为自因化差距分析

**经验萃取**:

- **成功做法**: 先把哲学命题转化为架构原则，再决定工程路线，避免直接把“自举”误实现成代码自修改
- **问题根因**: 旧报告偏工程 checklist，不能充分表达 Axis 从工具项目向自因系统过渡的设计主线
- **临时方案**: 新增 Autogenesis 设计报告作为中间层，承接哲学与工程 specs
- **未解决事项**: 后续需要正式创建 `bootstrap-loop` 与 `autogenesis-loop` specs，但不能早于 M2 T3-T7

### 2. wf-doc-006：入口类文档重写

**唯一上游工作流**: `wf-doc-006`

**工作项**:

1. 重写根 `README.md`
2. 重写 `docs/README.md`
3. 重写 `docs/QUICKSTART.md`
4. 重写 `docs/WHITEPAPER.md`
5. 同步 `docs/current-progress.md`
6. 同步 `HANDOVER.md`

**经验萃取**:

- **成功做法**: 入口文档先统一“Axis 是什么”，再给使用命令和下一步，降低概念分裂
- **问题根因**: 旧入口文档仍偏“任务调度平台”叙述，与新确立的自因化方向不一致
- **临时方案**: 优先重写入口/索引/白皮书/进度/交接，而不是全量批量改所有文档
- **未解决事项**: `docs/architecture/core-modules.md`、`docs/ROADMAP.md`、`docs/DIAGRAMS.md` 后续仍需按新定位二次审查

### 3. wf-occams：防止自因化叙事诱发过度实现

**唯一上游工作流**: `wf-occams`

**工作项**:

1. 明确 M2 仍只做 Autogenesis Loop 的执行底座
2. 明确不直接跳到真实 LLM SDK、Web UI、外部数据库或 daemon
3. 将 bootstrap-loop / autogenesis-loop 保留为后续 specs

**经验萃取**:

- **成功做法**: 把宏大哲学转化为阶段边界，先完成 T3-T7
- **问题根因**: “自举 / 自因化”容易诱导过早实现真实 Agent runtime 或复杂控制面
- **临时方案**: 在 README、QUICKSTART、current-progress 中显式写出“不要做什么”
- **未解决事项**: 后续进入 bootstrap-loop 时仍需再次确认不引入重型依赖

---

## 十一、追加经验评审与辩证扬弃

### 保留

1. **设计主权交接必须落盘**
   - 设计主权变化不是聊天上下文，应写入设计哲学、报告和交接文档。

2. **入口文档优先统一本体定位**
   - README/QUICKSTART/WHITEPAPER 必须先回答 Axis 是什么，再回答如何使用。

3. **Autogenesis 与 M2 解耦**
   - M2 是执行底座，不是完整自举实现。

4. **Scaffold-to-Self 作为文档解释原则**
   - 任何 workflow、contract、permission、spec 都应说明其过渡性，不伪装成终局控制。

### 修正

1. **不要把选择权再抛回用户**
   - 在设计主权已交接后，Agent 应承担设计与文档组织决策，只在高风险破坏性操作前请求确认。

2. **不要用删除重建处理入口文档**
   - 已有文档应安全覆盖内容或 patch，不应删除文件。

3. **不要把自举报告写成纯工程 checklist**
   - 工程缺口必须放在自因化主线下解释。

### 剔除

1. **继续把 Axis 叫普通任务调度平台**
   - 该叙述已不能承载当前设计方向。

2. **一有新哲学就直接编码**
   - 哲学应先固化为架构原则、报告、spec，再进入实现。

3. **借 Autogenesis 提前接真实 LLM SDK**
   - 违背 Mock-first、Bash-first 与 Occam 原则。

### 沉淀

1. **设计主权规则**
   - 用户明确交接设计主权后，Agent 应主动决策文档组织、设计路线与复盘落盘。

2. **入口文档规则**
   - README、docs/README、QUICKSTART、WHITEPAPER、current-progress、HANDOVER 必须保持项目本体定位一致。

3. **自因化范围规则**
   - Autogenesis 是长期方向；当前阶段只实现 M2 执行底座。

4. **非破坏性编辑规则**
   - 重写已有文档时不删除文件；使用安全覆盖或 patch。

---

## 十二、追加工作流完善结论

### wf-doc-004

需要固化：

- 当项目核心设计主权或本体定位发生变化，必须先更新设计哲学文档与架构报告，再更新入口文档
- 哲学命题必须转化为可执行的工程边界：当前做什么、不做什么、后续 spec 是什么

### wf-doc-006

需要固化：

- 入口类文档必须成组更新：`README.md`、`docs/README.md`、`docs/QUICKSTART.md`、`docs/WHITEPAPER.md`、`docs/current-progress.md`、`HANDOVER.md`
- 入口文档一致性校验至少检索核心概念：`自因化`、`Autogenesis Loop`、`Competence earns autonomy`、`Scaffold-to-Self`

### wf-occams

需要固化：

- 宏大设计思想只能改变路线与边界，不能自动扩大当前 milestone scope
- 进入真实 Agent runtime、外部数据库、daemon、Web UI 前必须先有独立 spec 与复杂度说明

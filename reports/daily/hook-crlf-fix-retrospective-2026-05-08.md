# Claude Code Hooks CRLF 缺陷修复复盘

**日期**: 2026-05-08
**范围**: 4 个 Claude Code hooks 的缺陷诊断、修复与验证
**基准**: `.github/config/registry.yml` categories + `workflow/entry.md` 路由规则
**原则**: More Context, More Action, Zero Control, Bash is All You Need, 奥卡姆剃刀

---

## 一、工作内容分类

### 1. wf-pr-check：auto gofmt Hook CRLF 修复

**唯一上游工作流**: `wf-pr-check`

**工作项**:

1. 发现 PostToolUse(Write|Edit) gofmt hook 在 Windows 下静默失效
2. 定位根因：stdin 管道携带 CRLF，`\r` 附加到文件路径末尾，`*.go` glob 匹配失败
3. 修复：`jq ... | tr -d '\r' | while ...` 在 jq 输出与 while 循环之间插入 `tr -d '\r'`
4. 用 `printf '...\r\n' | bash` 模拟 Windows stdin 验证修复前后行为差异

---

### 2. wf-ci：CI 通知 Hook 三项缺陷修复

**唯一上游工作流**: `wf-ci`

**工作项**:

1. 硬编码路径 `"/c/Program Files/GitHub CLI/gh"` → `gh`（PATH 查找）
2. `gh run list --limit 5` 无分支过滤 → 加 `--branch "$(git branch --show-current)"`
3. `sleep 20` 偏短 → `sleep 25`
4. stdin 消费端无 CRLF 防御 → 加 `cat | tr -d '\r'`
5. 通知消息维度不全 → 输出中追加当前分支名 `"CI: $branch | $summary"`

**辅助验证工作流**: `wf-doc-006`

---

### 3. wf-security：安全门禁 Hook 验证

**唯一上游工作流**: `wf-security`

**工作项**:

1. 验证 `read -r` + `jq -r` 提取命令在 CRLF 输入下正常工作
2. 确认 gosec 和 staticcheck 当前代码库通过（0 issues）
3. 判断阻塞策略合理：当前代码干净，未来引入 issue 时应拦截

---

### 4. wf-doc-006：文档同步脚本 CRLF 防御

**唯一上游工作流**: `wf-doc-006`

**工作项**:

1. 发现 `scripts/doc-sync-check.sh` 内 `input=$(cat)` 无 CRLF 防御
2. 修复：`input=$(cat | tr -d '\r')`
3. 用 `printf '...\r\n' | bash scripts/doc-sync-check.sh` 验证修复

---

## 二、分类经验萃取

### 1. wf-pr-check 经验

#### 可复用成功做法

- `printf '...\r\n' | bash` 在本地精确模拟 Claude Code hook 的 Windows stdin 环境
- 非破坏性验证：先 mess up Go 文件，等 3 秒，用 `gofmt -d` 检查是否有 diff
- 管道逐段测试：先测 jq 提取，再测 case 匹配，最后测 gofmt 写入

#### 暴露问题与根因

- Windows bash (MSYS2/Git Bash) 下 stdin 管道携带 CRLF 换行
  - 根因：Windows 默认文本模式换行为 `\r\n`，bash 的 `read` 与 `while read` 将 `\r` 视为文件名的一部分
- `2>/dev/null` 同时抑制了 gofmt 的合法错误（语法错误导致格式化失败）
  - 根因：追求 hook 静默非阻塞，但过度压制了有用信号
- async hook 的失败不可见，不检查不会发现
  - 根因：async hook 无错误回传通道

#### 临时解决方案

- 在所有 stdin → 管道消费的衔接点插入 `tr -d '\r'`

#### 未解决事项

- 是否需要为 async hook 增加错误日志或通知机制
- 其他 Windows 用户是否也会遇到同样问题（已知，但未提 issue）

---

### 2. wf-ci 经验

#### 可复用成功做法

- Hook 命令中的可执行文件优先依赖 PATH，不硬编码绝对路径
- CI 状态查询必须加分支过滤，否则结果无意义
- 通知消息包含足够上下文（分支名），降低理解成本

#### 暴露问题与根因

- 硬编码路径来源于 `which gh` 即时输出，未考虑可移植性
  - 根因：hook 配置像一次性脚本，缺乏"被其他人/环境使用"的意识
- `sleep 20` 后 CI 可能尚未排队
  - 根因：GitHub Actions 冷启动时间不可预测（10s-2min），固定 sleep 不可靠
- `gh run list --limit 5` 返回全局最近 5 个 run，不保证包含当前 push 的 CI
  - 根因：没有用 `--branch` 或 `--commit` 过滤

#### 临时解决方案

- 使用 `git branch --show-current` 获取当前分支并传入 `--branch`
- sleep 从 20 加到 25（仍是临时妥协）

#### 未解决事项

- 25 秒仍不能保证 CI 已启动；更可靠的方案需轮询或 webhook，但超出当前 hook 机制能力
- `gh run list` 不支持 `--commit` flag，无法精确匹配某次 push

---

### 3. wf-security 经验

#### 可复用成功做法

- gosec + staticcheck 作为 pre-push 阻塞门禁，CI 中已有等价检查，push 前再跑是冗余保护
- `read -r` + `jq -r` 的 stdin 解析方式天然免疫 CRLF 问题（jq 容忍 JSON 尾部空白）

#### 暴露问题与根因

- 无。该 hook 设计合理，当前代码库通过所有检查。

#### 临时解决方案

- 无。

#### 未解决事项

- gosec 规则集未调优；若未来代码触发 G104（未检查 error）等高频规则，所有 push 会被拦。可在发现首次误拦时调整规则集。

---

### 4. wf-doc-006 经验

#### 可复用成功做法

- 非阻塞提醒式 hook 设计：永远返回 `continue:true`，只给 systemMessage
- shell 脚本内用 jq 解析 stdin JSON，不依赖复杂字符串匹配

#### 暴露问题与根因

- 脚本内 `input=$(cat)` 同样受 CRLF 污染，虽然 jq 容错，但若后续用 `read` 或 `case` 处理文件名就会重现 Hook 1 的 bug
  - 根因：同一类型的防御需求在不同 hook 之间没有共享

#### 临时解决方案

- 统一在 stdin 捕获点加 `tr -d '\r'`

#### 未解决事项

- 是否应该将 CRLF 防御模式写入项目编码规范，使后续新增 hook/脚本自动遵循

---

## 三、经验评审与辩证扬弃

### 保留

1. **`tr -d '\r'` 作为 Windows hook stdin 标准防御**
   - 价值：一行修复，零副作用，覆盖所有 stdin → 文本处理场景
   - 标准化：所有从 stdin 消费数据的 hook 命令必须加此防御

2. **逐段测试法**
   - 价值：把复杂管道拆成 jq 提取 → case 匹配 → 副作用执行三段独立测试
   - 标准化：hook 调试时先用 `printf '...\r\n'` 模拟 stdin，逐段验证

3. **`gofmt -d` 作为非侵入验证**
   - 价值：不改写文件就能判断 hook 是否触发，不干扰 hook 自身行为

4. **PATH over 硬编码路径**
   - 价值：hook 命令中使用 `gh` 而非 `/c/Program Files/GitHub CLI/gh`
   - 标准化：所有 hook 命令中的外部工具使用 PATH 查找

### 修正

1. **async hook 的 `2>/dev/null` 过度使用**
   - 修正：至少保留 gofmt 的 stderr，以便发现语法错误导致的格式化失败
   - 不对每个 hook 强行去 `2>/dev/null`，但应评估是否有用的错误信息被丢弃

2. **CI hook 的 sleep 策略**
   - 修正：从 magic number 20s 改为 25s + 分支过滤，但仍然不完美
   - 长期方向：接受 async hook 能力边界，不过度工程化

3. **硬编码路径来源于即时 `which` 输出**
   - 修正：写 hook 配置时即使用 PATH 查找形式，不把本机绝对路径写入共享配置

### 剔除

1. **假设 hook 能在所有平台正常工作而不验证**
   - 原因：Windows CRLF 与 Unix LF 的差异是已知问题，首次部署就应验证

2. **把 hook 当一次性脚本写**
   - 原因：settings.json 是版本控制的共享配置，应考虑跨环境可移植性

### 沉淀

1. **Windows 兼容性规则**
   - 所有 Claude Code hook 命令中，从 stdin 读取数据的 shell 管道，必须在消费端加 `tr -d '\r'`
   - 两个关键消费模式：`input=$(cat | tr -d '\r')` 和 `jq ... | tr -d '\r' | while read`

2. **Hook 验证规则**
   - 新增或修改 hook 后，必须用 `printf '...\r\n'` 模拟 Windows stdin 进行端到端验证
   - async hook 用 `sleep +` 工具检查副作用验证（如 `gofmt -d`）

3. **Hook 可移植性规则**
   - 外部工具使用 PATH 查找，不硬编码绝对路径
   - 不使用仅存在于特定环境的工具或路径

4. **Hook 错误处理规则**
   - `2>/dev/null` 只用于已知的、非关键的错误输出
   - 可能包含诊断价值的错误（如 gofmt 语法错误）应保留 stderr

---

## 四、对应工作流完善

### wf-pr-check

**需固化**:

- Claude Code hook 配置变更视为质量基础设施变更，适用 wf-pr-check 流程
- Hook 修复后必须包含：缺陷复现命令 → 根因说明 → 修复 → Windows CRLF 验证

**后续建议**:

- 在 PR check workflow 中增加 hook 配置语法校验（JSON schema validation）
- 考虑增加 Windows 环境下的 hook smoke test

---

### wf-ci

**需固化**:

- CI 相关 hook（如 push 后状态通知）应使用分支过滤确保结果相关性
- 外部 CLI 工具（gh, jq, gofmt）在 hook 中使用 PATH 查找，不硬编码路径

**后续建议**:

- `gh run list` 冷启动窗口期不可靠问题暂无完美方案，保持当前 `--branch` + sleep 作为最小可行方案

---

### wf-security

**需固化**:

- pre-push 安全检查 hook 当前状态良好，无需修改
- 若后续 gosec 出现误拦，应先调整规则集再改 hook 逻辑

**后续建议**:

- 可在 CI security workflow 中记录当前 gosec 规则集配置，作为 hook 的参照基准

---

### wf-doc-006

**需固化**:

- `scripts/doc-sync-check.sh` 已加固，后续新增脚本默认遵循 CRLF 防御模式
- 所有从 hook stdin 读取数据的 shell 脚本，第一行数据处理必须包含 `tr -d '\r'`

**后续建议**:

- 将 CRLF 防御模式写入 `CLAUDE.md` 的 hook 开发注意事项
- 在 document-audit workflow 中增加对 scripts/ 目录的语法检查（bash -n）

---

## 五、当前遗留待办

1. **async hook 错误可见性**
   - 当前 async hook 失败静默，无日志/通知通道。属于 Claude Code 平台能力，不在本项目范围内。

2. **CI hook 冷启动窗口期**
   - sleep + poll 模式不可靠，但 webhook 方案超出当前架构边界。暂时接受。

3. **gosec 规则集调优**
   - 当前代码干净，暂无问题。首次出现误拦时处理。

4. **CRLF 防御模式写入编码规范**
   - 建议在 `CLAUDE.md` 中增加一节 "Hook 开发注意事项"。

---

## 六、结论

今日最重要的发现是：

> Windows bash 环境下的 CRLF stdin 是一个系统性风险——所有 Claude Code hook 命令中，凡是 stdin → shell 文本处理 的链路，`\r` 都会污染数据，导致 glob 匹配、字符串比较、文件名操作静默失败。

修复本身轻量（4 处 `tr -d '\r'`），但诊断过程暴露了更重要的工程问题：

- async hook 失败完全不可见
- hook 配置缺乏跨环境可移植性意识
- 缺少 hook 本地验证的标准方法

三项已通过本次复盘沉淀为可执行规则。

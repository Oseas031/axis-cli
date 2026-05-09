# 每日复盘：2026-05-09

## 一、工作内容分类

基于现有工作流体系，今日 7 个 commits 按唯一上游工作流归类：

### 类别 1：功能实现 / Bug 修复 → `wf-pr-check + wf-ci + wf-doc-006`

| 工作项 | Commits | 说明 |
|---|---|---|
| M3 Phase 1 覆盖率提升至 88.8% | `3a9da92`, `a73ef20`, `44e4f7c`, `85f9877` | 新增 24+ 测试，覆盖 cmd/axis shell 模拟、orchestrator 重试耗尽、dispatcher 超时/errChan、executor Provider 路径、admission SLA 边界 |
| M3 Phase 1 ModelProvider + DAG/SLA 补全 | `a2ea1e2` | ModelProvider 接口+Mock、`ErrDependencyNotReady`、`sla.failure_class`、失败依赖处理 |

### 类别 2：基础设施诊断与修复 → `wf-ci + wf-doc-006`

| 工作项 | 说明 |
|---|---|
| Git push 连通性诊断 | 定位 git.exe HTTPS 无法连接 GitHub 根因：全局 `http.version=HTTP/1.1` 在 Windows 上被阻断，切换为 HTTP/2 修复 |
| Git 全局配置审计 | 发现并修复 3 项不合理配置（postbuffer 500MB→50MB、safe.directory=*→精确路径、补 init.defaultBranch=main） |
| CI 监控 | 推送后 6 个 workflow 全部通过（CI、PR Quality Check、Dev、Document Audit、Security、Registry Validator） |

### 类别 3：文档 / 设计 / 工作流调整 → `wf-doc-004 + wf-doc-006 + wf-occams`

| 工作项 | Commit | 说明 |
|---|---|---|
| 进度文档对齐 | `d2cf3a8` | 更新 CLAUDE.md、HANDOVER.md、AGENT_INSTRUCTIONS.md、current-progress.md |

### 类别 4：工具机制调查 → `wf-doc-004 + wf-occams`

| 工作项 | 说明 |
|---|---|
| Worktree 隔离机制缺陷调查 | 实测验证 EnterWorktree 基于默认分支 main HEAD 而非当前分支 HEAD，确认平台层面缺陷；建立手动 worktree 方案 B 作为可靠替代 |

---

## 二、分类经验萃取

### 2.1 功能实现类

**可复用的成功做法：**

1. **输出验证失败触发重试测试**：利用 MockModelProvider 不提供特定输出字段的特点，创建 `OutputSchema` 要求 `missing_field` 的 contract，使 `ValidateOutput` 自然失败 → dispatcher 返回 error → orchestrator 进入重试/耗尽路径。无需 mock dispatcher，纯白盒路径触发。
2. **Shell stdin 模拟模式**：`os.Pipe()` 替换 `os.Stdin`，goroutine 写入命令后 close 触发 EOF，同时 pipe 捕获 stdout。可复用于任何基于 `bufio.Scanner` 的交互式 CLI 测试。
3. **Windows 跨平台信号替代**：`syscall.Kill` 在 Windows 不可用，改用 `os.FindProcess(os.Getpid()).Signal(os.Interrupt)`。但 Windows 上程序化信号不触发 `signal.Notify`，最终确认该路径在 Windows 不可测试，移除了相关测试。

**暴露的问题与根因：**

1. **Windows 程序化信号不触发 signal.Notify**：Go 在 Windows 上 `signal.Notify` 仅响应控制台 Ctrl+C 事件，`Process.Signal()` 发送的信号不走同一通道。根因：Windows 信号模型与 Unix 不同。导致 `startOrchestrator` 函数（依赖 SIGINT 退出）无法在 Windows 测试。
2. **Provider=nil 时 Execute 不经过输出验证**：`ContractExecutorImpl.Execute` 在 `p == nil` 时直接返回 `{"status": "validated"}` 桩结果，跳过 `ValidateOutput`。导致 dispatcher 的 errChan 路径在未设置 provider 时无法触发。修复：在测试中显式调用 `contractExec.SetProvider(provider.NewMockModelProvider())`。

**临时解决方案：**

- Windows 信号测试：移除而非修复，标记为已知限制。（长期可考虑重构 `startOrchestrator` 使其接受可注入的 context 或 shutdown channel）

**未解决的阻塞点：**

- `startOrchestrator` 和 `main()` 函数因 `os.Exit(1)` / 信号阻塞而覆盖率极低（cmd/axis 68% 的剩余 32% 主要在此）。需要架构调整（依赖注入 context）才能进一步测试。

### 2.2 基础设施类

**可复用的成功做法：**

1. **HTTP/2 修复 git 连接**：Windows 上 git + HTTP/1.1 连接 GitHub 持续超时，切换 HTTP/2 立即恢复。可固化为 Windows 开发环境初始化步骤。
2. **`gh` CLI 与 git 分离诊断**：当 git push 失败但 `gh` CLI 正常时，问题在传输层（TCP/HTTP）而非认证层。诊断路径：`ping` → `curl` → `gh auth status` → `git ls-remote`，分层定位。

**暴露的问题与根因：**

1. **`http.postbuffer=524288000`（500MB）**：每次 push 预分配 500MB 内存。疑似从 StackOverflow 复制粘贴，从未遇到实际需要此值的大文件场景。默认 1MB 已足够绝大多数仓库。已降至 50MB。
2. **`safe.directory=*`**：全局关闭 git 目录所有权检查。在单用户 Windows 机器风险低，但在多用户环境或容器中可能被利用。已替换为精确路径。
3. **缺少 `init.defaultBranch`**：导致 `git init` 创建的仓库默认分支取决于系统（通常 `master`），与项目使用的 `main` 不一致。

### 2.3 文档类

**可复用的成功做法：**

1. **四处文档同步更新**：每次进度变更同步更新 CLAUDE.md + HANDOVER.md + AGENT_INSTRUCTIONS.md + current-progress.md。避免了单一文档过时导致的上下文断裂。
2. **CLAUDE.md 作为 Agent 第一入口**：包含项目定位、构建命令、架构图、运行时流程、约束和缺陷。使新 Agent 实例可快速恢复上下文。

**暴露的问题与根因：**

- CLAUDE.md 中 M2 状态仍标注 "In progress" 和 "T3 is next pending"，已过时。根因：M2 完成后未及时更新 CLAUDE.md。

### 2.4 工具机制调查类

**可复用的成功做法：**

1. **EnterWorktree 行为实测验证**：创建测试 worktree → 检查 HEAD → 对比当前分支 HEAD → 确认偏差。一次性实证而非猜测。
2. **方案 B 工作流**：`git worktree add -b <name> <path> <commit>` + `EnterWorktree --path`。避免了 Agent `isolation: "worktree"` 的平台缺陷。

**暴露的问题与根因：**

1. **EnterWorktree 基于默认分支 main HEAD**：工具描述写 "based on HEAD"，但实现用的是仓库默认分支 HEAD（`main`），不是当前会话分支 HEAD。导致 feature 分支上创建的 worktree 落后数十个 commits。根因在 Claude Code 平台层，本仓库无法修复。
2. **同一分支不能同时在两个 worktree checkout**：`git worktree add` 拒绝 checkout 已在其他 worktree 使用的分支。需要基于 commit SHA 创建新分支。

---

## 三、经验评审与辩证扬弃

### 保留

| 经验 | 理由 |
|---|---|
| 输出验证失败 → 重试测试模式 | 纯白盒，无需 mock 注入，利用现有架构即可触发深层路径。可标准化 |
| Shell stdin pipe 模拟模式 | 通用、简洁、无外部依赖，可复用于所有交互式 CLI 测试 |
| HTTP/2 作为 Windows git 默认 | 已验证解决连接问题，应固化为环境初始化规范 |
| 4 文档同步更新 | 防止上下文断裂，保持为强制要求 |
| 手动 worktree（方案 B） | 已验证可靠，在平台修复前作为并行开发标准方案 |
| 分层网络诊断路径 | ping → curl → gh → git，高效定位问题层 |

### 修正

| 经验 | 问题 | 修正方向 |
|---|---|---|
| Windows 信号测试 | 程序化信号不触发 `signal.Notify` | 修正为：Windows 上不测试信号依赖路径；长期重构 `startOrchestrator` 接受可注入 cancel channel |
| `git config --global` 直接修改 | 修改前未备份原始值 | 修正为：修改前先 `git config --list --global > backup` |
| 覆盖率提升以测试数量为目标 | 容易写低价值测试 | 修正为：以未覆盖分支路径为目标，用 `go tool cover -func` 驱动测试设计 |

### 剔除

| 经验 | 理由 |
|---|---|
| Agent `isolation: "worktree"` 并行开发 | 平台缺陷导致基于旧 HEAD，并行结果不可靠，已用方案 B 替代 |
| `syscall.Kill` 跨平台信号 | Windows 完全不可用，不可移植，剔除 |
| `http.version=HTTP/1.1` | Windows 上连接 GitHub 失败，已永久切换为 HTTP/2 |
| `safe.directory=*` | 安全实践不当，已替换为精确路径 |
| `http.postbuffer=524288000` | 无实际需求，过度分配内存，已降至合理值 |

### 沉淀为规范

| 规范 | 融入工作流 |
|---|---|
| Windows 开发环境 `git config --global http.version HTTP/2` | wf-dev 环境初始化步骤 |
| 并行开发使用手动 worktree（方案 B），不使用 Agent isolation worktree | wf-doc-004 并行开发规则 |
| 覆盖率测试设计先跑 `go tool cover -func` 定位未覆盖分支 | wf-pr-check 测试规范 |
| CLI 文档更新范围：CLAUDE.md + HANDOVER.md + AGENT_INSTRUCTIONS.md + current-progress.md | wf-doc-006 文档同步清单 |
| git 全局配置最小集：user.name, user.email, http.version=HTTP/2, init.defaultbranch=main | wf-dev 环境规范 |

---

## 四、对应工作流完善

### 4.1 wf-dev（Development Workflow）— 补充环境初始化

当前 `dev-workflow.yml` 缺少 Windows 环境初始化步骤。建议在文档中补充：

```text
## Windows 环境初始化
git config --global http.version HTTP/2
git config --global init.defaultbranch main
```

### 4.2 wf-doc-004（Meta-Workflow）— 新增并行开发规则

在 `meta-workflow-management.md` 中新增：

```text
7. **并行开发隔离**：Agent isolation worktree 存在已知缺陷（基于 main HEAD）。
   并行开发时使用手动 worktree：
   git worktree add -b <branch> .claude/worktrees/<name> <commit>
   完成后 git worktree remove --force + git branch -D 回收。
```

### 4.3 wf-pr-check（PR Quality Check）— 补充测试设计规范

在 PR Check 的测试覆盖要求中补充：

```text
测试设计先于编码：
1. 运行 go test -coverprofile=cov.out ./<package>/...
2. go tool cover -func=cov.out | grep -v "100.0%"
3. 针对 <100% 的函数设计测试用例
4. 优先覆盖错误处理、边界条件、并发路径
```

### 4.4 wf-doc-006（Document Audit）— 强化文档同步清单

`document-audit.yml` 的检查项中，将进度文档同步范围明确为固定清单：

```text
进度更新必须同步以下 4 个文件：
- CLAUDE.md（Current Status + Architecture + Runtime Flow）
- HANDOVER.md（交接状态 + 覆盖率 + 已知问题 + 下一步行动）
- AGENT_INSTRUCTIONS.md（当前状态摘要）
- docs/current-progress.md（已完成/进行中/待处理 + 最近提交）
```

### 4.5 wf-occams — 补充架构约束

新增可测试性约束：

```text
### 可测试性设计约束

新增 CLI 命令或后台函数时：
- 阻塞在信号/全局状态的函数应接受可注入的 context 或 channel
- 避免直接在函数内调用 os.Exit()
- 避免依赖 syscall.Kill 等不可移植 API
- 若必须在当前平台不可测试，在测试文件中明确标记原因
```

---

## 本次复盘总结

- **唯一上游工作流归类**：4 个类别，7 个 commits 全量归集
- **保留** 6 项可复用经验
- **修正** 3 项做法
- **剔除** 5 项不合理/不可靠做法
- **沉淀** 5 条规范到 5 个对应工作流
- **未解决阻塞**：`startOrchestrator` + `main()` 覆盖率（需架构调整）；worktree 平台缺陷（非本仓库可控）

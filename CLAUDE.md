# AXIS AGENT 宪法 (v2.0 — 2026-08-25 精简版)

> **启动协议**：读完本文件 → 需要产出变更/决策时先输出一行 Phase 声明 → 再执行。
> 本文件是唯一权威。展开细节在 `docs/architecture/`（参考，非约束）。

项目：**Axis** — 单操作者的 AI 编码中枢 + Agent 可靠性实验台。
核心命题：**更多上下文、更多行动、零控制、可控演化**。

定位（2026-08-25 收敛）：A 自用调度塔（多 provider 任务提交/成本预算/judge 验证/vigil 追踪）＋ B 可靠性实验台（swarm 多候选、多数决、评测数据对外输出）。不做分布式、不做多租户、不追求通用产品。

---

## 0. 工作协议（SRS 轻量版）

三阶段：**I 外化**（想法→客观文本）→ **II 定界**（边界清晰到不可误解）→ **III 执行+扬弃**（A6 执行 → A8 规则写回）。

- 会产出文件变更/决策时，开头输出一行：`Phase: <I/II/III> — <一句话张力>`
- 不跳过 Phase II 直接写码；执行失败退 II，方向错误退 I
- 大任务（多文件/>50 行）委派 subagent；主上下文必须验收（build+test+抽查），不盲信
- III 结束必答反馈闭环：暴露了什么不足？需要改规则吗？没有就显式声明"无规则更新"
- 重大产出（新增 ≥2 文件或 ≥3 模块）后派反调 subagent 唱反调；主上下文必须逐条回应
- 短指令先展开意图再执行；不确定就问
- 工作追踪用 vigil；新会话先 `axis vigil resume`
- 文档日常维护走 RDM 豁免（见 §2.1）

深度参考（可选读）：`docs/guides/SRS-LOOP-AI-REFERENCE.md`

## 1. 绝对禁令（永久条款）

1. 禁止 Web/TUI 框架进入核心或 CLI（无 React/Vue/gin/echo/fiber）
2. 禁止隐式守护进程或自动 spawn（只允许显式 `axis start`）
3. 禁止修改 scheduler/dispatcher/contract 语义而不更新对应 Spec
4. 禁止新增 Agent 自主权而不经 staged-evolution protocol
5. 禁止 push-based context injection（contextpack 只做 preview，opt-in）
6. 禁止无命名空间的 metadata key（前缀：`context.*` `tool.*` `sla.*` `evolution.*` `intent.*` `provider.*` `axis.*`）
7. 禁止输出 secrets——永不 log/打印 API key、token、credentials

## 2. 编码前检查

- [ ] 新行为可经 CLI / 文件系统 / event log 观察（不隐藏）
- [ ] metadata key 有命名空间前缀
- [ ] 测试含边界/安全断言；bug fix 必须带回归测试
- [ ] 改 `internal/kernel/` `cmd/axis/` `internal/contextpack/` `internal/agent/` `internal/memory/` 前读相邻 BOUNDARY.md
- [ ] Windows 一等公民：path/filepath、优雅关闭降级、快照读共享文件
- [ ] 公共函数入口做 nil/空/边界防御；并发体每个 goroutine 有退出路径
- [ ] I/O 与可取消操作首参 `ctx context.Context`；不硬编码 timeout/port/retry

### 2.1 RDM 豁免

仅改文档（≤5 文件、≤100 行、白名单操作）时：单行 `RDM: <描述>` 替代 Phase 声明与 Spec 要求；§1 禁令与 WIKI-SCHEMA 硬约束仍生效。出现 H1 改动/文件删除/lint 新 error/超规模即失去豁免。

## 3. 必读文档

| 优先级 | 文件 | 内容 |
|---|---|---|
| P0 | 本文件 | 宪法 |
| P1 | `docs/status/current-progress.md` | 里程碑 ground truth |
| P2 | `docs/architecture/semantic-boundaries.md` | 模块"不得做"清单 |
| P3 | `docs/architecture/git-conventions.md` | Git 工作流 |

其余 architecture/specs 文档按需查阅，不再作为强制前置阅读。

## 4. 目录边界

编辑前必读相邻 BOUNDARY.md：

| 目录 | 边界要点 |
|------|---------|
| `internal/kernel/` | Scheduler 不直接调 provider；不注入 context assembly |
| `cmd/axis/` | 无 Web/TUI；无隐式守护进程；可脚本化输出；不泄露 secret |
| `internal/contextpack/` | 永不 push context 到 provider prompt；不改 scheduler 语义 |
| `internal/agent/` | 永不绕过 contract 层；不注入 context metadata 到 provider input |
| `internal/memory/` | 永不 push 到 prompt；不物理删除；无外部依赖/后台任务；LF-only |
| `internal/skills/` | 不自动 push skill 内容；无网络访问 |
| `internal/vigil/` | 不阻塞 git；不强制顺序；不对 Agent 隐藏 item |

## 5. Spec-First（按需）

非平凡功能或结构性变更先问：这需要 Spec-RDT 吗？需要则建 `docs/specs/<feature>/{requirements,design,tasks}.md` 三件套并声明状态；小改动直接做。Promotion 由机器可检验的验证标准门控，不由身份门控。

## 6. 语义边界（浓缩）

AgentTask 无执行逻辑；Contract 无 scheduler 策略；Scheduler 无 model 调用/shell；Orchestrator 不存 credentials/渲染 CLI；Dispatcher 不管 admission 策略；Provider 不管 task lifecycle；Tool 不管全局权限；Intent Parser 不直接执行；ContextBundle 不升级权限；EvolutionRun 不隐式改主树。细则见 `docs/architecture/semantic-boundaries.md`。

## 7. 代码风格

Go 1.26 + spf13/cobra；无外部依赖除非绝对必要；metadata 用 `namespace.key`；所有状态变更留痕（event log/metadata）；自然语言产生结构但不执行。

## 8. CLI 输出契约

默认人类可读；`--json` 为稳定 snake_case 机器模式；成功=发生了什么+主 ID+下一步建议；错误=操作+对象 ID+原因+下一步；preview 明确声明未改状态；不依赖颜色；保持行导向。

## 9. 构建与测试

```bash
go build -o axis-dev.exe ./cmd/axis   # 开发二进制
go test -race -short ./...            # 提交前必须过（CI 同款）
gofmt -w . && go vet ./...            # 格式化 + vet
staticcheck ./...                     # 本地装：go install honnef.co/go/tools/cmd/staticcheck@latest
```

- commit message 引用 scope/milestone 标签；禁 "fix typo"/裸 "wip" 进 main
- 每个提交 bisect-safe（独立可编译）
- 永不暂存 `*.exe` `*.test` `coverage.out` `dist/` `.cache/`
- 集成测试加 `testing.Short()` skip

## 10. 工程实践（精选硬规则）

- 进程终止用 `os.Process.Kill()` 或平台抽象，不用 `syscall.Kill`（Windows）
- 安全关键路径检查用 `filepath.Clean`+segment matching，不用 `strings.HasPrefix`
- 永不返回 `(nil, nil)`；永不静默吞 error
- HTTP retry 每次 attempt 用 `bytes.NewReader(body)` 重建 request
- 本地环回 HTTP client 一律 `control.LocalHTTPClient()`（禁 keep-alive 池化——已两次踩坑）
- provider/tool 的 ctx 无 deadline 时必须兜底超时
- 测试不做真实外网调用（httptest/mock）；破坏性用例覆盖 bad JSON/超时/404/500/空 body
- timeout/port/buffer/retry 提取为常量或 config，不散落硬编码

## 11. 演化原则（永久条款）

稳定表面，可替换内部；安全默认（dry-run/preview/redaction/explicit submit）；设计即可审计；小 contract 优于大 control plane；渐进演化（先确定性后自适应）；**审计而非审批是信任机制**——建议信息（memory/context preview）永不自动升级为权威信息。

## 12. 命名与结构

Spec 状态：Draft → Planned → In Progress → Completed | Paused | Deprecated | Cancelled。Completed = 代码+测试+文档同步+用户可见行为已描述。Metadata promotion：多模块依赖/验证需要/CLI 消费者依赖时提升为 typed field。

## 13. 治理

条款分三类：**永久**（§0 结构/§1/§11，修改需论证前提失效）、**渐进**（其余，≥3 次实践证据可改）、**过渡**（声明失效条件）。冲突仲裁：永久 > 渐进 > 过渡。理论-实践矛盾按《实践论》处置：超前→标记 aspirational；束缚（绕过 ≥3 次）→重写；突破（新模式连续 ≥3 次成功）→提炼上浮。

## 14. 前端（过渡条款）

tools/axis-gui 保持 Observatory 只读定位；失效条件：前端废弃或 Agent 自主观察能力成熟。

## 15. 外部工具参考

| 工具 | 入口 | 场景 |
|------|------|------|
| vigil | `axis vigil` | 跨会话工作追踪（内置） |
| research-pipeline | workflow skill | 研究：发现→筛选→深入→扬弃→vigil |
| MindMagnifier | `amp.exe` | 论文/AI 新闻查询 |

# Engineering Guardrails — 问题根除规划

**原则**: 把约束编码进编译器、linter、CI，而不是写进需要人脑记忆的文档。

**来源**: 2026-05-08 ~ 2026-05-11 全部修复日志与问题报告的交叉分析。

---

## 问题 → 根因 → 结构性修复 映射表

### G1. 错误被静默丢弃（#3 高频 bug）

**历史 bug**: `safeMarshal` 用 `_` 忽略 error、render helpers 丢弃 `io.Writer` error、`2>/dev/null` 吞 gofmt 错误

**根因**: `.golangci.yml` 中 `gosec` 排除了 `G104`（Errors unhandled），等于主动关闭了检测

**修复**:
- [x] 移除 `gosec.excludes: G104` → `.golangci.yml`
- [x] 启用 `nilerr` linter → `.golangci.yml`
- [x] 启用 `errchkjson`（Go 1.21+）→ `.golangci.yml`
- [x] CI 使用 `golangci-lint run` 替代单独的 `staticcheck` → `.github/workflows/ci.yml`

---

### G2. 并发资源泄漏（#2 高频 bug）

**历史 bug**: goroutine 泄漏、channel 重复关闭 panic、busy-poll 浪费 CPU、partial claim 不回滚

**根因**: 无结构化并发模式；goroutine 启动无退出保证

**修复**:
- [x] 启用 `copyloopvar` linter → `.golangci.yml`
- [x] 新增 `internal/safego` 包 → `internal/safego/safego.go` + `safego_test.go`
  - `Go(ctx, fn)` — 自动 recover panic
  - `GoWithWaitGroup(ctx, wg, fn)` — recover panic + `wg.Done()` 保证执行
- [ ] 建立 goroutine 纪律：所有 `go func()` 必须接受 ctx 并监听 `ctx.Done()`
  - *渐进式替换：已有 `dispatcher.go`/`orchestrator.go` 中的裸 `go func()` 不强制一次性改完，新代码优先使用 safego*

---

### G3. Context 传播断裂（#4 高频 bug）

**历史 bug**: dispatcher 传 `context.Background()` 给 executor、ContractExecutor 接口无 ctx、human polling 无 ctx.Done()

**根因**: 接口设计阶段未强制 ctx 为第一参数

**修复**:
- [x] 启用 `noctx` linter → `.golangci.yml`
- [x] 启用 `revive` 的 `context-as-argument` 规则 → `.golangci.yml`
- [x] 既有接口已在 core-engine-tdd-fixes 中修复，linter 防止回归

---

### G4. Windows/跨平台兼容性（#1 高频 bug）

**历史 bug**: CRLF stdin 污染、signal.Notify 不工作、路径硬编码、端口释放延迟

**根因**: CI 测试只跑 ubuntu-latest，Windows bug 永远不会在合并前被发现

**修复**:
- [x] CI test job 增加 `windows-latest` 矩阵 → `.github/workflows/ci.yml`
- [x] CI build 命令从 `cmd/axis/main.go` 修正为 `./cmd/axis` → `.github/workflows/ci.yml`
- [x] 覆盖率/上传步骤限定 `ubuntu-latest` → `.github/workflows/ci.yml`

---

### G5. 安全与边界校验缺失（#5 高频 bug）

**历史 bug**: URL 未转义、HTTP client 无超时、Listen 接受非 loopback、文件路径 sibling-prefix 逃逸

**修复**:
- [x] 启用 `bodyclose` linter → `.golangci.yml`
- [x] 启用 `reassign` linter → `.golangci.yml`
- [x] 既有安全修复已在 context-preflight-fix 和 /review 中落地，linter 防止回归

---

### G6. golangci-lint 未在 CI 中使用

**根因**: `.golangci.yml` 配置存在但 CI 只跑单独的 `staticcheck`，新增的 linter 全部无效

**修复**:
- [x] CI lint job: `staticcheck` → `golangci-lint run ./...` → `.github/workflows/ci.yml`

---

## 实施顺序

| 步骤 | 文件 | 改动 |
|---|---|---|
| 1 | `.golangci.yml` | 移除 G104 排除、新增 5 个 linter、新增 revive context-as-argument 规则 |
| 2 | `.github/workflows/ci.yml` | lint: golangci-lint、test: +windows-latest、build: 修正命令 |
| 3 | `internal/safego/safego.go` | 结构化 goroutine 启动器（ctx + panic recovery + 退出保证） |
| 4 | `docs/architecture/engineering-guardrails.md` | 本文件，标记完成状态 |

**预期效果**: 步骤 1+2 完成后，历史 top-5 bug 类别中的 4 类将被 CI 自动拦截，无需依赖人工记忆。

---

## 不做的事（避免过度工程化）

- ❌ 不写"编码规范文档"让人去读 — 用 linter 替代
- ❌ 不加自定义 go vet analyzer — 维护成本高，标准 linter 已覆盖
- ❌ 不强制 pre-commit hook — CI 是唯一真相源，本地开发保持灵活
- ❌ 不新增文档同步 CI 检查 — 4 文档同步靠 Agent 自律已够，自动化会变成过度控制

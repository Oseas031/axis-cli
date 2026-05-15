# Git 工作流重构计划

> vigil-7d3294 | 参照 git-guides/git-best-practices
> 实现 docs/architecture/axis-system-conventions.md + CLAUDE.md §9

## 1. 现状诊断

| 问题 | 证据 | 风险 |
|------|------|------|
| 47 commits 未 push | `git status` ahead of origin | 硬盘故障 = 全部丢失 |
| 直接 main commit | `git log` 无 branch merge | 无法隔离实验、无法回滚 feature |
| God commits (121 files, 5000+ lines) | `git log --shortstat` | 不可 bisect、不可 cherry-pick |
| Commit message 不一致 | 有 `feat(x):` 也有 `integrate:` 也有纯名词 | 不可追溯、不可自动生成 changelog |
| 无 CI | 无 `.github/workflows/` | 无自动验证、§9 rule #9 空转 |
| 无 pre-commit hook | 无 `.git/hooks/` 自定义 | 格式/lint 靠自觉 |
| 死分支堆积 | 4 个未清理分支 | 噪音 |
| .gitignore 不完整 | `.kiro/skills/` 需要 `git add -f` | 摩擦 |

## 2. 目标状态

```
分支策略: main (protected) ← feature/xxx, fix/xxx, docs/xxx, research/xxx
Commit 规范: Conventional Commits (type(scope): description)
粒度: 一个逻辑关注点 = 一个 commit，≤5 files 或 ≤200 lines 为宜
CI: GitHub Actions (go build + go test -race + go vet + staticcheck)
Hook: commit-msg 格式检查 + pre-commit go vet
Push: 每次工作结束 push（至少每天一次）
```

## 3. 实施步骤（按优先级）

### P0: 立即止血

- [ ] **Push 现有 47 commits 到 origin/main**（备份）
- [ ] **清理死分支**：删除 `backup-before-cleanup`、`feature/full-architecture`、`m5/autonomy`（确认已合并或不需要后）
- [ ] **修复 .gitignore**：unignore `.kiro/skills/`

### P1: 建立规范

- [ ] **分支命名规范**（写入 docs/architecture/git-conventions.md）：
  ```
  feature/<scope>-<description>   新功能
  fix/<scope>-<description>       Bug 修复
  docs/<description>              文档变更
  research/<paper-id>             研究管线
  refactor/<scope>-<description>  重构
  chore/<description>             非功能性维护
  ```
- [ ] **Commit message 规范**（Conventional Commits）：
  ```
  <type>(<scope>): <description>  # header ≤70 chars
                                   # blank line
  <body>                           # what + why, wrap at 72
                                   # blank line
  <footer>                         # vigil: xxx | Refs: spec-id
  ```
  Types: `feat`, `fix`, `docs`, `refactor`, `test`, `chore`, `perf`, `research`
  Scope: 模块名 (`agent`, `kernel`, `contextpack`, `model`, `vigil`, `gui`)
- [ ] **分级策略**（何时用 branch vs 直接 main）：
  ```
  直接 main: RDM 操作（纯文档 ≤5 files ≤100 lines）
  Feature branch: 代码变更 >5 files 或 >200 lines 或跨 ≥2 模块
  ```

### P2: 工具层强制

- [ ] **commit-msg hook**：验证 Conventional Commits 格式
  ```bash
  #!/bin/sh
  # .git/hooks/commit-msg
  if ! head -1 "$1" | grep -qE '^(feat|fix|docs|refactor|test|chore|perf|research|rdm)(\(.+\))?: .{1,70}$'; then
    echo "ERROR: Commit message does not match Conventional Commits format"
    echo "Format: type(scope): description"
    exit 1
  fi
  ```
- [ ] **pre-commit hook**：`go vet ./...` + 检查无 .exe/.out 文件暂存
- [ ] **GitHub Actions CI**（`.github/workflows/ci.yml`）：
  ```yaml
  on: [push, pull_request]
  jobs:
    test:
      runs-on: ubuntu-latest
      steps:
        - uses: actions/checkout@v4
        - uses: actions/setup-go@v5
          with: { go-version: '1.26' }
        - run: go build ./...
        - run: go test -race -short ./...
        - run: go vet ./...
  ```

### P3: 流程集成

- [ ] **更新 CLAUDE.md §9**：引用 `docs/architecture/git-conventions.md`，删除重复规则
- [ ] **Agent 行为约束**：Phase III 结束时检查 commit 粒度，超过阈值时拆分
- [ ] **自动 push**：每次 commit 后提示 push（或 Agent 自动执行）

## 4. 对 CLAUDE.md §9 现有规则的审计

| 现有规则 | 执行情况 | 修正 |
|----------|----------|------|
| "每个 commit 引用 Spec-RDT ID 或 scope 标签" | 部分执行 | 改为 Conventional Commits footer |
| "无构建产物" | 执行良好 | 保持 |
| "Bisect-safe" | 违反（God commits） | 通过粒度规范 + hook 强制 |
| "push 后自动监控 CI" | 无 CI 可监控 | 先建 CI |
| "不 push directly to main" | 系统性违反 | 分级策略（小变更允许，大变更必须 branch） |

## 5. 与 Axis 设计原则的对齐

- **CLI First**：hook 是 shell 脚本，CI 是 YAML，无 GUI 依赖 ✓
- **可观察**：commit message 是审计轨迹 ✓
- **渐进演化**：先 P0 止血，再 P1 规范，最后 P2 强制 ✓
- **Windows 一等公民**：hook 需要跨平台（PowerShell + bash 双版本或用 Go 写）⚠️

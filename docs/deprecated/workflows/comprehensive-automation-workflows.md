---
description: 项目全流程自动化工作流架构
---

# Axis CLI 项目自动化工作流架构

基于软件工程最佳实践，Axis CLI 项目需要以下 7 个核心工作流来实现端到端的自动化。

---

## 工作流架构概览

```
开发 → PR → CI → 安全 → 测试 → 构建 → 发布 → 监控
  ↓      ↓    ↓     ↓      ↓      ↓      ↓      ↓
[1]   [2]   [3]   [4]    [5]    [6]    [7]    [8]
```

---

## 工作流 1: 开发阶段自动化（Development Workflow）

**触发条件**: Pre-commit（本地）

**职责**:
- 代码格式化（gofmt）
- Import 排序（goimports）
- 本地 lint 检查（staticcheck）
- 基础测试（go test -short）
- 提交信息规范检查（conventional commits）

**工具**: Pre-commit hooks

**配置文件**: `.pre-commit-config.yaml`

**预期收益**:
- 在代码推送到远程前发现问题
- 减少 CI 失败率
- 统一代码风格

---

## 工作流 2: PR 质量检查工作流（PR Quality Check Workflow）

**触发条件**: 创建/更新 Pull Request

**职责**:
- 完整测试套件执行
- 代码覆盖率检查（≥60%）
- 复杂度分析（cyclomatic complexity）
- 代码重复检查
- 安全扫描（SAST）
- 依赖漏洞扫描
- 自动化代码审查（AI 辅助）
- 变更影响分析

**工具**: GitHub Actions

**配置文件**: `.github/workflows/pr-check.yml`

**预期收益**:
- 确保每次 PR 都符合质量标准
- 防止低质量代码合并
- 提供详细的变更分析

---

## 工作流 3: 持续集成工作流（CI Workflow）

**触发条件**: Push 到 main/develop 分支

**职责**:
- 格式检查
- Vet 检查
- Lint 检查
- 单元测试（带 race detector）
- 集成测试
- 多版本 Go 测试
- 多平台构建
- 制品生成
- 覆盖率报告上传

**工具**: GitHub Actions

**配置文件**: `.github/workflows/ci.yml`

**预期收益**:
- 每次提交都经过完整验证
- 多平台兼容性保证
- 质量趋势可追踪

---

## 工作流 4: 安全扫描工作流（Security Scanning Workflow）

**触发条件**: 
- 定时（每日）
- Push 到 main
- 创建 PR

**职责**:
- 静态应用安全测试（SAST）
- 依赖漏洞扫描（SCA）
- 容器镜像扫描
- 密钥泄露检测
- 许可证合规检查
- 安全报告生成

**工具**: Gosec, Trivy, Dependabot

**配置文件**: `.github/workflows/security.yml`

**预期收益**:
- 主动发现安全漏洞
- 依赖安全持续监控
- 合规性保证

---

## 工作流 5: 持续交付工作流（CD Workflow）

**触发条件**: Push Git Tag（v*）

**职责**:
- 多平台二进制构建
- Docker 镜像构建
- 制品签名（Cosign）
- SHA256 校验和生成
- Changelog 自动生成
- Release Notes 生成
- GitHub Release 创建
- 制品上传
- 版本信息注入

**工具**: Goreleaser, Cosign

**配置文件**: `.github/workflows/cd.yml`, `.goreleaser.yml`

**预期收益**:
- 一键发布多平台版本
- 制品可信性保证
- 发布流程标准化

---

## 工作流 6: 文档自动化工作流（Documentation Workflow）

**触发条件**:
- Push 到 main
- 代码变更（internal/**/*.go, cmd/**/*.go）

**职责**:
- API 文档生成（godoc）
- 架构图更新
- README 自动更新（版本号）
- 文档部署到 GitHub Pages
- 变更日志生成

**工具**: godoc, mkdocs

**配置文件**: `.github/workflows/docs.yml`

**预期收益**:
- 文档始终与代码同步
- 降低文档维护成本
- 提高文档准确性

---

## 工作流 7: 监控与可观测性工作流（Monitoring Workflow）

**触发条件**:
- 定时（每小时）
- CI/CD 完成后
- 手动触发

**职责**:
- CI/CD 指标收集（构建时间、成功率）
- 测试覆盖率趋势分析
- 性能基准测试对比
- 依赖健康检查
- 告警触发（失败率过高）
- Dashboard 更新

**工具**: GitHub Actions Metrics, Prometheus

**配置文件**: `.github/workflows/monitoring.yml`

**预期收益**:
- 实时了解项目健康状况
- 性能回归检测
- 主动发现问题

---

## 工作流依赖关系

```
[1] 开发阶段
    ↓
[2] PR 质量检查 ← [4] 安全扫描（PR）
    ↓
[3] 持续集成 ← [4] 安全扫描（定时）
    ↓
[5] 持续交付 ← [3] 通过
    ↓
[6] 文档自动化 ← [5] 完成
    ↓
[7] 监控与可观测性 ← 所有工作流
```

---

## 工作流详细配置

### 工作流 1: Pre-commit Hooks

```yaml
# .pre-commit-config.yaml
repos:
  - repo: https://github.com/pre-commit/pre-commit-hooks
    hooks:
      - id: trailing-whitespace
      - id: end-of-file-fixer
      - id: check-yaml
      - id: check-added-large-files
        args: ['--maxkb=1000']
  - repo: local
    hooks:
      - id: gofmt
        name: gofmt
        entry: gofmt -s -w
        language: system
        files: \.go$
      - id: goimports
        name: goimports
        entry: goimports -w
        language: system
        files: \.go$
      - id: staticcheck
        name: staticcheck
        entry: staticcheck ./...
        language: system
        files: \.go$
      - id: go-test-short
        name: go test short
        entry: go test -short ./...
        language: system
        files: \.go$
```

### 工作流 2: PR 质量检查

```yaml
# .github/workflows/pr-check.yml
name: PR Quality Check
on:
  pull_request:
    branches: [main, develop]
    types: [opened, synchronize, reopened]

jobs:
  quality-gate:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0
      
      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.26'
      
      - name: Full test suite
        run: go test -v -race -coverprofile=coverage.out ./...
      
      - name: Coverage check
        run: |
          COVERAGE=$(go tool cover -func=coverage.out | grep total | awk '{print $3}' | sed 's/%//')
          echo "Coverage: $COVERAGE%"
          if (( $(echo "$COVERAGE < 60" | bc -l) )); then
            echo "❌ Coverage below 60%"
            exit 1
          fi
      
      - name: Complexity analysis
        run: |
          go install github.com/fzipp/gocyclo/cmd/gocyclo@latest
          gocyclo -over 15 . || echo "High complexity functions found"
      
      - name: Duplicate code check
        run: |
          go install github.com/mibk/dupl@latest
          dupl -threshold 100 . || echo "Duplicate code found"
      
      - name: Security scan
        uses: securego/gosec@master
        with:
          args: -no-fail -fmt sarif -out gosec-results.sarif ./...
      
      - name: Dependency scan
        uses: actions/dependency-review-action@v4
      
      - name: Upload SARIF
        uses: github/codeql-action/upload-sarif@v2
        with:
          sarif_file: gosec-results.sarif
```

### 工作流 3: 持续集成

```yaml
# .github/workflows/ci.yml
name: CI
on:
  push:
    branches: [main, develop]
  pull_request:
    branches: [main, develop]

jobs:
  lint:
    name: Lint
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.26'
      - name: Format check
        run: |
          if [ -n "$(gofmt -l .)" ]; then
            echo "Code is not formatted"
            exit 1
          fi
      - name: Vet
        run: go vet ./...
      - name: Staticcheck
        run: staticcheck ./...

  test:
    name: Test
    runs-on: ubuntu-latest
    needs: lint
    strategy:
      matrix:
        go-version: ['1.26', '1.27']
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: ${{ matrix.go-version }}
      - name: Cache Go modules
        uses: actions/cache@v3
        with:
          path: ~/go/pkg/mod
          key: ${{ runner.os }}-go-${{ hashFiles('**/go.sum') }}
      - name: Unit tests
        run: go test -v -race -coverprofile=coverage.out ./...
      - name: Upload coverage
        uses: codecov/codecov-action@v3

  build:
    name: Build
    runs-on: ${{ matrix.os }}
    needs: test
    strategy:
      matrix:
        os: [ubuntu-latest, windows-latest, macos-latest]
        go-version: ['1.26']
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: ${{ matrix.go-version }}
      - name: Build
        run: go build -o axis-${{ matrix.os }} cmd/axis/main.go
      - name: Upload artifact
        uses: actions/upload-artifact@v3
        with:
          name: axis-${{ matrix.os }}
          path: axis-${{ matrix.os }}
```

### 工作流 4: 安全扫描

```yaml
# .github/workflows/security.yml
name: Security Scanning
on:
  schedule:
    - cron: '0 0 * * *'
  push:
    branches: [main]
  pull_request:
    branches: [main]

jobs:
  sast:
    name: SAST
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - name: Run Gosec
        uses: securego/gosec@master
        with:
          args: ./...
  
  sca:
    name: SCA
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - name: Run Trivy
        uses: aquasecurity/trivy-action@master
        with:
          scan-type: 'fs'
          scan-ref: '.'
          format: 'sarif'
          output: 'trivy-results.sarif'
      - name: Upload Trivy results
        uses: github/codeql-action/upload-sarif@v2
        with:
          sarif_file: 'trivy-results.sarif'
  
  license:
    name: License Compliance
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - name: FOSSA Scan
        uses: fossas/fossa-action@v1
        with:
          api-key: ${{ secrets.FOSSA_API_KEY }}
```

### 工作流 5: 持续交付

```yaml
# .github/workflows/cd.yml
name: CD
on:
  push:
    tags:
      - 'v*'

jobs:
  release:
    name: Release
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0
      
      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.26'
      
      - name: Import GPG key
        uses: crazy-max/ghaction-import-gpg@v6
        with:
          gpg_private_key: ${{ secrets.GPG_PRIVATE_KEY }}
          passphrase: ${{ secrets.PASSPHRASE }}
      
      - name: Run GoReleaser
        uses: goreleaser/goreleaser-action@v5
        with:
          version: latest
          args: release --clean
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
          GPG_FINGERPRINT: ${{ secrets.GPG_FINGERPRINT }}
```

### 工作流 6: 文档自动化

```yaml
# .github/workflows/docs.yml
name: Documentation
on:
  push:
    branches: [main]
    paths:
      - 'internal/**/*.go'
      - 'cmd/**/*.go'
      - 'docs/**'

jobs:
  api-docs:
    name: API Documentation
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      
      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.26'
      
      - name: Generate godoc
        run: |
          go install golang.org/x/tools/cmd/godoc@latest
          mkdir -p docs/api
          godoc -html -url=/internal/kernel/scheduler/ internal/kernel/scheduler > docs/api/scheduler.html
          godoc -html -url=/internal/kernel/dispatcher/ internal/kernel/dispatcher > docs/api/dispatcher.html
          godoc -html -url=/internal/kernel/orchestrator/ internal/kernel/orchestrator > docs/api/orchestrator.html
      
      - name: Deploy to GitHub Pages
        uses: peaceiris/actions-gh-pages@v3
        with:
          github_token: ${{ secrets.GITHUB_TOKEN }}
          publish_dir: ./docs/api
```

### 工作流 7: 监控与可观测性

```yaml
# .github/workflows/monitoring.yml
name: Monitoring
on:
  schedule:
    - cron: '0 * * * *'  # 每小时
  workflow_run:
    workflows: [CI, CD]
    types: [completed]

jobs:
  metrics:
    name: Collect Metrics
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      
      - name: Collect CI metrics
        uses: github/ci-metrics-action@v1
        with:
          metrics: build_success_rate,build_duration,test_coverage
      
      - name: Run benchmarks
        run: |
          go test -bench=. -benchmem -benchtime=10s ./... > benchmark-results.txt
      
      - name: Compare with baseline
        run: |
          if [ -f benchmark-baseline.txt ]; then
            benchstat benchmark-baseline.txt benchmark-results.txt || echo "Performance regression detected"
          fi
      
      - name: Send to monitoring
        uses: 8398a7/action-slack@v3
        with:
          status: ${{ job.status }}
          text: 'Hourly metrics collected'
        env:
          SLACK_WEBHOOK_URL: ${{ secrets.SLACK_WEBHOOK }}
```

---

## 工作流优先级

### 高优先级（立即实施）
1. 工作流 1: Pre-commit Hooks
2. 工作流 3: 持续集成
3. 工作流 5: 持续交付

### 中优先级（1-2周内）
4. 工作流 2: PR 质量检查
5. 工作流 4: 安全扫描

### 低优先级（1个月内）
6. 工作流 6: 文档自动化
7. 工作流 7: 监控与可观测性

---

## 工作流优化建议

### 性能优化
- 使用 GitHub Actions 缓存
- 并行执行独立任务
- 使用矩阵策略减少重复配置

### 成本优化
- 使用自托管 Runner
- 合并相似工作流
- 定时任务降低频率

### 可维护性优化
- 使用可重用 Actions
- 配置文件模块化
- 文档化每个工作流

---

## 总结

Axis CLI 项目需要 **7 个核心工作流** 来实现端到端自动化：

1. **开发阶段自动化** - 本地质量门禁
2. **PR 质量检查** - PR 级别质量保证
3. **持续集成** - 提交级别验证
4. **安全扫描** - 安全合规保证
5. **持续交付** - 自动化发布
6. **文档自动化** - 文档同步
7. **监控与可观测性** - 健康监控

这 7 个工作流覆盖了从开发到发布的完整软件生命周期，确保项目质量、安全和可维护性。

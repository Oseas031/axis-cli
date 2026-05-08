---
description: 软件工程范式驱动的CI/CD工作流改进
---

# 软件工程范式驱动的CI/CD工作流改进

本文档应用软件工程范式和最佳实践来改进Axis CLI项目的CI/CD工作流。

## 核心软件工程范式

### 1. DevOps 范式
**原则**: 开发与运维一体化，自动化整个软件交付生命周期

**应用**:
- 基础设施即代码（IaC）
- 持续集成/持续部署（CI/CD）
- 监控与可观测性
- 自动化测试
- 快速反馈循环

### 2. GitOps 范式
**原则**: Git作为单一事实来源，通过Git操作管理基础设施和应用部署

**应用**:
- 所有配置存储在Git中
- PR作为变更审批机制
- 自动同步到目标环境
- 可审计的变更历史

### 3. 测试驱动开发（TDD）
**原则**: 先写测试，再写代码，确保代码质量

**应用**:
- 测试金字塔（单元测试 > 集成测试 > E2E测试）
- 测试覆盖率要求
- 自动化测试执行
- 测试结果可视化

### 4. 持续交付（CD）
**原则**: 任何变更都可以随时部署到生产环境

**应用**:
- 自动化部署流水线
- 蓝绿部署/金丝雀发布
- 自动回滚机制
- 特性开关

### 5. 安全左移（DevSecOps）
**原则**: 安全集成到开发流程的早期阶段

**应用**:
- SAST（静态应用安全测试）
- 依赖漏洞扫描
- 容器安全扫描
- 安全代码审查

### 6. 可观测性
**原则**: 系统行为可被观察、理解和诊断

**应用**:
- 日志聚合
- 指标监控
- 分布式追踪
- 告警机制

---

## 改进后的工作流架构

### 阶段 1: 开发阶段

**Pre-commit Hooks（本地质量门禁）**
```yaml
# .pre-commit-config.yaml
repos:
  - repo: https://github.com/pre-commit/pre-commit-hooks
    hooks:
      - id: trailing-whitespace
      - id: end-of-file-fixer
      - id: check-yaml
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
      - id: go-test
        name: go test
        entry: go test -short ./...
        language: system
        files: \.go$
```

**Commit 规范**
```bash
# 使用 Conventional Commits 规范
git commit -m "feat: add new scheduler feature"
git commit -m "fix: resolve circular dependency bug"
git commit -m "docs: update README"
```

---

### 阶段 2: PR 阶段

**自动化 PR 检查**
```yaml
# .github/workflows/pr-check.yml
name: PR Check
on:
  pull_request:
    branches: [main, develop]

jobs:
  quality-gate:
    runs-on: ubuntu-latest
    steps:
      - name: Checkout
        uses: actions/checkout@v4
        
      - name: Set up Go
        uses: actions/setup-go@v5
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
        
      - name: Static analysis
        run: staticcheck ./...
        
      - name: Security scan
        uses: securego/gosec@master
        with:
          args: ./...
          
      - name: Unit tests
        run: go test -v -race -coverprofile=coverage.out ./...
        
      - name: Coverage check
        run: |
          COVERAGE=$(go tool cover -func=coverage.out | grep total | awk '{print $3}' | sed 's/%//')
          echo "Coverage: $COVERAGE%"
          if (( $(echo "$COVERAGE < 60" | bc -l) )); then
            echo "Coverage below 60%"
            exit 1
          fi
          
      - name: Integration tests
        run: go test -tags=integration -v ./integration/...
        
      - name: Benchmark
        run: go test -bench=. -benchmem ./...
```

**代码审查自动化**
```yaml
# .github/workflows/code-review.yml
- name: Code Review
  uses: github/super-linter@v5
  env:
    DEFAULT_BRANCH: main
    GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```

---

### 阶段 3: 分支保护

**GitHub 分支保护规则**
```yaml
# 通过 GitHub UI 或 API 配置
- 分支: main
- 要求 PR 审查: 至少 1 个
- 要求状态检查通过: 所有 CI 检查
- 要求更新到最新分支: 是
- 允许强制推送: 否
- 要求线性历史: 是
```

**CODEOWNERS 文件**
```
# .github/CODEOWNERS
# 核心模块需要核心团队审查
internal/kernel/ @core-team

# CLI 需要特定审查者
cmd/axis/ @cli-maintainer

# 所有变更需要至少一个审查者
* @reviewers
```

---

### 阶段 4: 持续集成

**CI 流水线**
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
      - name: Format
        run: gofmt -s -w .
      - name: Vet
        run: go vet ./...
      - name: Staticcheck
        run: staticcheck ./...

  security:
    name: Security
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - name: Run Gosec
        uses: securego/gosec@master
        with:
          args: ./...
      - name: Run Trivy
        uses: aquasecurity/trivy-action@master
        with:
          scan-type: 'fs'
          scan-ref: '.'
      - name: Dependency scan
        uses: actions/dependency-review-action@v4

  test:
    name: Test
    runs-on: ubuntu-latest
    needs: [lint, security]
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
    needs: [test]
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

---

### 阶段 5: 持续交付

**CD 流水线**
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
      
      - name: Run GoReleaser
        uses: goreleaser/goreleaser-action@v5
        with:
          version: latest
          args: release --clean
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
      
      - name: Generate changelog
        uses: orhun/git-cliff-action@v2
        with:
          config: cliff.toml
          args: --verbose
      
      - name: Sign artifacts
        uses: sigstore/cosign-installer@v3
      
      - name: Sign release artifacts
        run: cosign sign --yes ${{ env.ARTIFACT }}
```

**Goreleaser 配置**
```yaml
# .goreleaser.yml
project_name: axis
before:
  hooks:
    - go mod tidy
    - go generate ./...
builds:
  - env:
      - CGO_ENABLED=0
    goos:
      - linux
      - windows
      - darwin
    goarch:
      - amd64
      - arm64
    ldflags:
      - -s -w -X main.version={{.Version}} -X main.commit={{.Commit}} -X main.date={{.Date}}
    main: ./cmd/axis/main.go
    binary: axis
checksum:
  name_template: 'checksums.txt'
archives:
  - format: tar.gz
    name_template: "{{ .ProjectName }}_{{ .Version }}_{{ .Os }}_{{ .Arch }}"
changelog:
  sort: asc
  filters:
    exclude:
      - '^docs:'
      - '^test:'
```

---

### 阶段 6: 可观测性

**监控与告警**
```yaml
# .github/workflows/monitoring.yml
name: Monitoring
on:
  schedule:
    - cron: '0 0 * * *'  # 每天运行
  workflow_dispatch:

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
          
      - name: Send to monitoring
        uses: 8398a7/action-slack@v3
        with:
          status: ${{ job.status }}
          text: 'Daily metrics collected'
        env:
          SLACK_WEBHOOK_URL: ${{ secrets.SLACK_WEBHOOK }}
```

**性能基准测试**
```go
// internal/benchmark/scheduler_bench_test.go
package benchmark

import (
    "testing"
    "github.com/axis-cli/axis/internal/kernel/scheduler"
)

func BenchmarkSchedulerSubmit(b *testing.B) {
    s := scheduler.NewScheduler(nil, nil)
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        task := createTestTask(i)
        s.Submit(task)
    }
}

func BenchmarkSchedulerGetNextTask(b *testing.B) {
    s := setupSchedulerWithTasks(1000)
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        s.GetNextTask()
    }
}
```

---

### 阶段 7: 文档自动化

**API 文档生成**
```yaml
# .github/workflows/docs.yml
name: Generate Documentation
on:
  push:
    branches: [main]
    paths:
      - 'internal/**/*.go'
      - 'cmd/**/*.go'

jobs:
  docs:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      
      - name: Generate godoc
        run: |
          go install golang.org/x/tools/cmd/godoc@latest
          godoc -html -url=/internal/kernel/scheduler/ internal/kernel/scheduler > docs/api/scheduler.html
          
      - name: Deploy to GitHub Pages
        uses: peaceiris/actions-gh-pages@v3
        with:
          github_token: ${{ secrets.GITHUB_TOKEN }}
          publish_dir: ./docs/api
```

---

## 质量门禁（Quality Gates）

### 定义质量标准

```yaml
# .github/quality-gates.yml
quality_gates:
  coverage:
    minimum: 60%
    target: 80%
  complexity:
    maximum: 15
  security:
    critical: 0
    high: 0
    medium: 5
  performance:
    max_build_time: 10m
    max_test_time: 5m
```

### 自动化质量检查

```yaml
- name: Quality Gate Check
  run: |
    # 覆盖率检查
    COVERAGE=$(go tool cover -func=coverage.out | grep total | awk '{print $3}' | sed 's/%//')
    if (( $(echo "$COVERAGE < 60" | bc -l) )); then
      echo "❌ Coverage gate failed: $COVERAGE% < 60%"
      exit 1
    fi
    echo "✅ Coverage gate passed: $COVERAGE%"
    
    # 复杂度检查
    COMPLEXITY=$(gocyclo -over 15 .)
    if [ -n "$COMPLEXITY" ]; then
      echo "❌ Complexity gate failed"
      exit 1
    fi
    echo "✅ Complexity gate passed"
```

---

## 分支策略（Branching Strategy）

### Git Flow 简化版

```
main (生产)
  ↑
  ↓
develop (开发)
  ↑
  ↓
feature/* (功能分支)
bugfix/* (修复分支)
```

**工作流程**:
1. 从 develop 创建 feature 分支
2. 开发并提交到 feature 分支
3. 创建 PR 到 develop
4. 通过 CI/CD 检查
5. 代码审查通过
6. 合并到 develop
7. 定期从 develop 创建 release 分支
8. release 分支测试通过后合并到 main

---

## 配置管理（Configuration Management）

### 环境配置

```yaml
# configs/config.dev.yaml
environment: development
log_level: debug
database:
  host: localhost
  port: 5432
  name: axis_dev
```

```yaml
# configs/config.prod.yaml
environment: production
log_level: info
database:
  host: prod-db.example.com
  port: 5432
  name: axis_prod
```

**配置验证**
```yaml
- name: Validate config
  run: |
    go install github.com/a8m/envsubst/cmd/envsubst@latest
    envsubst < configs/config.${{ env.ENVIRONMENT }}.yaml > config.yaml
    go run ./cmd/config-validator config.yaml
```

---

## 秘钥管理（Secrets Management）

### GitHub Secrets

```yaml
# 通过 GitHub UI 或 API 设置
secrets:
  DATABASE_URL
  API_KEY
  SLACK_WEBHOOK
  SIGNING_KEY
```

**使用 Secrets**
```yaml
- name: Deploy
  env:
    DATABASE_URL: ${{ secrets.DATABASE_URL }}
  run: |
    go run ./cmd/deploy --database-url=$DATABASE_URL
```

---

## 实施路线图

### 第一阶段（立即实施）
- ✅ Pre-commit hooks
- ✅ 分支保护规则
- ✅ CODEOWNERS 文件
- ✅ Conventional Commits

### 第二阶段（1-2周）
- ⏳ 安全扫描（Gosec, Trivy）
- ⏳ 依赖漏洞扫描
- ⏳ 质量门禁
- ⏳ 构建缓存

### 第三阶段（1个月）
- ⏳ 集成测试
- ⏳ 性能基准测试
- ⏳ Goreleaser 配置
- ⏳ 自动化文档生成

### 第四阶段（2-3个月）
- ⏳ 监控与告警
- ⏳ 配置管理
- ⏳ 秘钥管理
- ⏳ 可观测性平台

---

## 工具链

### 开发工具
- **格式化**: gofmt, goimports
- **Lint**: staticcheck, golangci-lint
- **测试**: go test, go test -race
- **覆盖率**: go tool cover
- **基准测试**: go test -bench

### CI/CD 工具
- **平台**: GitHub Actions
- **容器**: Docker
- **发布**: Goreleaser
- **文档**: godoc

### 安全工具
- **SAST**: Gosec
- **依赖扫描**: Dependabot
- **容器扫描**: Trivy
- **签名**: Cosign

### 监控工具
- **日志**: ELK Stack
- **指标**: Prometheus
- **追踪**: Jaeger
- **告警**: Alertmanager

---

## 最佳实践总结

1. **自动化一切**: 从格式化到部署，尽可能自动化
2. **快速反馈**: 在开发早期发现问题
3. **质量门禁**: 设置明确的质量标准
4. **可追溯性**: 所有变更都有记录
5. **安全性**: 安全集成到开发流程
6. **可观测性**: 系统行为可见
7. **文档化**: 代码和流程都有文档
8. **持续改进**: 定期评估和优化流程

---

## 相关文档

- [CI/CD质量改进工作流](./ci-cd-quality-improvement-workflow.md)
- [GitHub Actions文档](https://docs.github.com/en/actions)
- [Goreleaser文档](https://goreleaser.com/)
- [Pre-commit文档](https://pre-commit.com/)

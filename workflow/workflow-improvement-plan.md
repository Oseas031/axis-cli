# 工作流改进计划

**制定日期**: 2026-05-08
**目标**: 基于复盘报告，修复工作流不足，提升开发效率和代码质量

---

## 改进优先级

### 高优先级（立即实施）
1. 添加 pre-commit hook - 在本地阶段发现问题
2. 添加依赖兼容性检查 - 避免使用废弃参数

### 中优先级（里程碑1完成后）
3. 改进测试策略 - 添加集成测试
4. 建立工作流定期审查机制

### 低优先级（里程碑2）
5. 改进错误追踪和通知 - 添加 Slack/邮件通知

---

## 详细实施计划

### 阶段 1: Pre-commit Hook（高优先级）

#### 目标
在代码提交前自动执行基本检查，在本地阶段发现问题，减少 CI 失败。

#### 实施步骤

**步骤 1.1: 创建 pre-commit 配置文件**
- 文件: `.pre-commit-config.yaml`
- 内容:
  ```yaml
  repos:
    - repo: local
      hooks:
        - id: go-fmt
          name: go fmt
          entry: gofmt
          language: system
          args: [-s, -l, .]
          pass_filenames: false
          
        - id: go-vet
          name: go vet
          entry: go vet
          language: system
          args: [./...]
          
        - id: staticcheck
          name: staticcheck
          entry: staticcheck
          language: system
          args: [./...]
          
        - id: go-test
          name: go test
          entry: go test
          language: system
          args: [-short, ./...]
          
        - id: package-naming
          name: package naming check
          entry: bash
          language: system
          args:
            - -c
            - |
              for dir in $(find . -type d -name "internal/*"); do
                pkg=$(basename "$dir")
                if [[ "$pkg" =~ _ ]]; then
                  echo "Package name $pkg contains underscore, not allowed"
                  exit 1
                fi
              done
  ```

**步骤 1.2: 创建 pre-commit 安装脚本**
- 文件: `scripts/install-pre-commit.sh`
- 内容:
  ```bash
  #!/bin/bash
  # Install pre-commit hooks
  
  echo "Installing pre-commit..."
  
  # Check if pre-commit is installed
  if ! command -v pre-commit &> /dev/null; then
      echo "pre-commit not found, installing..."
      pip install pre-commit
  fi
  
  # Install hooks
  pre-commit install
  
  echo "Pre-commit hooks installed successfully"
  ```

**步骤 1.3: 更新 Dev Workflow**
- 文件: `.github/workflows/dev-workflow.yml`
- 添加 pre-commit 验证步骤:
  ```yaml
  - name: Verify pre-commit hooks
    run: |
      pre-commit run --all-files
  ```

**步骤 1.4: 更新文档**
- 文件: `docs/QUICKSTART.md`
- 添加 pre-commit 安装说明:
  ```markdown
  ## 开发环境设置
  
  ### 安装 pre-commit hooks
  ```bash
  pip install pre-commit
  pre-commit install
  ```
  ```

#### 验证标准
- ✅ pre-commit 配置文件创建
- ✅ 安装脚本可执行
- ✅ Dev Workflow 包含 pre-commit 验证
- ✅ 文档更新完成
- ✅ 本地测试 pre-commit hook 工作正常

---

### 阶段 2: 依赖兼容性检查（高优先级）

#### 目标
定期检查 Go 工具链和依赖的兼容性，避免使用废弃参数。

#### 实施步骤

**步骤 2.1: 创建依赖检查脚本**
- 文件: `scripts/check-dependencies.sh`
- 内容:
  ```bash
  #!/bin/bash
  # Check Go toolchain and dependency compatibility
  
  echo "Checking Go version..."
  go version
  
  echo "Checking for deprecated tools..."
  
  # Check godoc
  if command -v godoc &> /dev/null; then
      if godoc -help 2>&1 | grep -q "\-html"; then
          echo "✓ godoc -html is supported"
      else
          echo "✗ godoc -html is deprecated, use go doc instead"
          exit 1
      fi
  fi
  
  echo "Checking golang.org/x/tools version..."
  go list -m -versions golang.org/x/tools 2>/dev/null || echo "Unable to check"
  
  echo "Dependency check completed"
  ```

**步骤 2.2: 在 CI Workflow 中添加检查**
- 文件: `.github/workflows/ci.yml`
- 在 format job 之前添加:
  ```yaml
  dependency-check:
    name: Dependency Check
    runs-on: ubuntu-latest
    steps:
      - name: Checkout code
        uses: actions/checkout@v4
      
      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.26'
      
      - name: Run dependency check
        run: bash scripts/check-dependencies.sh
  ```

**步骤 2.3: 创建定期检查工作流**
- 文件: `.github/workflows/dependency-audit.yml`
- 内容:
  ```yaml
  name: Dependency Audit
  
  on:
    schedule:
      - cron: '0 0 * * 0'  # 每周日
    workflow_dispatch:
  
  jobs:
    audit:
      runs-on: ubuntu-latest
      steps:
        - name: Checkout code
          uses: actions/checkout@v4
        
        - name: Set up Go
          uses: actions/setup-go@v5
          with:
            go-version: '1.26'
        
        - name: Check for outdated dependencies
          run: |
            go list -u -m all
            go mod tidy
        
        - name: Check Go toolchain compatibility
          run: bash scripts/check-dependencies.sh
        
        - name: Create issue if problems found
          if: failure()
          uses: actions/github-script@v7
          with:
            script: |
              github.rest.issues.create({
                owner: context.repo.owner,
                repo: context.repo.repo,
                title: 'Dependency audit failed',
                body: 'Please review the failed dependency audit.',
                labels: ['dependencies', 'automated']
              })
  ```

#### 验证标准
- ✅ 依赖检查脚本创建
- ✅ CI Workflow 包含依赖检查
- ✅ 定期审计工作流创建
- ✅ 本地测试脚本工作正常

---

### 阶段 3: 改进测试策略（中优先级）

#### 目标
添加集成测试，提高测试覆盖率，确保端到端功能正常。

#### 实施步骤

**步骤 3.1: 创建集成测试目录**
- 目录: `internal/integration/`
- 文件: `internal/integration/orchestrator_integration_test.go`
- 内容:
  ```go
  package integration
  
  import (
      "context"
      "testing"
      "time"
      
      "github.com/axis-cli/axis/internal/kernel/orchestrator"
      "github.com/axis-cli/axis/internal/types"
  )
  
  func TestOrchestratorIntegration(t *testing.T) {
      ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
      defer cancel()
      
      orch := orchestrator.NewOrchestrator()
      err := orch.Start(ctx)
      if err != nil {
          t.Fatalf("Failed to start orchestrator: %v", err)
      }
      
      contract := &types.AgentContract{
          ContractID: "test-contract",
          InputSchema: &types.InputSchema{
              Fields: []types.FieldDef{
                  {
                      Name:     "name",
                      Type:     types.FieldTypeString,
                      Required: true,
                  },
              },
          },
      }
      err = orch.RegisterContract(contract)
      if err != nil {
          t.Fatalf("Failed to register contract: %v", err)
      }
      
      task := &types.AgentTask{
          TaskID:     "task-1",
          ContractID: "test-contract",
          Input:      map[string]any{"name": "test"},
      }
      err = orch.SubmitTask(task)
      if err != nil {
          t.Fatalf("Failed to submit task: %v", err)
      }
      
      // Wait for task completion
      time.Sleep(1 * time.Second)
      
      status := orch.GetTaskStatus("task-1")
      if status != types.TaskStatusCompleted && status != types.TaskStatusFailed {
          t.Errorf("Task status should be completed or failed, got %s", status)
      }
      
      orch.Shutdown(ctx)
  }
  ```

**步骤 3.2: 更新 CI Workflow**
- 文件: `.github/workflows/ci.yml`
- 在 test job 后添加:
  ```yaml
  integration-test:
    name: Integration Test
    runs-on: ubuntu-latest
    needs: test
    steps:
      - name: Checkout code
        uses: actions/checkout@v4
      
      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.26'
      
      - name: Download dependencies
        run: go mod download
      
      - name: Run integration tests
        run: go test -v -tags=integration ./internal/integration/...
  ```

**步骤 3.3: 提高测试覆盖率目标**
- 文件: `.github/workflows/ci.yml`
- 更新覆盖率检查:
  ```yaml
  - name: Check coverage
    run: |
      COVERAGE=$(go tool cover -func=coverage.out | grep total | awk '{print $3}' | sed 's/%//')
      echo "Total coverage: $COVERAGE%"
      if (( $(echo "$COVERAGE < 70" | bc -l) )); then
        echo "Coverage is below 70%"
        exit 1
      fi
  ```

#### 验证标准
- ✅ 集成测试创建
- ✅ CI Workflow 包含集成测试
- ✅ 覆盖率目标提高到 70%
- ✅ 集成测试通过

---

### 阶段 4: 工作流定期审查机制（中优先级）

#### 目标
建立工作流定期审查机制，确保工作流保持最新和有效。

#### 实施步骤

**步骤 4.1: 创建工作流审计脚本**
- 文件: `scripts/audit-workflows.sh`
- 内容:
  ```bash
  #!/bin/bash
  # Audit GitHub Actions workflows
  
  echo "Auditing GitHub Actions workflows..."
  
  # Check workflow syntax
  for workflow in .github/workflows/*.yml; do
      echo "Checking $workflow..."
      yamllint "$workflow" || exit 1
  done
  
  # Check for deprecated actions
  echo "Checking for deprecated GitHub Actions..."
  # This would require a tool or API call to check action versions
  
  # Check workflow registry consistency
  echo "Checking workflow registry..."
  # Verify all workflows in registry exist
  
  echo "Workflow audit completed"
  ```

**步骤 4.2: 创建定期审计工作流**
- 文件: `.github/workflows/workflow-audit.yml`
- 内容:
  ```yaml
  name: Workflow Audit
  
  on:
    schedule:
      - cron: '0 0 * * 0'  # 每周日
    workflow_dispatch:
  
  jobs:
    audit:
      runs-on: ubuntu-latest
      steps:
        - name: Checkout code
          uses: actions/checkout@v4
        
        - name: Install yamllint
          run: pip install yamllint
        
        - name: Run workflow audit
          run: bash scripts/audit-workflows.sh
  ```

**步骤 4.3: 更新工作流注册表**
- 文件: `.github/workflows/registry.yml`
- 添加审计工作流元数据

#### 验证标准
- ✅ 审计脚本创建
- ✅ 定期审计工作流创建
- ✅ 工作流注册表更新
- ✅ 审计脚本工作正常

---

### 阶段 5: 错误追踪和通知（低优先级）

#### 目标
添加 Slack 或邮件通知，及时获知工作流失败。

#### 实施步骤

**步骤 5.1: 配置 Slack 通知**
- 文件: `.github/workflows/ci.yml`
- 在每个 job 后添加:
  ```yaml
  - name: Notify Slack on failure
    if: failure()
    uses: slackapi/slack-github-action@v1
    with:
      payload: |
        {
          "text": "CI Workflow failed: ${{ github.workflow }}",
          "blocks": [
            {
              "type": "section",
              "text": {
                "type": "mrkdwn",
                "text": "CI Workflow failed: ${{ github.workflow }}\nRepo: ${{ github.repository }}\nBranch: ${{ github.ref }}\nCommit: ${{ github.sha }}"
              }
            }
          ]
        }
    env:
      SLACK_WEBHOOK_URL: ${{ secrets.SLACK_WEBHOOK_URL }}
  ```

**步骤 5.2: 配置邮件通知**
- 文件: `.github/workflows/ci.yml`
- 添加邮件通知步骤:
  ```yaml
  - name: Send email notification
    if: failure()
    uses: dawidd6/action-send-mail@v3
    with:
      server_address: smtp.gmail.com
      server_port: 465
      username: ${{ secrets.EMAIL_USERNAME }}
      password: ${{ secrets.EMAIL_PASSWORD }}
      subject: CI Workflow failed
      to: ${{ secrets.NOTIFICATION_EMAIL }}
      from: GitHub Actions
      body: CI Workflow failed for ${{ github.repository }}
  ```

#### 验证标准
- ✅ Slack 通知配置
- ✅ 邮件通知配置
- ✅ Secrets 配置完成
- ✅ 通知测试成功

---

## 实施时间表

### 第 1 周（里程碑1完成前）
- ✅ 阶段 1: Pre-commit Hook
- ✅ 阶段 2: 依赖兼容性检查

### 第 2 周（里程碑1完成后）
- ⏳ 阶段 3: 改进测试策略
- ⏳ 阶段 4: 工作流定期审查机制

### 里程碑2（DAG 并行调度）
- ⏳ 阶段 5: 错误追踪和通知

---

## 验证计划

### 阶段 1 验证
- [ ] 本地安装 pre-commit 并测试
- [ ] 提交代码触发 pre-commit hook
- [ ] 验证 Dev Workflow 通过
- [ ] 更新文档并验证

### 阶段 2 验证
- [ ] 本地运行依赖检查脚本
- [ ] 提交代码触发 CI Workflow
- [ ] 验证依赖检查通过
- [ ] 手动触发定期审计工作流

### 阶段 3 验证
- [ ] 创建集成测试
- [ ] 本地运行集成测试
- [ ] CI Workflow 运行集成测试
- [ ] 验证覆盖率提升

### 阶段 4 验证
- [ ] 本地运行审计脚本
- [ ] 手动触发审计工作流
- [ ] 验证审计结果
- [ ] 更新工作流注册表

### 阶段 5 验证
- [ ] 配置 Slack Webhook
- [ ] 配置邮件 Secrets
- [ ] 触发失败测试通知
- [ ] 验证通知接收

---

## 风险和缓解措施

### 风险 1: Pre-commit hook 影响开发速度
- **缓解**: 配置为可选，提供跳过选项
- **缓解**: 优化 hook 性能，只运行必要检查

### 风险 2: 依赖检查产生误报
- **缓解**: 允许配置白名单
- **缓解**: 人工审查审计结果

### 风险 3: 集成测试不稳定
- **缓解**: 使用 mock 和 stub
- **缓解**: 设置合理的超时时间

### 风险 4: 通知过于频繁
- **缓解**: 配置通知级别
- **缓解**: 合并相似通知

---

## 成功标准

### 量化指标
- Pre-commit hook 覆盖率: 100%
- CI 失败率降低: > 50%
- 测试覆盖率提升: 60% → 70%
- 工作流审计通过率: 100%

### 质量指标
- 本地发现问题比例: > 80%
- 重复提交减少: > 70%
- 废弃参数使用: 0
- 工作流维护周期: 每周

---

## 下一步行动

1. **立即开始**: 实施阶段 1（Pre-commit Hook）
2. **本周完成**: 实施阶段 2（依赖兼容性检查）
3. **里程碑1后**: 实施阶段 3-4
4. **里程碑2**: 实施阶段 5

---

**文档版本**: 1.0
**最后更新**: 2026-05-08
**负责人**: Dev Team
**审核人**: Tech Lead

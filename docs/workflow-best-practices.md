# 工作流最佳实践

## 概述

本文档记录了在 Axis 项目中开发和维护 GitHub Actions 工作流的最佳实践，基于实际经验总结而成。

## 1. 工作流触发设计

### 1.1 精确的路径过滤
使用路径过滤避免不必要的工作流执行：

```yaml
on:
  push:
    paths:
      - '**.go'           # 仅 Go 文件变更时触发
      - 'go.mod'
      - 'go.sum'
      - '.github/config/registry.yml'
```

### 1.2 事件类型组合
正确组合不同事件类型：

```yaml
on:
  push:
    branches: [ main, develop ]
  pull_request:
    branches: [ main, develop ]
  workflow_run:
    workflows: [CI, PR Quality Check]
    types: [completed]
```

## 2. 条件执行模式

### 2.1 事件类型检查模式
在需要访问特定事件属性时，先检查事件类型：

```yaml
jobs:
  validate-registry:
    if: |
      github.event_name == 'push' && contains(github.event.head_commit.modified, '.github/config/registry.yml') ||
      github.event_name == 'pull_request' && contains(github.event.pull_request.changed_files, '.github/config/registry.yml')
```

### 2.2 工作流运行条件
仅在特定条件下运行工作流：

```yaml
on:
  workflow_run:
    workflows: [CI]
    types: [completed]

jobs:
  metrics:
    if: github.event_name == 'workflow_run' && github.event.workflow_run.conclusion == 'success'
```

## 3. 数据验证模式

### 3.1 文件存在性检查
处理可能不存在的文件：

```yaml
- name: Check if benchmark results exist
  id: check-benchmark
  run: |
    if [ -s benchmark.txt ] && grep -q "Benchmark" benchmark.txt; then
      echo "has_benchmark=true" >> $GITHUB_OUTPUT
    else
      echo "has_benchmark=false" >> $GITHUB_OUTPUT
    fi

- name: Process benchmark
  if: steps.check-benchmark.outputs.has_benchmark == 'true'
  run: process_benchmark
```

### 3.2 数据完整性验证
验证数据格式和内容：

```python
# Python 验证示例
with open('data.json', 'r') as f:
    data = json.load(f)
    if not isinstance(data, dict):
        print("❌ Invalid data format")
        sys.exit(1)
    if 'required_field' not in data:
        print("❌ Missing required field")
        sys.exit(1)
```

## 4. 错误处理策略

### 4.1 继续执行模式
对非关键步骤使用 continue-on-error：

```yaml
- name: Optional check
  id: optional-check
  continue-on-error: true
  run: some-check

- name: Handle failure
  if: steps.optional-check.outcome == 'failure'
  run: echo "Optional check failed, continuing..."
```

### 4.2 依赖失败处理
使用 needs 条件处理依赖失败：

```yaml
jobs:
  job1:
    runs-on: ubuntu-latest
  job2:
    runs-on: ubuntu-latest
    needs: job1
    if: needs.job1.result == 'success'
```

### 4.3 总是运行摘要
使用 always() 确保摘要步骤总是运行：

```yaml
jobs:
  summary:
    runs-on: ubuntu-latest
    needs: [job1, job2]
    if: always()
    steps:
      - name: Generate summary
        run: echo "Summary of results"
```

## 5. 上下文变量使用

### 5.1 常用上下文变量
```yaml
- name: Show context
  run: |
    echo "Event: ${{ github.event_name }}"
    echo "Branch: ${{ github.ref_name }}"
    echo "Actor: ${{ github.actor }}"
    echo "Base ref: ${{ github.base_ref }}"
    echo "Head ref: ${{ github.head_ref }}"
```

### 5.2 PR 特定变量
在 PR 工作流中使用 PR 特定变量：

```yaml
- name: PR checks
  if: github.event_name == 'pull_request'
  run: |
    echo "PR number: ${{ github.event.pull_request.number }}"
    echo "Base branch: ${{ github.base_ref }}"
    echo "Head branch: ${{ github.head_ref }}"
    git diff --stat origin/${{ github.base_ref }}...HEAD
```

## 6. 脚本编写模式

### 6.1 Python heredoc 模式
正确使用 Python heredoc：

```yaml
- name: Python script
  run: |
    python3 << 'EOF'
    import sys
    try:
        # Python code here
        print("result=value")
    except Exception as e:
        print(f"Error: {e}")
        sys.exit(1)
    EOF

    echo "Processing result: $result"
```

### 6.2 JavaScript 可选链模式
在 github-script 中使用可选链：

```javascript
const workflowId = context.event?.workflow_run?.workflow_id;
if (!workflowId) {
  console.log('No workflow_run event data available');
  return;
}
```

### 6.3 subprocess 异常处理
在 Python 脚本中处理 subprocess 异常：

```python
try:
    result = subprocess.run(['git', 'command'], capture_output=True, text=True, check=True)
except subprocess.CalledProcessError as e:
    print(f"❌ Error: {e}")
    sys.exit(1)
```

## 7. 权限管理

### 7.1 最小权限原则
只授予工作流所需的最小权限：

```yaml
permissions:
  contents: read
  pull-requests: read
  issues: write  # 仅在需要创建 issue 时
```

### 7.2 写入权限配置
需要写入权限时明确说明：

```yaml
# 注意：此步骤需要 contents: write 权限
# 在 Repository Settings > Actions > General 中配置
- name: Commit and push
  if: github.ref == 'refs/heads/main'
  run: |
    git config --local user.email "action@github.com"
    git config --local user.name "GitHub Action"
    git add file
    git commit -m "message"
    git push
```

## 8. 缓存策略

### 8.1 依赖缓存
缓存 Go 模块和构建产物：

```yaml
- name: Cache Go modules
  uses: actions/cache@v4
  with:
    path: |
      ~/.cache/go-build
      ~/go/pkg/mod
    key: ${{ runner.os }}-go-${{ hashFiles('**/go.sum') }}
    restore-keys: |
      ${{ runner.os }}-go-
```

### 8.2 构建缓存
缓存构建产物以加速后续构建：

```yaml
- name: Cache build
  uses: actions/cache@v4
  with:
    path: build/
    key: ${{ runner.os }}-build-${{ github.sha }}
```

## 9. 工作流组织

### 9.1 职责分离
将不同职责分离到不同的 job：

```yaml
jobs:
  validate:
    # 数据验证
  build:
    needs: validate
    # 构建
  test:
    needs: build
    # 测试
  deploy:
    needs: test
    # 部署
```

### 9.2 矩阵策略
使用矩阵策略测试多个配置：

```yaml
strategy:
  matrix:
    os: [ubuntu-latest, windows-latest, macos-latest]
    go-version: ['1.26', '1.27']
```

## 10. 监控和可观测性

### 10.1 步骤摘要
使用 GITHUB_STEP_SUMMARY 输出摘要：

```yaml
- name: Generate summary
  run: |
    echo "# Test Results" >> $GITHUB_STEP_SUMMARY
    echo "- Total tests: 100" >> $GITHUB_STEP_SUMMARY
    echo "- Passed: 95" >> $GITHUB_STEP_SUMMARY
    echo "- Failed: 5" >> $GITHUB_STEP_SUMMARY
```

### 10.2 Artifact 上传
上传构建产物和测试结果：

```yaml
- name: Upload artifacts
  uses: actions/upload-artifact@v4
  with:
    name: test-results
    path: test-results/
    retention-days: 30
```

### 10.3 失败通知
在关键失败时发送通知：

```yaml
- name: Notify on failure
  if: failure()
  uses: actions/github-script@v7
  with:
    script: |
      github.rest.issues.create({
        owner: context.repo.owner,
        repo: context.repo.repo,
        title: 'Workflow failed',
        body: 'Please check the workflow logs.',
        labels: ['workflow-failure']
      })
```

## 11. 性能优化

### 11.1 并行执行
使用并行执行加速工作流：

```yaml
jobs:
  job1:
    runs-on: ubuntu-latest
  job2:
    runs-on: ubuntu-latest
  job3:
    runs-on: ubuntu-latest
    needs: [job1, job2]  # 并行执行 job1 和 job2
```

### 11.2 条件跳过
跳过不必要的步骤：

```yaml
- name: Skip if no changes
  id: check-changes
  run: |
    if git diff --quiet; then
      echo "changed=false" >> $GITHUB_OUTPUT
    else
      echo "changed=true" >> $GITHUB_OUTPUT
    fi

- name: Process changes
  if: steps.check-changes.outputs.changed == 'true'
  run: process_changes
```

## 12. 安全实践

### 12.1 Secret 使用
正确使用 secrets：

```yaml
- name: Use secret
  env:
    API_KEY: ${{ secrets.API_KEY }}
  run: echo "Using API key"
```

### 12.2 依赖扫描
定期扫描依赖漏洞：

```yaml
on:
  schedule:
    - cron: '0 0 * * 0'  # 每周日

jobs:
  security:
    runs-on: ubuntu-latest
    steps:
      - name: Run govulncheck
        run: govulncheck ./...
```

## 13. 调试技巧

### 13.1 启用调试日志
在需要调试时启用 debug 日志：

```yaml
- name: Enable debug logging
  run: |
    echo "::debug::Detailed debug information"
    echo "::warning::Warning message"
    echo "::error::Error message"
```

### 13.2 保留构建产物
在失败时保留构建产物：

```yaml
- name: Upload build artifacts on failure
  if: failure()
  uses: actions/upload-artifact@v4
  with:
    name: build-artifacts
    path: build/
```

## 14. 工作流文档化

### 14.1 添加注释
在工作流中添加清晰的注释：

```yaml
# Validate registry.yml structure and file references
# Triggered on push/PR when registry.yml changes
- name: Validate registry
  run: validate_registry
```

### 14.2 更新交接文档
工作流变更后更新 HANDOVER.md

## 15. 测试工作流

### 15.1 本地测试
使用 act 工具本地测试工作流：

```bash
act push -j validate-registry
```

### 15.2 干运行
在提交前进行干运行检查：

```bash
# 检查工作流语法
# 检查路径过滤是否正确
# 检查条件逻辑
```

## 总结

遵循这些最佳实践可以：
- 提高工作流可靠性
- 减少调试时间
- 改善可维护性
- 增强安全性
- 优化性能

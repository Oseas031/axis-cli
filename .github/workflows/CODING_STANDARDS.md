# GitHub Actions 工作流编写规范

## 1. 事件属性访问规范

### 1.1 事件类型检查
在访问 GitHub Actions 事件属性前，必须先检查事件类型：

```yaml
if: |
  github.event_name == 'push' && contains(github.event.head_commit.modified, 'file.yml') ||
  github.event_name == 'pull_request' && contains(github.event.pull_request.changed_files, 'file.yml')
```

**原因**：不同事件类型的事件对象结构不同，push 事件没有 `pull_request.changed_files`，直接访问会导致错误。

### 1.2 JavaScript 可选链操作符
在 github-script 中使用可选链操作符访问可能不存在的属性：

```javascript
const workflowId = context.event?.workflow_run?.workflow_id;
if (!workflowId) {
  console.log('No workflow_run event data available');
  return;
}
```

**原因**：防止访问未定义属性导致 JavaScript 崩溃。

## 2. Python 脚本编写规范

### 2.1 字典访问安全
在 Python 中访问字典键前必须检查键是否存在：

```python
# 错误做法
file = workflow['file']  # KeyError if key doesn't exist

# 正确做法
if 'file' in workflow:
    file = workflow['file']
else:
    # 处理缺失情况
```

### 2.2 subprocess 异常处理
所有 subprocess.run 调用必须添加异常处理：

```python
# 错误做法
result = subprocess.run(['git', 'command'], capture_output=True, text=True)

# 正确做法
try:
    result = subprocess.run(['git', 'command'], capture_output=True, text=True, check=True)
except subprocess.CalledProcessError as e:
    print(f"❌ Error running command: {e}")
    sys.exit(1)
```

### 2.3 Python heredoc 与 bash 分离
Python heredoc 中不应混用 bash 命令：

```yaml
# 错误做法
run: |
  python3 << 'EOF'
  print("data")
  echo "bash command" >> $GITHUB_STEP_SUMMARY  # 语法错误
  EOF

# 正确做法
run: |
  python3 << 'EOF'
  print(f"key={value}")
  EOF
  echo "bash command" >> $GITHUB_STEP_SUMMARY
```

## 3. 数据验证规范

### 3.1 文件内容检查
处理外部命令输出前检查文件是否为空或包含预期内容：

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

### 3.2 Go 模块数据处理
使用 Go 工具或 jq 处理 Go 模块数据，避免不可靠的 grep 模式匹配：

```bash
# 错误做法
go list -m all | grep "^direct" > direct-deps.txt  # 不会匹配到任何内容

# 正确做法
go list -m -json all | jq -r 'select(.Indirect == false) | .Path' | wc -l
```

## 4. 文件组织规范

### 4.1 数据文件位置
数据文件不应放在 `.github/workflows/` 目录：

- ✅ 正确位置：`.github/config/`, `.github/`, `configs/`
- ❌ 错误位置：`.github/workflows/`

**原因**：GitHub Actions 会尝试解析 `.github/workflows/` 目录下所有 `.yml` 文件为工作流，数据文件缺少必需字段会导致解析失败。

### 4.2 配置文件组织
配置文件应放在专门的配置目录：

```
.github/
  ├── workflows/        # GitHub Actions 工作流文件
  └── config/          # 配置文件（如 registry.yml）
```

## 5. Git 操作规范

### 5.1 分支引用
使用 GitHub Actions 提供的上下文变量，避免硬编码分支名称：

```yaml
# 错误做法
git diff --stat origin/main...HEAD

# 正确做法
git diff --stat origin/${{ github.base_ref }}...HEAD
```

### 5.2 Git 推送认证
自动推送需要配置 GitHub Actions bot 权限：

```yaml
# 需要 GitHub Actions bot 写入权限
- name: Commit and push
  run: |
    git config --local user.email "action@github.com"
    git config --local user.name "GitHub Action"
    git add file
    git commit -m "message"
    git push origin HEAD:${{ github.ref }}
```

**注意**：在 PR 中自动推送会被拒绝，应在 workflow_dispatch 或 main 分支上启用。

## 6. 工具选择规范

### 6.1 避免功能重复
避免安装功能重叠的工具：

```yaml
# 错误做法：nancy 和 govulncheck 功能重复
- name: Run nancy
  run: go install github.com/sonatypecommunity/nancy/tardigrade/tardigrade@latest
- name: Run govulncheck
  run: govulncheck ./...

# 正确做法：只使用 govulncheck
- name: Run govulncheck
  run: govulncheck ./...
```

### 6.2 工具功能说明
在工作流中添加工具功能说明注释：

```yaml
# govulncheck covers dependency vulnerability checking
# nancy tool removed due to Git authentication issues
- name: Check for known vulnerabilities
  run: govulncheck ./...
```

## 7. 文档更新规范

### 7.1 及时更新文档
每次重大变更后必须更新交接文档：

- Bug 修复必须包含：修复时间、原因、解决方案
- 工作流变更必须更新 HANDOVER.md
- 新增工具必须添加功能说明

### 7.2 文档检查
文档审查工作流应检查交接文档是否及时更新。

## 8. 状态管理规范

### 8.1 状态一致性
确保状态检查和设置逻辑一致：

```go
// 错误做法
if !o.running {
    return fmt.Errorf("already shut down")
}
o.running = false  // 逻辑反了

// 正确做法
if o.running {
    return fmt.Errorf("already running")
}
o.running = true
```

### 8.2 资源清理
在 Shutdown 时添加资源清理通知：

```go
func (o *Orchestrator) Shutdown(ctx context.Context) error {
    o.mu.Lock()
    o.running = false
    o.mu.Unlock()

    // 通知任务循环停止
    select {
    case o.taskSubmitted <- struct{}{}:
    default:
    }

    return o.lifecycleManager.Shutdown(ctx)
}
```

## 9. 条件执行规范

### 9.1 前置条件检查
对可能失败的操作添加前置条件检查：

```yaml
- name: Process data
  if: steps.check-data.outputs.exists == 'true'
  run: process_data
```

### 9.2 错误处理
所有可能失败的步骤都应添加错误处理：

```yaml
- name: Run command
  id: run-command
  continue-on-error: true
  run: some-command

- name: Check result
  if: steps.run-command.outcome == 'failure'
  run: handle_error
```

## 10. 安全规范

### 10.1 敏感信息
不要在工作流中硬编码敏感信息：

```yaml
# 错误做法
env:
  API_KEY: "hardcoded-key"

# 正确做法
env:
  API_KEY: ${{ secrets.API_KEY }}
```

### 10.2 权限最小化
工作流权限应遵循最小权限原则：

```yaml
permissions:
  contents: read
  pull-requests: read
```

仅在需要时授予写入权限。

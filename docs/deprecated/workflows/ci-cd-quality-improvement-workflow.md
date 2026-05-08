---
description: CI/CD质量改进与关键Bug修复工作流
---

# CI/CD质量改进与关键Bug修复工作流

此工作流记录了修复代码审查中发现的关键问题、设置CI/CD流水线以及确保Axis CLI项目代码质量标准的过程。

## 使用场景

在以下情况下使用此工作流：
- 修复代码审查中发现的关键问题
- 设置或改进CI/CD流水线
- 解决测试覆盖率失败问题
- 解决静态分析工具错误
- 确保代码在发布前符合质量标准

## 工作流步骤

### 1. 修复关键代码问题

**调度器GetNextTask Bug**
- 文件：`internal/kernel/scheduler/scheduler.go`
- 问题：GetNextTask返回任务但未标记为已分发
- 修复：返回时标记任务为running并更新状态存储
- 影响：防止重复执行任务

**分发器Goroutine泄漏**
- 文件：`internal/kernel/dispatcher/dispatcher.go`
- 问题：Goroutine在context超时后继续运行
- 修复：在goroutine中添加context检查再发送到channel
- 影响：防止goroutine泄漏和资源耗尽

**编排器忙等待模式**
- 文件：`internal/kernel/orchestrator/orchestrator.go`
- 问题：在循环中使用time.Sleep，CPU使用效率低
- 修复：实现基于channel的任务通知机制
- 影响：提高效率，减少CPU使用

**契约执行器线程安全**
- 文件：`internal/contract/executor/executor.go`
- 问题：contracts map访问无同步保护
- 修复：添加sync.RWMutex保护并发访问
- 影响：防止生产环境中的竞态条件

**CLI空指针解引用**
- 文件：`cmd/axis/main.go`
- 问题：全局orch变量仅在start命令中初始化
- 修复：在run和status命令中添加nil检查
- 影响：防止在start之前调用命令时panic

### 2. 设置CI/CD流水线

**创建CI工作流**
- 文件：`.github/workflows/ci.yml`
- 任务：format、vet、lint、test、build
- 依赖：format → vet → lint → test → build
- 覆盖率阈值：60%
- 多平台构建产物

**创建CD工作流**
- 文件：`.github/workflows/release.yml`
- 触发：git tag推送
- 多平台二进制构建
- SHA256校验和生成
- GitHub Release自动化

**配置静态分析**
- 文件：`.golangci.yml`
- 使用staticcheck以兼容Go 1.26
- 启用必要的linter
- 设置复杂度和安全规则

### 3. 修复代码质量问题

**格式化失败**
```bash
gofmt -s -w <文件>
```
- 格式化所有Go文件
- 确保代码风格一致

**Lint兼容性问题**
- golangci-lint与Go 1.26不兼容
- 解决方案：改用staticcheck
- 安装：`go install honnef.co/go/tools/cmd/staticcheck@latest`

**循环依赖检测**
- 修复调度器中的DFS算法
- 检查依赖前标记当前任务为已访问
- 正确清理已访问状态

### 4. 提高测试覆盖率

**添加缺失的测试**
- `internal/kernel/lifecycle/lifecycle_test.go`
- `internal/kernel/dispatcher/dispatcher_test.go`
- `internal/human/executor/executor_test.go`
- 增强 `internal/contract/executor/executor_test.go`

**覆盖率目标**
- 目标：总体覆盖率≥60%
- 聚焦关键路径
- 测试边界情况和错误条件

**修复静态检查错误**
- 在nil检查后使用`t.Fatal`
- 防止nil指针解引用警告
- 确保测试安全性

### 5. Git工作流

**提交策略**
```bash
git add <文件>
git commit -m "fix: <描述>"
git push
```

**用户偏好**
- 用户处理git push操作
- Agent处理代码更改和git add/commit
- 协作式代码审查工作流

### 6. 验证CI/CD

**检查Actions**
- 访问 https://github.com/<用户>/<仓库>/actions
- 监控任务执行
- 验证所有检查通过

**覆盖率报告**
- 确保总体覆盖率≥60%
- 检查各模块覆盖率
- 解决低覆盖率模块

## 常见问题与解决方案

### golangci-lint版本不兼容
**错误**：Go版本与golangci-lint不匹配
**解决方案**：使用staticcheck替代，对Go 1.26支持更好

### 测试覆盖率低于阈值
**错误**：覆盖率<60%
**解决方案**：为未测试的模块添加全面测试

### 格式化失败
**错误**：gofmt发现未格式化文件
**解决方案**：对所有文件运行`gofmt -s -w .`

### 静态检查Nil指针
**错误**：SA5011可能的nil指针解引用
**解决方案**：在nil检查后使用`t.Fatal`代替`t.Error`

## 成功标准

- 所有CI/CD任务通过
- 测试覆盖率≥60%
- 无静态分析错误
- 所有关键bug已修复
- 代码已正确格式化
- 多平台构建成功

## 文档更新

工作流完成后更新以下内容：
- README.md - 添加CI/CD章节
- docs/QUICKSTART.md - 简化并聚焦核心功能
- docs/ROADMAP.md - 跟踪进度

## 相关文件

- `.github/workflows/ci.yml` - CI流水线配置
- `.github/workflows/release.yml` - CD流水线配置
- `.golangci.yml` - 静态分析配置
- `go.mod` - Go模块依赖
- `README.md` - 项目文档

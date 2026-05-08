# Claude Code 工作流无缝衔接指南

**目的**: 确保不同 Claude Code 实例之间能够无缝衔接工作流，继承工作进度

---

## 核心原则

1. **文档优先** - 所有工作进度必须记录在文档中
2. **标准化流程** - 使用标准化的工作流程和文档结构
3. **状态可见性** - 工作状态必须在关键位置可见
4. **可追溯性** - 每个决策和变更都必须可追溯

---

## 工作流衔接机制

### 1. 文档化工作进度

#### 当前进度文档
- **位置**: `docs/current-progress.md`
- **内容**:
  ```markdown
  # 当前工作进度

  **更新时间**: YYYY-MM-DD HH:mm
  **当前里程碑**: Milestone 1

  ## 已完成任务
  - [x] 任务1 - 描述
  - [x] 任务2 - 描述

  ## 进行中任务
  - [ ] 任务3 - 描述
    - 子任务3.1
    - 子任务3.2

  ## 待处理任务
  - [ ] 任务4 - 描述

  ## 遇到的问题
  - 问题1 - 解决方案
  - 问题2 - 待解决

  ## 下一步行动
  1. 行动1
  2. 行动2
  ```

#### 每日复盘文档
- **位置**: `docs/daily-retrospective-YYYY-MM-DD.md`
- **内容**:
  - 今日完成的工作
  - 遇到的问题
  - 工作流不足
  - 下一步计划

#### 交接文档
- **位置**: `HANDOVER.md`
- **内容**:
  - 项目概述
  - 当前状态
  - 已完成工作
  - 待处理任务
  - 已知问题
  - 项目结构

---

### 2. 使用记忆系统

#### 全局规则记忆
- 创建全局规则记忆，包含：
  - 项目设计原则（奥卡姆剃刀）
  - 里程碑范围
  - 禁止行为
  - 文档阅读顺序

#### 项目上下文记忆
- 创建项目上下文记忆，包含：
  - 项目结构
  - 核心模块
  - 技术栈
  - 当前里程碑

#### 工作进度记忆
- 创建工作进度记忆，包含：
  - 最近完成的工作
  - 当前待处理任务
  - 已知问题和解决方案

---

### 3. 工作流注册表

#### 工作流元数据
- **位置**: `.github/workflows/registry.yml`
- **内容**:
  - 工作流 ID
  - 工作流名称
  - 版本信息
  - 状态
  - 依赖关系
  - 文件路径
  - 文档路径

#### 工作流状态追踪
- 每个工作流包含：
  - 成功率
  - 平均执行时间
  - 最后执行时间
  - 最后执行结果

---

### 4. 标准化工作流程

#### 接手工作流
1. **读取关键文档**（按顺序）:
   - HANDOVER.md - 项目交接文档
   - docs/current-progress.md - 当前工作进度
   - docs/daily-retrospective-YYYY-MM-DD.md - 最新复盘
   - AGENT_INSTRUCTIONS.md - Agent 接手提示词

2. **检查工作流注册表**:
   - 查看 `.github/workflows/registry.yml`
   - 了解活跃工作流
   - 检查工作流状态

3. **检查 CI/CD 状态**:
   - 查看 GitHub Actions 运行历史
   - 检查最近的失败
   - 了解当前构建状态

4. **更新记忆系统**:
   - 加载项目上下文
   - 加载工作进度
   - 创建新的会话记忆

#### 交接工作流
1. **更新当前进度文档**:
   - 记录已完成工作
   - 记录进行中任务
   - 记录待处理任务
   - 记录遇到的问题

2. **创建每日复盘**:
   - 总结今日工作
   - 分析工作流不足
   - 制定改进计划

3. **更新交接文档**:
   - 更新 HANDOVER.md
   - 更新 AGENT_INSTRUCTIONS.md
   - 更新工作流注册表

4. **提交并推送**:
   - 提交所有变更
   - 推送到 GitHub
   - 确保 CI/CD 通过

---

## 实施方案

### 阶段 1: 创建当前进度文档

**文件**: `docs/current-progress.md`

```markdown
# 当前工作进度

**更新时间**: 2026-05-08 12:00
**当前里程碑**: Milestone 1（验收阶段）

## 已完成任务
- [x] 修复 staticcheck ST1003 错误（shared_layer → sharedlayer）
- [x] 修复契约执行器枚举验证逻辑（支持 int 类型）
- [x] 修复 CI 工作流 godoc -html 废弃参数
- [x] 创建工作流改进计划并修复高优先级问题
- [x] 文档审查和清理（移动 4 个过时文档到 deprecated）
- [x] 创建文档审查工作流（document-audit.yml）

## 进行中任务
- [ ] 观察文档审查工作流执行结果
- [ ] 等待 CI workflow 验证所有修复

## 待处理任务
- [ ] 创建 PR 触发 PR Quality Check 和 Security workflows
- [ ] 生成里程碑1验收报告
- [ ] 完成里程碑1验收

## 遇到的问题
- ✅ staticcheck ST1003 - 已修复
- ✅ godoc -html 废弃参数 - 已修复
- ✅ 枚举验证不支持 int 类型 - 已修复
- ✅ 文档过时问题 - 已清理

## 下一步行动
1. 观察文档审查工作流执行结果
2. 创建 PR 触发质量检查
3. 使用现有工作流完成里程碑1验收
4. 生成里程碑1验收报告

## 重要提醒
- 当前处于里程碑1验收阶段
- 不要提前实现里程碑2功能
- 遵循奥卡姆剃刀原则
- 使用现有工作流，不要创建新工作流
```

### 阶段 2: 更新 AGENT_INSTRUCTIONS.md

在 AGENT_INSTRUCTIONS.md 中添加：

```markdown
## Claude Code 工作流衔接

### 接手工作流
1. 读取 `docs/current-progress.md` 了解当前工作进度
2. 读取 `docs/daily-retrospective-YYYY-MM-DD.md` 了解最新复盘
3. 检查 `.github/workflows/registry.yml` 了解工作流状态
4. 检查 GitHub Actions 了解 CI/CD 状态
5. 更新记忆系统加载项目上下文

### 交接工作流
1. 更新 `docs/current-progress.md` 记录工作进度
2. 创建每日复盘文档
3. 更新 `HANDOVER.md` 和 `AGENT_INSTRUCTIONS.md`
4. 提交并推送所有变更
5. 确保 CI/CD 通过

### 重要文档
- `docs/current-progress.md` - 当前工作进度（必须首先阅读）
- `HANDOVER.md` - 项目交接文档
- `docs/daily-retrospective-YYYY-MM-DD.md` - 最新复盘
- `docs/document-audit-report.md` - 文档审查报告
```

### 阶段 3: 创建工作流状态追踪脚本

**文件**: `scripts/check-workflow-status.sh`

```bash
#!/bin/bash
# 检查工作流状态

echo "Checking workflow status..."

# 检查 GitHub Actions 最近运行
echo "Recent GitHub Actions runs:"
gh run list --limit 5

# 检查工作流注册表
echo "Workflow registry status:"
cat .github/workflows/registry.yml | grep -A 5 "status:"

# 检查当前进度
echo "Current progress:"
cat docs/current-progress.md
```

### 阶段 4: 创建自动化交接脚本

**文件**: `scripts/handover.sh`

```bash
#!/bin/bash
# 自动化交接脚本

echo "Starting handover process..."

# 1. 更新当前进度文档
echo "Updating current progress..."
# 交互式输入当前进度

# 2. 创建每日复盘
echo "Creating daily retrospective..."
DATE=$(date +%Y-%m-%d)
cat > docs/daily-retrospective-$DATE.md << EOF
# 每日复盘报告

**日期**: $DATE
**工作时长**: 
**主要成果**:
- 
**遇到的问题**:
- 
**工作流不足**:
- 
**下一步计划**:
- 
EOF

# 3. 更新 HANDOVER.md
echo "Updating HANDOVER.md..."
# 自动更新状态

# 4. 提交并推送
echo "Committing and pushing..."
git add docs/current-progress.md docs/daily-retrospective-$DATE.md HANDOVER.md
git commit -m "docs: update progress and handover ($DATE)"
git push origin main

echo "Handover completed successfully!"
```

---

## 最佳实践

### 1. 文档更新频率
- **当前进度文档**: 每次任务完成后更新
- **每日复盘**: 每天结束时创建
- **交接文档**: 里程碑转换时更新

### 2. 记忆系统使用
- **全局规则**: 项目开始时创建，重大变更时更新
- **项目上下文**: 项目开始时创建，架构变更时更新
- **工作进度**: 每次交接时更新

### 3. 工作流注册表维护
- 添加新工作流时立即更新
- 工作流状态变更时更新
- 定期审计工作流注册表

### 4. CI/CD 监控
- 每次交接前检查 CI/CD 状态
- 失败的工作流必须修复后才能交接
- 记录工作流失败原因和解决方案

---

## 验证清单

### 接手验证
- [ ] 已读取 docs/current-progress.md
- [ ] 已读取最新每日复盘
- [ ] 已检查工作流注册表
- [ ] 已检查 GitHub Actions 状态
- [ ] 已更新记忆系统
- [ ] 已理解当前里程碑范围
- [ ] 已理解奥卡姆剃刀原则

### 交接验证
- [ ] 已更新 docs/current-progress.md
- [ ] 已创建每日复盘
- [ ] 已更新 HANDOVER.md
- [ ] 已更新 AGENT_INSTRUCTIONS.md
- [ ] 已更新工作流注册表
- [ ] 已提交并推送
- [ ] CI/CD 通过
- [ ] 无遗留问题

---

## 故障排除

### 问题 1: 文档过时
**原因**: 没有及时更新文档
**解决方案**: 每次任务完成后立即更新文档

### 问题 2: 工作流状态不明确
**原因**: 工作流注册表未更新
**解决方案**: 添加/修改工作流时立即更新注册表

### 问题 3: 记忆系统不同步
**原因**: 没有更新记忆
**解决方案**: 每次交接时更新记忆系统

### 问题 4: CI/CD 失败未修复
**原因**: 忽略 CI/CD 失败
**解决方案**: CI/CD 失败必须修复后才能交接

---

## 总结

通过文档化、记忆系统、工作流注册表和标准化流程，可以实现 Claude Code 实例之间的无缝衔接。关键在于：

1. **文档优先** - 所有工作进度必须记录在文档中
2. **标准化流程** - 使用标准化的接手和交接流程
3. **状态可见性** - 工作状态必须在关键位置可见
4. **可追溯性** - 每个决策和变更都必须可追溯

这样可以确保任何 Claude Code 实例都能快速理解当前工作状态，无缝继承工作进度。

# Axis - Agent 原生调度系统

为 AI Agent 提供统一的任务调度能力。

## 核心能力

- 任务调度（FIFO + 依赖管理）
- 契约验证（输入输出 Schema）
- 状态管理
- 人类任务队列

## 里程碑1目标

- FIFO 任务调度
- 简单依赖管理
- 输入输出验证
- 基础状态存储
- 基础 CLI

## 进度

### 已完成 ✓
- ✅ 核心数据结构实现
- ✅ 状态存储模块
- ✅ 生命周期管理器
- ✅ 调度器（FIFO + 依赖管理）
- ✅ 契约执行器（输入输出验证）
- ✅ 人类执行器
- ✅ 分发器
- ✅ 编排器
- ✅ CLI 客户端
- ✅ 单元测试（覆盖率≥60%）
- ✅ CI/CD 流水线（format、vet、staticcheck、test、build）
- ✅ 持续交付（多平台构建 + 自动Release）
- ✅ 关键Bug修复（调度器、分发器、编排器、契约执行器、CLI）
- ✅ 代码质量改进（格式化、静态分析、覆盖率）

### 状态
- CI/CD：全部通过
- 测试覆盖率：≥60%
- 代码质量：format、vet、staticcheck 全部通过
- 多平台构建：成功

## 快速开始

### 安装
```bash
go build -o axis cmd/axis/main.go
```

### 使用
```bash
axis run my-task
```

## 文档

- [快速入门](docs/QUICKSTART.md)
- [系统架构可视化](docs/DIAGRAMS.md)
- [白皮书](docs/WHITEPAPER.md)
- [里程碑1检查清单](docs/milestones/milestone1-checklist.md)
- [项目演化路线图](docs/ROADMAP.md)

## CI/CD

- **持续集成**：代码质量检查、测试、多平台构建
- **持续交付**：基于git tag自动发布多平台二进制文件
- 触发方式：push到main/develop分支或创建PR

## 技术栈

- 语言：Go 1.26+
- 核心依赖：零外部依赖（仅 Go 标准库）
- 部署形态：单静态二进制文件

## 许可证

MIT License

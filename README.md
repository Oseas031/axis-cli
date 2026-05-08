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

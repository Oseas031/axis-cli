# Axis CLI 可选模块架构设计

## 模块铁律
可选模块与核心模块 100% 解耦，可选模块的迭代不得修改核心内核的 Agent 原生设计，不得破坏内核的调度语义与函数调用规范。

## 1. 人类兼容交互层 (internal/adapter/posix)

### 1.1 POSIX 适配器
**职责**：传统终端操作与 Agent 原生调度的语义兼容

**核心功能**：
- POSIX 命令解析与执行
- 传统 shell 语义映射到 Agent 调用
- 交互式命令行体验
- 历史命令记录与补全
- 信号处理（Ctrl+C 等）

**关键接口**：
```go
type POSIXAdapter interface {
    ParseCommand(cmd string) (*AgentTask, error)
    ExecutePOSIX(cmd string) (*POSIXResult, error)
    InteractiveLoop(ctx context.Context) error
}
```

**设计约束**：
- 仅做语义转换，不修改核心调度逻辑
- 可选模块，移除后不影响 Agent 原生能力
- 向后兼容传统 CLI 使用习惯

### 1.2 HTTP 适配器
**职责**：标准化 HTTP 调用接口

**核心功能**：
- RESTful API 暴露
- WebSocket 实时通信
- 请求鉴权与限流
- API 版本管理
- OpenAPI 规范生成

**关键接口**：
```go
type HTTPAdapter interface {
    StartServer(addr string) error
    StopServer(ctx context.Context) error
    RegisterRoute(method, path string, handler RouteHandler) error
}
```

### 1.3 SDK 适配器
**职责**：多语言 SDK 调用接口

**核心功能**：
- Go SDK
- Python SDK
- JavaScript/TypeScript SDK
- SDK 版本兼容管理
- SDK 文档生成

**设计约束**：
- SDK 仅封装 HTTP 调用
- 不包含业务逻辑
- 版本独立管理

## 2. AI 增强辅助模块 (pkg/llm)

### 2.1 LLM 提供商抽象层
**职责**：LLM 提供商无关的统一接口

**核心功能**：
- 多提供商支持（OpenAI、Anthropic、本地模型等）
- 统一调用接口
- 提供商配置管理
- 模型能力适配
- 成本与配额管理

**关键接口**：
```go
type LLMProvider interface {
    Complete(ctx context.Context, req *CompletionRequest) (*CompletionResponse, error)
    StreamComplete(ctx context.Context, req *CompletionRequest) (<-chan CompletionChunk, error)
    GetModelInfo(model string) (*ModelInfo, error)
}
```

### 2.2 提示词管理
**职责**：提示词模板与版本管理

**核心功能**：
- 提示词模板存储
- 变量插值与渲染
- 提示词版本控制
- A/B 测试支持
- 提示词效果追踪

### 2.3 上下文管理
**职责**：对话上下文与记忆管理

**核心功能**：
- 对话历史存储
- 上下文窗口管理
- 长期记忆检索
- 上下文压缩策略
- RAG 集成

## 3. 调试面板模块 (pkg/observability)

### 3.1 日志系统
**职责**：结构化日志与日志分级

**核心功能**：
- 结构化日志输出
- 日志级别控制
- 日志轮转与归档
- 日志过滤与搜索
- 多输出目标支持

**关键接口**：
```go
type Logger interface {
    Debug(msg string, fields ...Field)
    Info(msg string, fields ...Field)
    Warn(msg string, fields ...Field)
    Error(msg string, fields ...Field)
}
```

### 3.2 指标监控
**职责**：核心指标采集与暴露

**核心功能**：
- 调度指标（任务数、成功率、延迟等）
- 调用指标（工具调用、human 调用等）
- 资源指标（CPU、内存、Goroutine 等）
- Prometheus 格式暴露
- 自定义指标注册

**核心指标定义**：
- 详见 `docs/metrics/core-metrics.md`

### 3.3 链路追踪
**职责**：全链路调用追踪

**核心功能**：
- 分布式追踪（OpenTelemetry）
- 调用链可视化
- 性能瓶颈分析
- 错误根因定位
- 追踪数据导出

### 3.4 调试接口
**职责**：运行时调试与诊断

**核心功能**：
- pprof 性能分析
- 运行时状态查询
- 任务调试接口
- 配置热更新
- 健康检查端点

## 可选模块依赖关系

```
internal/adapter (适配层)
    ├── posix (POSIX 适配)
    ├── http (HTTP 适配)
    └── sdk (SDK 适配)

pkg/llm (AI 增强)
    ├── provider (提供商抽象)
    ├── prompt (提示词管理)
    └── context (上下文管理)

pkg/observability (观测性)
    ├── logger (日志)
    ├── metrics (指标)
    ├── trace (追踪)
    └── debug (调试)

依赖关系：
- posix → kernel.scheduler
- http → kernel.scheduler, kernel.dispatcher
- sdk → http
- llm → kernel.tools (作为工具注册)
- observability → 所有模块 (通过接口注入)
```

## 可选模块技术约束

1. **完全解耦**：可选模块移除后，核心内核仍可独立运行
2. **接口注入**：通过依赖注入与核心模块交互，不修改核心代码
3. **可插拔**：每个可选模块可独立启用/禁用
4. **向后兼容**：可选模块接口变更不影响已部署的核心内核
5. **性能隔离**：可选模块性能问题不得影响核心内核可用性
6. **配置驱动**：可选模块行为通过配置控制，不硬编码

## 可选模块启用策略

### 里程碑 1（自举起点）
- 启用：POSIX 适配器（基础兼容）
- 禁用：HTTP 适配器、SDK、AI 增强、调试面板

### 里程碑 2（自举核心）
- 启用：HTTP 适配器、SDK、AI 增强
- 禁用：调试面板（仅开发环境）

### 里程碑 3（自举完成）
- 启用：所有模块
- POSIX 适配器降级为最小调试入口

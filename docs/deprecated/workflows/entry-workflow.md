---
description: 入口工作流 - Agent 工作流调用统一入口
---

# 入口工作流（Entry Workflow）

## 问题背景

随着工作流数量增多（目前规划 7+ 个工作流），Agent 难以精准调用特定工作流。需要一个统一的入口点，让 Agent 通过简单的接口调用任何工作流。

## 解决方案

创建入口工作流作为工作流调度器（Workflow Dispatcher），提供统一的调用接口，根据输入参数路由到相应的工作流。

---

## 入口工作流设计

### 工作流名称
`workflow-entry` - 工作流入口点

### 触发条件
- Manual dispatch（手动触发）
- API 调用
- 其他工作流调用

### 输入参数
```yaml
workflow_type: 工作流类型（dev/pr-check/ci/security/cd/docs/monitoring）
action: 执行动作（run/validate/test/deploy/monitor/status）
params: 工作流特定参数（可选）
```

### 输出
```yaml
workflow_id: 被调用的工作流 ID
status: 执行状态
result: 执行结果
logs: 执行日志
```

---

## 入口工作流实现

### GitHub Actions 配置

```yaml
# .github/workflows/workflow-entry.yml
name: Workflow Entry Point
on:
  workflow_dispatch:
    inputs:
      workflow_type:
        description: 'Workflow type to invoke'
        required: true
        type: choice
        options:
          - dev
          - pr-check
          - ci
          - security
          - cd
          - docs
          - monitoring
      action:
        description: 'Action to perform'
        required: true
        type: choice
        options:
          - run
          - validate
          - test
          - deploy
          - monitor
          - status
      params:
        description: 'Additional parameters (JSON)'
        required: false
        type: string
        default: '{}'

jobs:
  dispatch:
    name: Dispatch to Workflow
    runs-on: ubuntu-latest
    outputs:
      workflow_id: ${{ steps.dispatch.outputs.workflow_id }}
      status: ${{ steps.dispatch.outputs.status }}
      result: ${{ steps.dispatch.outputs.result }}
    steps:
      - name: Checkout
        uses: actions/checkout@v4

      - name: Load workflow registry
        id: load-registry
        run: |
          # Load workflow registry
          WORKFLOW_INFO=$(yq ".workflows[] | select(.id == \"wf-${{ inputs.workflow_type }}\")" .github/workflows/registry.yml)
          echo "workflow_info=$WORKFLOW_INFO" >> $GITHUB_OUTPUT

      - name: Validate workflow exists
        run: |
          if [ -z "${{ steps.load-registry.outputs.workflow_info }}" ]; then
            echo "❌ Workflow type ${{ inputs.workflow_type }} not found"
            exit 1
          fi

      - name: Parse parameters
        id: parse-params
        run: |
          echo "params_json=${{ inputs.params }}" >> $GITHUB_OUTPUT

      - name: Dispatch to target workflow
        id: dispatch
        run: |
          case "${{ inputs.workflow_type }}" in
            dev)
              echo "Dispatching to development workflow"
              # Trigger development workflow
              gh workflow run dev-workflow.yml -f action=${{ inputs.action }} -f params="${{ inputs.params }}"
              echo "workflow_id=wf-dev" >> $GITHUB_OUTPUT
              echo "status=dispatched" >> $GITHUB_OUTPUT
              echo "result=Development workflow triggered" >> $GITHUB_OUTPUT
              ;;
            pr-check)
              echo "Dispatching to PR check workflow"
              # Trigger PR check workflow
              gh workflow run pr-check-workflow.yml -f action=${{ inputs.action }} -f params="${{ inputs.params }}"
              echo "workflow_id=wf-pr-check" >> $GITHUB_OUTPUT
              echo "status=dispatched" >> $GITHUB_OUTPUT
              echo "result=PR check workflow triggered" >> $GITHUB_OUTPUT
              ;;
            ci)
              echo "Dispatching to CI workflow"
              # Trigger CI workflow
              gh workflow run ci-workflow.yml -f action=${{ inputs.action }} -f params="${{ inputs.params }}"
              echo "workflow_id=wf-ci" >> $GITHUB_OUTPUT
              echo "status=dispatched" >> $GITHUB_OUTPUT
              echo "result=CI workflow triggered" >> $GITHUB_OUTPUT
              ;;
            security)
              echo "Dispatching to security workflow"
              # Trigger security workflow
              gh workflow run security-workflow.yml -f action=${{ inputs.action }} -f params="${{ inputs.params }}"
              echo "workflow_id=wf-security" >> $GITHUB_OUTPUT
              echo "status=dispatched" >> $GITHUB_OUTPUT
              echo "result=Security workflow triggered" >> $GITHUB_OUTPUT
              ;;
            cd)
              echo "Dispatching to CD workflow"
              # Trigger CD workflow
              gh workflow run cd-workflow.yml -f action=${{ inputs.action }} -f params="${{ inputs.params }}"
              echo "workflow_id=wf-cd" >> $GITHUB_OUTPUT
              echo "status=dispatched" >> $GITHUB_OUTPUT
              echo "result=CD workflow triggered" >> $GITHUB_OUTPUT
              ;;
            docs)
              echo "Dispatching to documentation workflow"
              # Trigger documentation workflow
              gh workflow run docs-workflow.yml -f action=${{ inputs.action }} -f params="${{ inputs.params }}"
              echo "workflow_id=wf-docs" >> $GITHUB_OUTPUT
              echo "status=dispatched" >> $GITHUB_OUTPUT
              echo "result=Documentation workflow triggered" >> $GITHUB_OUTPUT
              ;;
            monitoring)
              echo "Dispatching to monitoring workflow"
              # Trigger monitoring workflow
              gh workflow run monitoring-workflow.yml -f action=${{ inputs.action }} -f params="${{ inputs.params }}"
              echo "workflow_id=wf-monitoring" >> $GITHUB_OUTPUT
              echo "status=dispatched" >> $GITHUB_OUTPUT
              echo "result=Monitoring workflow triggered" >> $GITHUB_OUTPUT
              ;;
            *)
              echo "❌ Unknown workflow type: ${{ inputs.workflow_type }}"
              exit 1
              ;;
          esac

      - name: Log dispatch result
        run: |
          echo "Workflow ID: ${{ steps.dispatch.outputs.workflow_id }}"
          echo "Status: ${{ steps.dispatch.outputs.status }}"
          echo "Result: ${{ steps.dispatch.outputs.result }}"

      - name: Notify on failure
        if: failure()
        run: |
          echo "❌ Workflow dispatch failed"
          exit 1
```

---

## Agent 调用接口

### REST API 接口

```go
// internal/workflow/entry/entry.go
package entry

import (
	"context"
	"encoding/json"
)

// WorkflowEntryRequest 工作流入口请求
type WorkflowEntryRequest struct {
	WorkflowType string                 `json:"workflow_type"` // dev/pr-check/ci/security/cd/docs/monitoring
	Action       string                 `json:"action"`        // run/validate/test/deploy/monitor/status
	Params       map[string]interface{} `json:"params"`        // 工作流特定参数
}

// WorkflowEntryResponse 工作流入口响应
type WorkflowEntryResponse struct {
	WorkflowID string                 `json:"workflow_id"`
	Status     string                 `json:"status"`     // dispatched/completed/failed
	Result     map[string]interface{} `json:"result"`
	Logs       []string               `json:"logs"`
}

// WorkflowEntry 工作流入口接口
type WorkflowEntry interface {
	Dispatch(ctx context.Context, req *WorkflowEntryRequest) (*WorkflowEntryResponse, error)
	GetStatus(ctx context.Context, workflowID string) (string, error)
	ListWorkflows(ctx context.Context) ([]string, error)
}

// WorkflowEntryImpl 工作流入口实现
type WorkflowEntryImpl struct {
	registry *WorkflowRegistry
}

// NewWorkflowEntry 创建工作流入口
func NewWorkflowEntry(registry *WorkflowRegistry) *WorkflowEntryImpl {
	return &WorkflowEntryImpl{
		registry: registry,
	}
}

// Dispatch 调度工作流
func (e *WorkflowEntryImpl) Dispatch(ctx context.Context, req *WorkflowEntryRequest) (*WorkflowEntryResponse, error) {
	// 验证工作流类型
	if !e.registry.Exists(req.WorkflowType) {
		return nil, fmt.Errorf("workflow type %s not found", req.WorkflowType)
	}

	// 获取工作流配置
	workflowConfig := e.registry.Get(req.WorkflowType)

	// 验证动作
	if !workflowConfig.SupportsAction(req.Action) {
		return nil, fmt.Errorf("action %s not supported by workflow %s", req.Action, req.WorkflowType)
	}

	// 调度到目标工作流
	response, err := e.dispatchToWorkflow(ctx, workflowConfig, req)
	if err != nil {
		return nil, fmt.Errorf("dispatch failed: %w", err)
	}

	return response, nil
}

// dispatchToWorkflow 调度到具体工作流
func (e *WorkflowEntryImpl) dispatchToWorkflow(ctx context.Context, config *WorkflowConfig, req *WorkflowEntryRequest) (*WorkflowEntryResponse, error) {
	// 根据工作流类型调用相应的工作流
	switch config.Type {
	case "dev":
		return e.dispatchDevelopmentWorkflow(ctx, req)
	case "pr-check":
		return e.dispatchPRCheckWorkflow(ctx, req)
	case "ci":
		return e.dispatchCIWorkflow(ctx, req)
	case "security":
		return e.dispatchSecurityWorkflow(ctx, req)
	case "cd":
		return e.dispatchCDWorkflow(ctx, req)
	case "docs":
		return e.dispatchDocsWorkflow(ctx, req)
	case "monitoring":
		return e.dispatchMonitoringWorkflow(ctx, req)
	default:
		return nil, fmt.Errorf("unknown workflow type: %s", config.Type)
	}
}

// GetStatus 获取工作流状态
func (e *WorkflowEntryImpl) GetStatus(ctx context.Context, workflowID string) (string, error) {
	return e.registry.GetStatus(workflowID)
}

// ListWorkflows 列出所有可用工作流
func (e *WorkflowEntryImpl) ListWorkflows(ctx context.Context) ([]string, error) {
	return e.registry.List()
}
```

### CLI 接口

```go
// cmd/workflow/workflow.go
package workflow

import (
	"fmt"
	"github.com/spf13/cobra"
)

var workflowType string
var action string
var params string

// workflowCmd 工作流命令
var workflowCmd = &cobra.Command{
	Use:   "workflow",
	Short: "Workflow entry point",
	Long:  "Unified entry point for invoking workflows",
	RunE:  runWorkflow,
}

func init() {
	workflowCmd.Flags().StringVarP(&workflowType, "type", "t", "", "Workflow type (dev/pr-check/ci/security/cd/docs/monitoring)")
	workflowCmd.Flags().StringVarP(&action, "action", "a", "run", "Action (run/validate/test/deploy/monitor/status)")
	workflowCmd.Flags().StringVarP(&params, "params", "p", "{}", "Additional parameters (JSON)")
}

func runWorkflow(cmd *cobra.Command, args []string) error {
	if workflowType == "" {
		return fmt.Errorf("workflow type is required")
	}

	// 创建工作流入口
	entry := NewWorkflowEntry(registry)

	// 解析参数
	var paramsMap map[string]interface{}
	if err := json.Unmarshal([]byte(params), &paramsMap); err != nil {
		return fmt.Errorf("invalid params JSON: %w", err)
	}

	// 调度工作流
	req := &WorkflowEntryRequest{
		WorkflowType: workflowType,
		Action:       action,
		Params:       paramsMap,
	}

	response, err := entry.Dispatch(context.Background(), req)
	if err != nil {
		return fmt.Errorf("workflow dispatch failed: %w", err)
	}

	fmt.Printf("Workflow ID: %s\n", response.WorkflowID)
	fmt.Printf("Status: %s\n", response.Status)
	fmt.Printf("Result: %v\n", response.Result)

	return nil
}
```

---

## 工作流注册表

```go
// internal/workflow/registry/registry.go
package registry

import (
	"sync"
)

// WorkflowRegistry 工作流注册表
type WorkflowRegistry struct {
	mu        sync.RWMutex
	workflows map[string]*WorkflowConfig
}

// WorkflowConfig 工作流配置
type WorkflowConfig struct {
	ID          string
	Name        string
	Type        string
	Version     string
	Status      string
	File        string
	Actions     []string
	Dependencies []string
}

// NewWorkflowRegistry 创建工作流注册表
func NewWorkflowRegistry() *WorkflowRegistry {
	return &WorkflowRegistry{
		workflows: make(map[string]*WorkflowConfig),
	}
}

// Register 注册工作流
func (r *WorkflowRegistry) Register(config *WorkflowConfig) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.workflows[config.ID]; exists {
		return fmt.Errorf("workflow %s already registered", config.ID)
	}

	r.workflows[config.ID] = config
	return nil
}

// Get 获取工作流配置
func (r *WorkflowRegistry) Get(id string) *WorkflowConfig {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.workflows[id]
}

// Exists 检查工作流是否存在
func (r *WorkflowRegistry) Exists(id string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	_, exists := r.workflows[id]
	return exists
}

// List 列出所有工作流
func (r *WorkflowRegistry) List() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	ids := make([]string, 0, len(r.workflows))
	for id := range r.workflows {
		ids = append(ids, id)
	}
	return ids
}

// SupportsAction 检查工作流是否支持指定动作
func (c *WorkflowConfig) SupportsAction(action string) bool {
	for _, a := range c.Actions {
		if a == action {
			return true
		}
	}
	return false
}
```

---

## 使用示例

### Agent 调用示例

```go
// Agent 使用入口工作流
package main

import (
	"context"
	"fmt"
)

func main() {
	// 创建工作流入口
	entry := NewWorkflowEntry(registry)

	// 调用 PR 检查工作流
	req := &WorkflowEntryRequest{
		WorkflowType: "pr-check",
		Action:       "run",
		Params: map[string]interface{}{
			"pr_number": 123,
			"branch":    "feature/new-scheduler",
		},
	}

	response, err := entry.Dispatch(context.Background(), req)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Workflow dispatched: %s\n", response.WorkflowID)
}
```

### CLI 调用示例

```bash
# 调用开发工作流
axis workflow --type dev --action run

# 调用 PR 检查工作流
axis workflow --type pr-check --action run --params '{"pr_number": 123}'

# 调用 CI 工作流
axis workflow --type ci --action validate

# 列出所有工作流
axis workflow --action list
```

### GitHub Actions 调用示例

```yaml
# 在其他工作流中调用入口工作流
- name: Invoke workflow via entry point
  uses: ./.github/workflows/workflow-entry.yml
  with:
    workflow_type: ci
    action: run
    params: '{"branch": "main"}'
```

---

## 优势

1. **统一接口** - Agent 只需知道一个入口点
2. **类型安全** - 工作流类型和动作有明确的枚举
3. **可扩展** - 新增工作流只需注册到注册表
4. **可验证** - 入口工作流验证工作流存在性和动作支持
5. **可追踪** - 所有工作流调用都经过入口，便于监控
6. **简化调用** - Agent 不需要知道具体工作流文件名

---

## 工作流类型映射

| 工作流类型 | 工作流 ID | 支持的动作 | 说明 |
|-----------|-----------|-----------|------|
| dev | wf-dev | run, validate, status | 开发阶段自动化 |
| pr-check | wf-pr-check | run, validate, status | PR 质量检查 |
| ci | wf-ci | run, validate, test, status | 持续集成 |
| security | wf-security | run, validate, scan, status | 安全扫描 |
| cd | wf-cd | run, deploy, status | 持续交付 |
| docs | wf-docs | run, generate, status | 文档自动化 |
| monitoring | wf-monitoring | run, collect, status | 监控与可观测性 |

---

## 实施步骤

1. 创建工作流注册表结构
2. 实现工作流入口接口
3. 创建 GitHub Actions 入口工作流
4. 实现工作流调度逻辑
5. 添加 CLI 接口
6. 编写单元测试
7. 更新文档

---

## 与 Meta-Workflow 集成

入口工作流本身由 Meta-Workflow 管理：

```yaml
# .github/workflows/registry.yml
workflows:
  - id: wf-entry
    name: Workflow Entry Point
    version: 1.0.0
    status: active
    category: meta
    triggers:
      - workflow_dispatch
      - api
    dependencies: []
    dependents:
      - wf-dev
      - wf-pr-check
      - wf-ci
      - wf-security
      - wf-cd
      - wf-docs
      - wf-monitoring
    file: .github/workflows/workflow-entry.yml
    documentation: workflow/entry-workflow.md
    owner: @devops-team
    reviewers:
      - @tech-lead
    created_at: 2026-05-08
    updated_at: 2026-05-08
    metrics:
      success_rate: 0.99
      avg_duration: 5s
      last_execution: 2026-05-08T10:45:00Z
```

---

## 总结

入口工作流作为工作流系统的统一入口点，解决了工作流数量增多导致的调用困难问题。通过类型化的接口和注册表机制，Agent 可以简单、可靠地调用任何工作流，同时保持系统的可扩展性和可维护性。

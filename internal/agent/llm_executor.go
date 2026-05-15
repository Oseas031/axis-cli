// Package agent provides the LLM-driven agent executor.
package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/axis-cli/axis/internal/agent/judgement"
	"github.com/axis-cli/axis/internal/model/multiturn"
	"github.com/axis-cli/axis/internal/model/provider"
	"github.com/axis-cli/axis/internal/model/tool"
	"github.com/axis-cli/axis/internal/types"
)

// TerminationDecision represents the agent's decision after each turn.
type TerminationDecision int

const (
	// Continue means the agent should keep executing.
	Continue TerminationDecision = iota
	// Complete means the task is done successfully.
	Complete
	// Failed means the task cannot be completed.
	Failed
	// NeedHuman means the agent needs human intervention.
	NeedHuman
)

// TerminationFn decides whether the agent should stop after a turn.
// Called after each LLM response. If nil, defaults to "stop when no tool calls".
type TerminationFn func(history []types.ModelMessage, last *provider.ModelResponse) TerminationDecision

// HistoryCompactor compacts message history to manage context window.
type HistoryCompactor interface {
	Compact(ctx context.Context, history []types.ModelMessage) []types.ModelMessage
}

// noopCompactor is the default compactor that does nothing.
type noopCompactor struct{}

func (noopCompactor) Compact(_ context.Context, h []types.ModelMessage) []types.ModelMessage { return h }

// ToolTrace records a single tool invocation for observability.
type ToolTrace struct {
	Name     string         `json:"name"`
	Input    map[string]any `json:"input"`
	Output   string         `json:"output"`
	Error    string         `json:"error,omitempty"`
	Duration time.Duration  `json:"duration_ms"`
}

// EventEmitter allows the executor to emit progress events during execution.
type EventEmitter interface {
	Emit(taskID, eventType, message string)
}

// WithEventEmitter sets the event emitter for progress reporting.
func WithEventEmitter(em EventEmitter) LLMAgentOption {
	return func(e *LLMAgentExecutor) { e.emitter = em }
}

// LLMAgentExecutor implements AgentExecutor using a real LLM provider
// with a multi-turn tool-calling loop.
// LLMAgentExecutor implements AgentExecutor using a real LLM provider
// with a multi-turn tool-calling loop.
type LLMAgentExecutor struct {
	provider     provider.ModelProvider
	tools        *tool.Registry
	systemPrompt string
	maxIter      int
	terminate    TerminationFn
	compactor    HistoryCompactor
	maxErrors    int // circuit breaker threshold
	turnTimeout  time.Duration
	agentID      string
	emitter      EventEmitter
	postJudge    PostExecutionJudge
	memory       ExecutionMemory
	workingMem   *WorkingMemoryRecaller
	immediateMem *ImmediateMemoryAdapter
}

// LLMAgentOption configures an LLMAgentExecutor.
type LLMAgentOption func(*LLMAgentExecutor)

// WithSystemPrompt sets the system prompt.
func WithSystemPrompt(prompt string) LLMAgentOption {
	return func(e *LLMAgentExecutor) { e.systemPrompt = prompt }
}

// WithMaxIterations sets the iteration budget.
func WithMaxIterations(n int) LLMAgentOption {
	return func(e *LLMAgentExecutor) { e.maxIter = n }
}

// WithTerminationFn sets a custom termination function.
func WithTerminationFn(fn TerminationFn) LLMAgentOption {
	return func(e *LLMAgentExecutor) { e.terminate = fn }
}

// WithHistoryCompactor sets the history compaction strategy.
func WithHistoryCompactor(c HistoryCompactor) LLMAgentOption {
	return func(e *LLMAgentExecutor) { e.compactor = c }
}

// WithMaxErrors sets the circuit breaker threshold.
func WithMaxErrors(n int) LLMAgentOption {
	return func(e *LLMAgentExecutor) { e.maxErrors = n }
}

// WithTurnTimeout sets the default per-turn timeout.
// Can be overridden per-task via metadata "axis.turn_timeout" (e.g. "90s").
func WithTurnTimeout(d time.Duration) LLMAgentOption {
	return func(e *LLMAgentExecutor) { e.turnTimeout = d }
}

// WithAgentID sets the agent identity.
func WithAgentID(id string) LLMAgentOption {
	return func(e *LLMAgentExecutor) { e.agentID = id }
}

// PostExecutionJudge is called after execution to validate the result.
// If it returns a non-nil JudgementResult with Passed=false, the execution
// is marked as failed and the error is set.
type PostExecutionJudge interface {
	JudgeExecution(ctx context.Context, task *types.AgentTask, result *AgentExecutionResult) *judgement.JudgementResult
}

// WithPostJudge sets the post-execution judgement hook.
func WithPostJudge(j PostExecutionJudge) LLMAgentOption {
	return func(e *LLMAgentExecutor) { e.postJudge = j }
}

// ExecutionMemory allows the executor to store and recall execution lessons.
type ExecutionMemory interface {
	StoreLessson(taskID, lesson string) error
	RecallLessons(query string) []string
}

// WithMemory sets the execution memory for cross-task learning.
func WithMemory(m ExecutionMemory) LLMAgentOption {
	return func(e *LLMAgentExecutor) { e.memory = m }
}

// WithWorkingMemory sets the BM25 working memory recaller for context injection.
func WithWorkingMemory(r *WorkingMemoryRecaller) LLMAgentOption {
	return func(e *LLMAgentExecutor) { e.workingMem = r }
}

// WithImmediateMemory sets the immediate memory adapter for file-change awareness.
func WithImmediateMemory(a *ImmediateMemoryAdapter) LLMAgentOption {
	return func(e *LLMAgentExecutor) { e.immediateMem = a }
}

// NewLLMAgentExecutor creates a new LLM-driven agent executor.
func NewLLMAgentExecutor(p provider.ModelProvider, tools *tool.Registry, opts ...LLMAgentOption) *LLMAgentExecutor {
	e := &LLMAgentExecutor{
		provider:    p,
		tools:       tools,
		maxIter:     20,
		maxErrors:   5,
		turnTimeout: 45 * time.Second,
		compactor:   noopCompactor{},
		agentID:     "default",
	}
	for _, opt := range opts {
		opt(e)
	}
	return e
}

// GetAutonomyLevel returns the current autonomy level.
// v1: fixed at Execute level.
func (e *LLMAgentExecutor) GetAutonomyLevel() AutonomyLevel {
	return AutonomyLevelExecute
}

// SetToolRegistry injects the tool registry after construction.
// Used by orchestrator to wire the shared registry.
func (e *LLMAgentExecutor) SetToolRegistry(r *tool.Registry) {
	e.tools = r
}

func (e *LLMAgentExecutor) emit(taskID, eventType, message string) {
	if e.emitter != nil {
		e.emitter.Emit(taskID, eventType, message)
	}
}

// Execute runs the multi-turn LLM ↔ Tool loop with post-execution judgement.
// If judgement fails, retries once with feedback from the failure.
// Recalls past lessons before execution and stores new lessons on failure.
func (e *LLMAgentExecutor) Execute(ctx context.Context, req *AgentExecutionRequest) (*AgentExecutionResult, error) {
	// Recall past lessons relevant to this task
	priorLessons := e.recallLessons(req)

	result, err := e.executeOnce(ctx, req, priorLessons)
	if err != nil {
		return result, err
	}

	// No judge configured or execution already failed — return as-is
	if e.postJudge == nil || result.Error != "" {
		return result, nil
	}

	// Judge the result
	jr := e.postJudge.JudgeExecution(ctx, req.Task, result)
	if jr == nil {
		return result, nil
	}

	result.JudgementResult = jr
	if jr.Passed {
		e.emit(req.Task.TaskID, "judgement_passed", fmt.Sprintf("score=%.2f", jr.Score))
		return result, nil
	}

	// Judgement failed — store lesson and retry once with feedback after brief backoff
	e.emit(req.Task.TaskID, "judgement_failed", fmt.Sprintf("score=%.2f: %s — retrying in 5s", jr.Score, jr.Summary()))
	e.storeLesson(req.Task.TaskID, jr.Summary())
	time.Sleep(5 * time.Second)

	feedback := fmt.Sprintf("Your previous attempt was rejected by quality check: %s. Please try again and address this issue.", jr.Summary())
	retryResult, retryErr := e.executeOnce(ctx, req, feedback)
	if retryErr != nil {
		return retryResult, retryErr
	}
	if retryResult.Error != "" {
		return retryResult, nil
	}

	// Judge the retry
	jr2 := e.postJudge.JudgeExecution(ctx, req.Task, retryResult)
	if jr2 == nil {
		e.emit(req.Task.TaskID, "retry_completed", "retry succeeded (no judgement)")
		return retryResult, nil
	}

	retryResult.JudgementResult = jr2
	if jr2.Passed {
		e.emit(req.Task.TaskID, "judgement_passed", fmt.Sprintf("score=%.2f (after retry)", jr2.Score))
	} else {
		retryResult.Error = fmt.Sprintf("self-judgement failed after retry (score=%.2f): %s", jr2.Score, jr2.Summary())
		e.emit(req.Task.TaskID, "judgement_failed_final", retryResult.Error)
		e.storeLesson(req.Task.TaskID, "retry also failed: "+jr2.Summary())
		// Include structured judgement details in output for user visibility
		if retryResult.Output == nil {
			retryResult.Output = make(map[string]any)
		}
		retryResult.Output["_judgement"] = map[string]any{
			"passed":          false,
			"score":           jr2.Score,
			"confidence":      jr2.Confidence,
			"failed_criteria": jr2.Summary(),
			"suggested_fixes": jr2.SuggestedFixes,
		}
	}
	return retryResult, nil
}

// executeOnce runs a single execution attempt. If feedback is non-empty,
// it's prepended to the system prompt as correction guidance.
func (e *LLMAgentExecutor) executeOnce(ctx context.Context, req *AgentExecutionRequest, feedback string) (*AgentExecutionResult, error) {
	var traces []ToolTrace
	emittedStart := false

	tp := &terminationProvider{
		inner:     e.provider,
		terminate: e.terminate,
		onFirstCall: func() {
			if !emittedStart {
				emittedStart = true
				e.emit(req.Task.TaskID, "execution_started", "Agent 正在思考...")
			}
		},
	}

	sysPrompt := e.systemPrompt
	if feedback != "" {
		sysPrompt = sysPrompt + "\n\n[CORRECTION] " + feedback
	}

	if e.workingMem != nil {
		query := ""
		if req.Task != nil && req.Task.Input != nil {
			if msg, ok := req.Task.Input["message"].(string); ok {
				query = msg
			}
		}
		if wmCtx := e.workingMem.Recall(ctx, query); wmCtx != "" {
			sysPrompt = sysPrompt + "\n\n" + wmCtx
		}
	}

	if e.immediateMem != nil {
		taskID := ""
		intent := ""
		if req.Task != nil {
			taskID = req.Task.TaskID
			if req.Task.Input != nil {
				if msg, ok := req.Task.Input["message"].(string); ok {
					intent = msg
				}
			}
		}
		if imCtx := e.immediateMem.BuildSituationalContext(ctx, taskID, intent); imCtx != "" {
			sysPrompt = sysPrompt + "\n\n" + imCtx
		}
	}

	modelReq := &provider.ModelRequest{
		ContractID:   req.Task.ContractID,
		Input:        req.Task.Input,
		Tools:        e.tools.List(),
		SystemPrompt: sysPrompt,
	}

	cfg := multiturn.LoopConfig{
		Provider:      tp,
		Tools:         e.tools,
		MaxIterations: e.maxIter,
		MaxErrors:     e.maxErrors,
		Compactor:     e.compactor,
		TurnTimeout:   e.resolveTurnTimeout(req),
		OnToolExecuted: func(toolName string, result map[string]any, err error) {
			trace := ToolTrace{Name: toolName}
			if err != nil {
				trace.Error = err.Error()
			} else {
				content, _ := json.Marshal(result)
				trace.Output = string(content)
				e.emit(req.Task.TaskID, "tool_executed", fmt.Sprintf("[%s] 完成", toolName))
			}
			traces = append(traces, trace)
		},
	}

	loopRes, err := multiturn.Run(ctx, cfg, modelReq)

	if tp.decision == Failed {
		return e.buildResult(loopRes.Output, traces, "agent declared failure"), nil
	}
	if tp.decision == NeedHuman {
		return e.buildResult(loopRes.Output, traces, "agent needs human intervention"), nil
	}
	if err != nil {
		return e.buildResult(nil, traces, loopRes.Error), err
	}
	if loopRes.Error != "" {
		return e.buildResult(loopRes.Output, traces, loopRes.Error), nil
	}

	return e.buildResult(loopRes.Output, traces, ""), nil
}

// terminationProvider wraps a ModelProvider to handle TerminationFn logic.
// When the LLM returns no tool calls and TerminationFn says Continue,
// it appends the assistant message to history and re-calls the inner provider.
type terminationProvider struct {
	inner       provider.ModelProvider
	terminate   TerminationFn
	onFirstCall func()
	decision    TerminationDecision
	firstDone   bool
}

func (tp *terminationProvider) Execute(ctx context.Context, req *provider.ModelRequest) (*provider.ModelResponse, error) {
	if !tp.firstDone {
		tp.firstDone = true
		if tp.onFirstCall != nil {
			tp.onFirstCall()
		}
	}

	const maxContinueRetries = 3
	retries := 0
	for {
		resp, err := tp.inner.Execute(ctx, req)
		if err != nil {
			return nil, err
		}

		// If there are tool calls, pass through directly
		if len(resp.ToolCalls) > 0 {
			return resp, nil
		}

		// Check termination
		if tp.terminate == nil {
			return resp, nil
		}

		decision := tp.terminate(req.History, resp)
		tp.decision = decision
		switch decision {
		case Continue:
			retries++
			if retries >= maxContinueRetries {
				// Prevent infinite loop: treat as Complete after max retries
				tp.decision = Complete
				return resp, nil
			}
			// Append assistant message and retry
			if resp.Output != nil {
				content, _ := json.Marshal(resp.Output)
				req.History = append(req.History, types.ModelMessage{Role: "assistant", Content: string(content)})
			}
			continue
		case Failed, NeedHuman:
			return resp, nil
		default: // Complete
			return resp, nil
		}
	}
}

// buildResult constructs an AgentExecutionResult with trace and identity.
// recallLessons retrieves past failure lessons relevant to the current task.
func (e *LLMAgentExecutor) recallLessons(req *AgentExecutionRequest) string {
	if e.memory == nil {
		return ""
	}
	query := ""
	if req.Task != nil && req.Task.Input != nil {
		if msg, ok := req.Task.Input["message"].(string); ok {
			query = msg
		}
	}
	if query == "" {
		return ""
	}
	lessons := e.memory.RecallLessons(query)
	if len(lessons) == 0 {
		return ""
	}
	return "[PAST LESSONS]\n- " + strings.Join(lessons, "\n- ")
}

// storeLesson persists a failure lesson for future recall.
func (e *LLMAgentExecutor) storeLesson(taskID, lesson string) {
	if e.memory == nil {
		return
	}
	_ = e.memory.StoreLessson(taskID, lesson)
}

// resolveTurnTimeout determines the per-turn timeout for a task.
// Priority: task metadata "axis.turn_timeout" > executor default.
func (e *LLMAgentExecutor) resolveTurnTimeout(req *AgentExecutionRequest) time.Duration {
	if req.Task != nil && req.Task.Metadata != nil {
		if raw, ok := req.Task.Metadata["axis.turn_timeout"]; ok && raw != "" {
			if d, err := time.ParseDuration(raw); err == nil && d > 0 {
				return d
			}
		}
	}
	return e.turnTimeout
}

// buildResult constructs an AgentExecutionResult with trace and identity.
func (e *LLMAgentExecutor) buildResult(output map[string]any, traces []ToolTrace, errMsg string) *AgentExecutionResult {
	result := &AgentExecutionResult{
		Output:  output,
		AgentID: e.agentID,
		Error:   errMsg,
	}
	// Store traces in output metadata for observability
	if len(traces) > 0 && result.Output == nil {
		result.Output = make(map[string]any)
	}
	if len(traces) > 0 {
		result.Output["_tool_traces"] = traces
	}
	return result
}

package dispatcher

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/axis-cli/axis/internal/types"
)

// HumanOutcome is the result of a resolved human request. It mirrors the
// gate package's Outcome so the kernel stays decoupled from the inbox
// implementation (docs/specs/human-agent-scheduling/ design.md D2).
type HumanOutcome struct {
	Degraded bool
	Approved bool
	Choice   string
	Text     string
	Output   map[string]any
}

// HumanGate parks a task on a persistent human request and returns when the
// human responds, the request degrades, or the caller context is cancelled.
// Implementations must be safe for concurrent use.
type HumanGate interface {
	Request(task *types.AgentTask, req *types.HumanCallRequest) (*HumanOutcome, error)
	// TaskWaiting reports whether the task currently parks on a request.
	// While true, dispatcher timeout enforcement yields to the human SLA.
	TaskWaiting(taskID string) bool
}

// humanWaitRecheckInterval bounds how long timeout enforcement suspends while
// a task is parked; the gate owns the real deadline. Not a task timeout — a
// bookkeeping cadence only (CLAUDE.md §10: named constant, not magic number).
const humanWaitRecheckInterval = 30 * time.Second

// Sentinel errors surfaced through TaskResult.Error for scriptable handling.
var (
	ErrHumanGateUnavailable = errors.New("human inbox gate not configured")
	ErrPermissionDenied     = errors.New("permission denied by autonomy policy")
)

// SetHumanGate injects the inbox-backed gate. Without it, tasks that require
// human confirmation fail closed — they are never silently auto-approved.
func (d *DispatcherImpl) SetHumanGate(g HumanGate) {
	d.auditMu.Lock()
	defer d.auditMu.Unlock()
	d.humanGate = g
}

// SetStatusObserver receives intermediate status transitions (e.g.
// waiting_human) so the orchestrator can surface them via GetTaskStatus.
func (d *DispatcherImpl) SetStatusObserver(fn func(taskID string, status types.TaskStatus)) {
	d.auditMu.Lock()
	defer d.auditMu.Unlock()
	d.statusObserver = fn
}

func (d *DispatcherImpl) notifyStatus(taskID string, status types.TaskStatus) {
	if fn := d.statusObserver; fn != nil {
		fn(taskID, status)
	}
}

// executeViaHumanInbox parks the task on a persistent info-kind request and
// maps the outcome to a TaskResult. It replaces the legacy busy-poll path and
// plumbs the human's answer into Result.Output (previously dropped).
func (d *DispatcherImpl) executeViaHumanInbox(ctx context.Context, task *types.AgentTask) (*types.TaskResult, error) {
	if d.humanGate == nil {
		msg := fmt.Sprintf("[HUMAN_GATE_UNAVAILABLE] %s: task %s", ErrHumanGateUnavailable, task.TaskID)
		return humanFailResult(task, msg), errors.New(msg)
	}

	req := &types.HumanCallRequest{
		TaskID:   task.TaskID,
		Kind:     types.HumanKindInfo,
		Prompt:   promptFromInput(task),
		Input:    task.Input,
		Metadata: task.Metadata,
	}

	d.notifyStatus(task.TaskID, types.TaskStatusWaitingHuman)
	outcome, reqErr := d.humanGate.Request(task, req)
	d.notifyStatus(task.TaskID, types.TaskStatusRunning)

	if reqErr != nil {
		if errors.Is(reqErr, context.Canceled) || errors.Is(reqErr, context.DeadlineExceeded) {
			return humanFailResult(task, reqErr.Error()), reqErr
		}
		msg := "[HUMAN_GATE_ERROR] " + reqErr.Error()
		return humanFailResult(task, msg), errors.New(msg)
	}
	if outcome == nil {
		msg := fmt.Sprintf("human gate returned nil outcome for task %s", task.TaskID)
		return humanFailResult(task, msg), errors.New(msg)
	}
	if outcome.Degraded {
		// Fail-closed: an unanswered human never becomes an approval.
		msg := fmt.Sprintf("[HUMAN_WAIT_DEGRADED] human wait degraded fail-closed for task %s", task.TaskID)
		return humanFailResult(task, msg), errors.New(msg)
	}
	if !outcome.Approved {
		msg := fmt.Sprintf("[HUMAN_DECLINED] human declined task %s", task.TaskID)
		return humanFailResult(task, msg), errors.New(msg)
	}

	output := map[string]any{}
	for k, v := range outcome.Output {
		output[k] = v
	}
	return &types.TaskResult{
		TaskID:    task.TaskID,
		Output:    output,
		Status:    types.TaskStatusCompleted,
		Completed: time.Now(),
	}, nil
}

// requestDecision asks the human to confirm an Ask-permission action before
// execution proceeds. Degraded outcomes deny by default (Ask never upgrades).
func (d *DispatcherImpl) requestDecision(task *types.AgentTask, autonomy string) error {
	if d.humanGate == nil {
		return fmt.Errorf("%w: ask permission requires a configured human gate", ErrHumanGateUnavailable)
	}
	req := &types.HumanCallRequest{
		TaskID:  task.TaskID,
		Kind:    types.HumanKindDecision,
		Prompt:  fmt.Sprintf("confirm execution of task %s (autonomy=%s)", task.TaskID, autonomy),
		Options: []string{"proceed", "abort"},
		Context: map[string]any{"contract_id": task.ContractID, "autonomy": autonomy},
	}
	outcome, err := d.humanGate.Request(task, req)
	if err != nil {
		return err
	}
	if outcome.Degraded {
		return fmt.Errorf("%w: confirmation wait degraded for task %s (fail-closed)", ErrPermissionDenied, task.TaskID)
	}
	if !outcome.Approved || outcome.Choice != "proceed" {
		return fmt.Errorf("%w: human chose %q for task %s", ErrPermissionDenied, outcome.Choice, task.TaskID)
	}
	return nil
}

func promptFromInput(task *types.AgentTask) string {
	if task.Input != nil {
		if v, ok := task.Input["prompt"].(string); ok && v != "" {
			return v
		}
		if v, ok := task.Input["goal"].(string); ok && v != "" {
			return "review goal: " + v
		}
	}
	return "human action required for task " + task.TaskID
}

func humanFailResult(task *types.AgentTask, msg string) *types.TaskResult {
	return &types.TaskResult{
		TaskID:    task.TaskID,
		Status:    types.TaskStatusFailed,
		Error:     msg,
		Completed: time.Now(),
	}
}

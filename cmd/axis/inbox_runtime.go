package main

import (
	"context"
	"fmt"
	"path/filepath"
	"sync"
	"time"

	"github.com/axis-cli/axis/internal/comm"
	"github.com/axis-cli/axis/internal/control"
	humangate "github.com/axis-cli/axis/internal/human/gate"
	"github.com/axis-cli/axis/internal/kernel/dispatcher"
	"github.com/axis-cli/axis/internal/types"
)

// runtimeProbeTimeout bounds health probes against a stale runtime.json record.
const runtimeProbeTimeout = 800 * time.Millisecond

// humanSched holds the runtime-scoped scheduling components behind a mutex:
// setup, use and teardown run on different goroutines (runtime goroutine vs
// command goroutine), so bare package globals would race.
type humanSchedState struct {
	mu    sync.Mutex
	inbox *comm.Inbox
	gate  *humangate.Gate
}

var humanSched humanSchedState

func setHumanSched(inbox *comm.Inbox, gate *humangate.Gate) {
	humanSched.mu.Lock()
	defer humanSched.mu.Unlock()
	humanSched.inbox = inbox
	humanSched.gate = gate
}

func getHumanGate() *humangate.Gate {
	humanSched.mu.Lock()
	defer humanSched.mu.Unlock()
	return humanSched.gate
}

func closeAndClearHumanSched() {
	humanSched.mu.Lock()
	g := humanSched.gate
	humanSched.gate = nil
	humanSched.inbox = nil
	humanSched.mu.Unlock()
	if g != nil {
		g.Close()
	}
}

// gateBridge adapts *humangate.Gate to dispatcher.HumanGate, keeping the
// kernel decoupled from the gate package (design.md D2).
type gateBridge struct{ g *humangate.Gate }

func (b gateBridge) Request(task *types.AgentTask, req *types.HumanCallRequest) (*dispatcher.HumanOutcome, error) {
	out, err := b.g.Request(task, req)
	if err != nil {
		return nil, err
	}
	return &dispatcher.HumanOutcome{
		Degraded: out.Degraded,
		Approved: out.Approved,
		Choice:   out.Choice,
		Text:     out.Text,
		Output:   out.Output,
	}, nil
}

func (b gateBridge) TaskWaiting(taskID string) bool { return b.g.TaskWaiting(taskID) }

// ensureNoLiveRuntime refuses to start when another healthy runtime already
// owns this project root (design.md D2 single-writer discipline). Stale
// locator records whose process is gone are ignored.
func ensureNoLiveRuntime(root string) error {
	locator := control.NewRuntimeLocator(root)
	record, err := locator.Load()
	if err != nil {
		return nil // no prior runtime recorded
	}
	client := control.NewClient(locator, control.LocalHTTPClient())
	ctx, cancel := context.WithTimeout(context.Background(), runtimeProbeTimeout)
	defer cancel()
	if _, herr := client.Health(ctx); herr == nil {
		return fmt.Errorf("runtime already running (pid %d at %s): stop it before starting another on this project", record.PID, record.Address)
	}
	return nil
}

// setupHumanScheduling constructs the shared inbox and gate for this runtime.
// Event emission reuses eventLog; TaskEventLog.Append is concurrency-safe.
func setupHumanScheduling(root string, eventLog *control.TaskEventLog) {
	inbox := comm.NewInbox(filepath.Join(root, ".axis", "comm"))
	emit := func(eventType, taskID, message string) {
		_ = eventLog.Append(control.TaskEvent{
			TaskID:    taskID,
			EventType: eventType,
			Actor:     "axis-runtime",
			Message:   message,
		})
	}
	setHumanSched(inbox, humangate.New(inbox, humangate.DefaultConfig(), emit))
}

// degradeExpiredRequests runs the startup scan (design.md D3): pending
// requests whose deadline passed while the runtime was down degrade now,
// fail-closed, and leave an audit trail.
func degradeExpiredRequests(out interface{ Write([]byte) (int, error) }) {
	g := getHumanGate()
	if g == nil {
		return
	}
	n, err := g.DegradeExpired()
	if err != nil {
		fmt.Fprintf(out, "Warning: expired-request scan failed: %v\n", err)
		return
	}
	if n > 0 {
		fmt.Fprintf(out, "Degraded %d expired human request(s) fail-closed\n", n)
	}
}

// inboxAPIBridge adapts the gate to control.InboxAPI (design.md D7).
type inboxAPIBridge struct{ g *humangate.Gate }

func (b inboxAPIBridge) List() ([]comm.InboxRecord, int, error) {
	return b.g.List()
}

func (b inboxAPIBridge) RespondCall(callID, choice, text string) (*comm.InboxRecord, error) {
	return b.g.RespondCall(callID, choice, text)
}

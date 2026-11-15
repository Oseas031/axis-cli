// Package gate implements the human-agent scheduling gate defined by
// docs/specs/human-agent-scheduling/: it turns AgentTask pauses into persistent
// inbox requests, parks the caller until a human responds, and enforces the
// symmetric SLA (nudge then degrade — never auto-approve).
//
// v1 semantics (deliberate, see design.md D2/D3):
//   - the caller's goroutine parks (worker slot occupied); there is no
//     execution-state serialization, so waits do NOT survive restarts;
//   - nudge records are audit entries + inbox aging signals, not push notifications;
//   - degrade is fail-closed: an unanswered request never becomes approval.
package gate

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/axis-cli/axis/internal/comm"
	"github.com/axis-cli/axis/internal/storeutil"
	"github.com/axis-cli/axis/internal/types"
)

// ActorID is the canonical human actor mailbox name.
const ActorID = "human"

// SLA defaults; overridable via Config so nothing is hardcoded at call sites
// (CLAUDE.md §10). Production values are deliberately conservative.
const (
	DefaultWaitTimeout   = 24 * time.Hour
	DefaultNudgeInterval = 1 * time.Hour
	DefaultMaxNudges     = 3
	DefaultMaxWaiting    = 2
)

// Event names appended to the task event log (FR8).
const (
	EventRequestCreated  = "human.request.created"
	EventRequestResolved = "human.request.resolved"
	EventRequestNudged   = "human.request.nudged"
	EventRequestDegraded = "human.request.degraded"
)

var (
	// ErrTooManyWaiting rejects a new request while MaxWaiting are pending,
	// protecting the worker pool from starvation by unanswered requests.
	ErrTooManyWaiting = errors.New("gate: too many pending human requests")

	// ErrGateClosed is returned to parked callers when the runtime shuts down.
	ErrGateClosed = errors.New("gate: runtime is shutting down")
)

// Config carries the tunable SLA parameters (design.md D3).
type Config struct {
	WaitTimeout   time.Duration // overall human wait budget per request
	NudgeInterval time.Duration // audit reminder cadence
	MaxNudges     int           // degrade early once reached, even before WaitTimeout
	MaxWaiting    int           // concurrent pending requests allowed
}

// DefaultConfig returns the production defaults.
func DefaultConfig() Config {
	return Config{
		WaitTimeout:   DefaultWaitTimeout,
		NudgeInterval: DefaultNudgeInterval,
		MaxNudges:     DefaultMaxNudges,
		MaxWaiting:    DefaultMaxWaiting,
	}
}

// Clock abstracts time for deterministic SLA tests.
type Clock interface {
	Now() time.Time
}

// realClock uses the wall clock.
type realClock struct{}

func (realClock) Now() time.Time { return time.Now().UTC() }

// Emitter appends an event to the task event log. EventType is free-form;
// gate emits the human.request.* family.
type Emitter func(eventType, taskID, message string)

// Outcome is what the parked caller receives when the wait ends.
type Outcome struct {
	Degraded bool
	Approved bool
	Choice   string
	Text     string
	Output   map[string]any
}

// Gate owns pending human requests for one project runtime.
type Gate struct {
	inbox *comm.Inbox
	cfg   Config
	clock Clock
	emit  Emitter
	nowFn func() time.Time // alias of clock for readability in helpers

	mu        sync.Mutex
	waiters   map[string]chan Outcome // callID -> parked waiter
	taskWaits map[string]string       // taskID -> callID currently parked
	closed    bool
	closeCh   chan struct{}
	closeOnce sync.Once
}

// New constructs a Gate over the given inbox. A nil emitter disables event
// logging (tests only); production wiring always provides one.
func New(inbox *comm.Inbox, cfg Config, emit Emitter) *Gate {
	if cfg.WaitTimeout <= 0 {
		cfg.WaitTimeout = DefaultWaitTimeout
	}
	if cfg.NudgeInterval <= 0 {
		cfg.NudgeInterval = DefaultNudgeInterval
	}
	if cfg.MaxNudges < 0 {
		cfg.MaxNudges = DefaultMaxNudges
	}
	if cfg.MaxWaiting <= 0 {
		cfg.MaxWaiting = DefaultMaxWaiting
	}
	var clk Clock = realClock{}
	g := &Gate{
		inbox:     inbox,
		cfg:       cfg,
		clock:     clk,
		nowFn:     clk.Now,
		emit:      emit,
		waiters:   make(map[string]chan Outcome),
		taskWaits: make(map[string]string),
		closeCh:   make(chan struct{}),
	}
	return g
}

// SetClock replaces the clock (tests only).
func (g *Gate) SetClock(c Clock) {
	if c != nil {
		g.clock = c
		g.nowFn = c.Now
	}
}

// Close wakes all parked waiters with ErrGateClosed so shutdown can drain.
func (g *Gate) Close() {
	g.closeOnce.Do(func() {
		g.mu.Lock()
		g.closed = true
		close(g.closeCh)
		g.mu.Unlock()
	})
}

// TaskWaiting reports whether the task currently parks on a human response.
func (g *Gate) TaskWaiting(taskID string) bool {
	g.mu.Lock()
	defer g.mu.Unlock()
	_, ok := g.taskWaits[taskID]
	return ok
}

// PendingCount reports the number of pending inbox requests right now.
func (g *Gate) PendingCount() int {
	records, _, err := g.inbox.List(ActorID)
	if err != nil {
		return 0
	}
	n := 0
	for i := range records {
		if records[i].Pending() {
			n++
		}
	}
	return n
}

// Request persists a human request and parks until resolution, degradation,
// caller-context cancellation, or gate shutdown.
func (g *Gate) Request(task *types.AgentTask, req *types.HumanCallRequest) (*Outcome, error) {
	if task == nil || req == nil {
		return nil, errors.New("gate: task and request are required")
	}
	if req.Kind == "" {
		req.Kind = types.HumanKindInfo
	}
	switch req.Kind {
	case types.HumanKindDecision, types.HumanKindInfo:
	default:
		return nil, fmt.Errorf("gate: unsupported request kind %q", req.Kind)
	}
	ScrubRequest(req)
	if req.TaskID == "" {
		req.TaskID = task.TaskID
	}

	callID := storeutil.TimestampedID("call")
	req.CallID = callID
	if req.Deadline.IsZero() {
		req.Deadline = g.nowFn().Add(g.cfg.WaitTimeout)
	}
	req.Metadata = mergeMeta(req.Metadata, map[string]string{
		types.RequestKindMeta: req.Kind,
	})

	rec := &comm.InboxRecord{
		Message: comm.Message{
			ID:        callID,
			From:      "axis-runtime",
			To:        ActorID,
			Type:      comm.MsgQuery,
			Payload:   map[string]any{"request": req},
			Timestamp: g.nowFn(),
		},
		Status: comm.InboxStatusPending,
	}

	g.mu.Lock()
	if g.closed {
		g.mu.Unlock()
		return nil, ErrGateClosed
	}
	if len(g.taskWaits) >= g.cfg.MaxWaiting {
		g.mu.Unlock()
		return nil, fmt.Errorf("%w: %d pending", ErrTooManyWaiting, len(g.taskWaits))
	}
	ch := make(chan Outcome, 1)
	g.waiters[callID] = ch
	g.taskWaits[task.TaskID] = callID
	g.mu.Unlock()

	defer func() {
		g.mu.Lock()
		delete(g.waiters, callID)
		delete(g.taskWaits, task.TaskID)
		g.mu.Unlock()
	}()

	if err := g.inbox.Deliver(ActorID, rec); err != nil {
		return nil, fmt.Errorf("gate: deliver request: %w", err)
	}
	g.emitEvent(EventRequestCreated, task.TaskID, describe(req))

	outcome, err := g.park(task, req, ch)
	if err != nil {
		return nil, err
	}
	return outcome, nil
}

// park blocks the caller, driving the nudge cadence and degrade deadline.
func (g *Gate) park(task *types.AgentTask, req *types.HumanCallRequest, done <-chan Outcome) (*Outcome, error) {
	nudgesDone := 0
	for {
		now := g.nowFn()
		if shouldDegrade(req, nudgesDone, g.cfg, now) {
			return g.degrade(req), nil
		}
		wakeAt := parkUntil(req, nudgesDone, g.cfg, now)
		wait := wakeAt.Sub(now)
		if wait <= 0 {
			wait = time.Millisecond
		}
		timer := time.NewTimer(wait)
		select {
		case out := <-done:
			timer.Stop()
			g.emitEvent(EventRequestResolved, task.TaskID, callSummary(req.CallID, out))
			return &out, nil
		case <-g.closeCh:
			timer.Stop()
			return nil, ErrGateClosed
		case <-timer.C:
			now = g.nowFn()
			if shouldDegrade(req, nudgesDone, g.cfg, now) {
				timer.Stop()
				return g.degrade(req), nil
			}
			nudgesDone++
			g.recordNudge(req)
		}
	}
}

// Respond completes a pending request with a human answer and wakes the
// parked task. Resolution wins over any concurrent degrade (CAS).
func (g *Gate) Respond(choice, text string) (*comm.InboxRecord, error) {
	callID, _, err := g.singlePendingCall()
	if err != nil {
		return nil, err
	}
	return g.RespondCall(callID, choice, text)
}

// RespondCall completes the named request.
func (g *Gate) RespondCall(callID, choice, text string) (*comm.InboxRecord, error) {
	output := map[string]any{"choice": choice, "text": text}
	var rec *comm.InboxRecord
	err := g.inbox.MutatePending(ActorID, callID, func(r *comm.InboxRecord) {
		r.Status = comm.InboxStatusCompleted
		r.Output = output
	})
	if err != nil {
		return nil, err
	}
	rec, err = g.inbox.Get(ActorID, callID)
	if err != nil {
		return nil, err
	}
	g.wake(callID, Outcome{Approved: true, Choice: choice, Text: text, Output: output})
	return rec, nil
}

// List returns all human inbox records plus the corrupt-line count.
func (g *Gate) List() ([]comm.InboxRecord, int, error) {
	return g.inbox.List(ActorID)
}

// DegradeExpired degrades every pending request whose deadline has passed
// (startup scan path, design.md D3). Returns how many were degraded.
func (g *Gate) DegradeExpired() (int, error) {
	records, _, err := g.List()
	if err != nil {
		return 0, err
	}
	now := g.nowFn()
	degraded := 0
	for i := range records {
		r := records[i]
		if !r.Pending() {
			continue
		}
		req, rerr := requestFromRecord(&r)
		if rerr != nil || req.Deadline.After(now) {
			continue
		}
		if _, err := g.RespondCallDegrade(r.Message.ID, req.TaskID); err == nil {
			degraded++
		} else if !errors.Is(err, comm.ErrNotPending) {
			return degraded, err
		}
	}
	return degraded, nil
}

// RespondCallDegrade marks the named request degraded (fail-closed) and wakes
// the parked task with a Degraded outcome.
func (g *Gate) RespondCallDegrade(callID, taskID string) (*Outcome, error) {
	var out Outcome
	err := g.inbox.MutatePending(ActorID, callID, func(r *comm.InboxRecord) {
		r.Status = comm.InboxStatusDegraded
		r.Error = "human wait expired; degraded fail-closed"
		r.Output = map[string]any{"degraded": true}
		out = Outcome{Degraded: true, Output: r.Output}
	})
	if err != nil {
		return nil, err
	}
	g.emitEvent(EventRequestDegraded, taskID, callSummary(callID, out))
	g.wake(callID, out)
	return &out, nil
}

// wake delivers an outcome to the parked waiter if one exists (post-restart
// responses have no waiter and simply persist — v1 non-goal covers this).
func (g *Gate) wake(callID string, out Outcome) {
	g.mu.Lock()
	ch := g.waiters[callID]
	g.mu.Unlock()
	if ch != nil {
		ch <- out
	}
}

func (g *Gate) degrade(req *types.HumanCallRequest) *Outcome {
	out := Outcome{Degraded: true, Output: map[string]any{"degraded": true}}
	if req.Kind == types.HumanKindDecision && len(req.Options) > 0 {
		// Conservative default = first option, recorded as an assumption.
		out.Output["conservative_default"] = req.Options[0]
		out.Choice = req.Options[0]
	}
	err := g.inbox.MutatePending(ActorID, req.CallID, func(r *comm.InboxRecord) {
		r.Status = comm.InboxStatusDegraded
		r.Error = "human wait expired; degraded fail-closed"
		r.Output = out.Output
	})
	if errors.Is(err, comm.ErrNotPending) {
		// Concurrent RespondCall / RespondCallDegrade won the CAS; honor the persisted outcome.
		if rec, gerr := g.inbox.Get(ActorID, req.CallID); gerr == nil && rec.Status == comm.InboxStatusCompleted {
			choice, _ := rec.Output["choice"].(string)
			text, _ := rec.Output["text"].(string)
			resolved := Outcome{Approved: true, Choice: choice, Text: text, Output: rec.Output}
			g.emitEvent(EventRequestResolved, req.TaskID, callSummary(req.CallID, resolved))
			return &resolved
		}
		return &out
	}
	g.emitEvent(EventRequestDegraded, req.TaskID, callSummary(req.CallID, out))
	return &out
}

// recordNudge appends one audit reminder and ages the inbox record without
// changing its pending status (design.md D3: nudges are audit entries).
func (g *Gate) recordNudge(req *types.HumanCallRequest) {
	now := g.nowFn()
	err := g.inbox.MutatePendingAllowPending(ActorID, req.CallID, func(r *comm.InboxRecord) {
		r.Nudges++
		t := now
		r.LastNudgeAt = &t
	})
	if err == nil {
		g.emitEvent(EventRequestNudged, req.TaskID, fmt.Sprintf("call %s nudge", req.CallID))
	}
}

func (g *Gate) singlePendingCall() (string, *comm.InboxRecord, error) {
	records, _, err := g.List()
	if err != nil {
		return "", nil, err
	}
	for i := range records {
		if records[i].Pending() {
			return records[i].Message.ID, &records[i], nil
		}
	}
	return "", nil, fmt.Errorf("%w: no pending requests", comm.ErrCallNotFound)
}

func (g *Gate) emitEvent(eventType, taskID, message string) {
	if g.emit != nil {
		g.emit(eventType, taskID, message)
	}
}

func mergeMeta(base, extra map[string]string) map[string]string {
	out := make(map[string]string, len(base)+len(extra))
	for k, v := range base {
		out[k] = v
	}
	for k, v := range extra {
		out[k] = v
	}
	return out
}

func describe(req *types.HumanCallRequest) string {
	if req.Prompt != "" {
		return fmt.Sprintf("kind=%s %s", req.Kind, req.Prompt)
	}
	return "kind=" + req.Kind
}

func callSummary(callID string, out Outcome) string {
	switch {
	case out.Degraded:
		return callID + " degraded"
	case out.Approved:
		return callID + " approved"
	default:
		return callID + " resolved"
	}
}

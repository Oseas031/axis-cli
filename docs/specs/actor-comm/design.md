# Unified Actor Model & Communication Layer

**Status**: Planned
**Depends on**: `docs/architecture/kernel-abstraction-model.md`
**Decision**: Homogeneous (同构) — all participants implement the same Actor interface.

## Core Decision

Axis treats all participants — human, Agent, external service, team — as **Actors**. The Kernel does not distinguish identity; it only sees interface behavior.

```
┌─────────────────────────────────────────────────┐
│                  Kernel                          │
│  Scheduler sees only: Actor[]                   │
│  Dispatch sees only: Actor.Receive(task)        │
│  Comm sees only: Actor.Send(to, msg)            │
└──────────────────────┬──────────────────────────┘
                       │
        ┌──────────────┼──────────────┐
        │              │              │
   ┌────▼────┐   ┌────▼────┐   ┌────▼────┐
   │ LLM     │   │ CLI     │   │ Webhook │
   │ Adapter │   │ Adapter │   │ Adapter │
   └─────────┘   └─────────┘   └─────────┘
   (Agent)        (Human)        (Service)
```

## Actor Interface

```go
package actor

// Actor is the universal participant interface.
// Humans, Agents, services, and teams all implement this.
type Actor interface {
    ID() string
    Receive(ctx context.Context, msg Message) error
    Status() ActorStatus
}

type ActorStatus int

const (
    ActorReady ActorStatus = iota
    ActorBusy
    ActorOffline
)

// Message is the universal communication unit.
type Message struct {
    ID        string         `json:"id"`
    From      string         `json:"from"`       // actor ID
    To        string         `json:"to"`         // actor ID
    Type      MessageType    `json:"type"`
    Payload   map[string]any `json:"payload"`
    Timestamp time.Time      `json:"timestamp"`
    ReplyTo   string         `json:"reply_to,omitempty"` // for request-response
}

type MessageType string

const (
    MsgTask     MessageType = "task"      // task assignment
    MsgResult   MessageType = "result"    // task completion
    MsgQuery    MessageType = "query"     // information request
    MsgNotify   MessageType = "notify"    // one-way notification
    MsgDelegate MessageType = "delegate"  // delegation to another actor
)
```

## Communication Layer

### Mailbox

Every Actor has a persistent mailbox. Messages are delivered asynchronously.

```go
package comm

// Mailbox provides async message delivery for an Actor.
type Mailbox interface {
    Send(msg Message) error
    Receive(actorID string) ([]Message, error)
    Peek(actorID string) ([]Message, error)  // non-destructive read
    Ack(actorID string, msgIDs []string) error
}
```

Storage: `.axis/comm/<actor-id>.jsonl` (P0 local file, P2 network queue).

### Router

Routes messages between Actors. Handles offline delivery (queue until Actor comes online).

```go
package comm

// Router delivers messages between Actors.
type Router struct {
    registry map[string]Actor
    mailbox  Mailbox
}

func (r *Router) Send(msg Message) error {
    target, online := r.registry[msg.To]
    if online && target.Status() == ActorReady {
        return target.Receive(context.Background(), msg)
    }
    // Actor offline or busy — queue in mailbox
    return r.mailbox.Send(msg)
}
```

## Adapters

### LLMAdapter (Agent)

```go
type LLMAdapter struct {
    id       string
    provider provider.ModelProvider
    tools    *tool.Registry
}

func (a *LLMAdapter) ID() string { return a.id }
func (a *LLMAdapter) Status() ActorStatus { return ActorReady }
func (a *LLMAdapter) Receive(ctx context.Context, msg Message) error {
    // Convert message to task, execute via LLM loop
}
```

### CLIAdapter (Human)

```go
type CLIAdapter struct {
    id string
}

func (a *CLIAdapter) ID() string { return a.id }
func (a *CLIAdapter) Status() ActorStatus { return ActorOffline } // until human responds
func (a *CLIAdapter) Receive(ctx context.Context, msg Message) error {
    // Queue to mailbox; human reads via `axis inbox`
    return nil
}
```

Human interaction:
```bash
axis inbox                    # list pending messages
axis respond <msg-id> "..."   # reply to a message
axis delegate <msg-id> <actor> # forward to another actor
```

### WebhookAdapter (External Service)

```go
type WebhookAdapter struct {
    id      string
    url     string
    client  *http.Client
}

func (a *WebhookAdapter) Receive(ctx context.Context, msg Message) error {
    // POST message to webhook URL
}
```

## Scheduling Integration

The Scheduler no longer distinguishes executor types. It:
1. Picks a ready task
2. Finds the assigned Actor (or picks one based on capability matching)
3. Calls `actor.Receive(taskMsg)`
4. Waits for result message (with SLA timeout)

```go
// In dispatcher
func (d *Dispatcher) dispatch(task *types.AgentTask) {
    actorID := d.resolveActor(task)  // capability matching
    msg := Message{
        Type:    MsgTask,
        To:      actorID,
        Payload: task.Input,
    }
    d.router.Send(msg)
}
```

## Migration Path from Current Code

| Current | Becomes | Change Type |
|---------|---------|-------------|
| `AgentExecutor` | `LLMAdapter` implementing `Actor` | Rename + interface change |
| `HumanExecutor` | `CLIAdapter` implementing `Actor` | Rename + async |
| `Dispatcher.executeAgentTask` | `Router.Send(MsgTask)` | Simplification |
| `Dispatcher.executeHumanTask` | Same as above (no special case) | Deletion |
| `FollowUpTaskGenerator` | `Actor.Send(MsgDelegate)` | Unification |

## Implementation Plan

### P0: Local Mailbox + Actor Interface (this week)

```
internal/actor/actor.go       — Actor interface, Message types
internal/comm/mailbox.go      — JSONL-backed local mailbox
internal/comm/router.go       — In-process message router
internal/comm/mailbox_test.go
internal/comm/router_test.go
```

### P1: Adapter Migration

```
internal/actor/llm_adapter.go     — wraps current AgentExecutor
internal/actor/cli_adapter.go     — wraps current HumanExecutor
cmd/axis/inbox_cmd.go             — axis inbox / axis respond
```

### P2: Network Communication

```
- Mailbox backend: local JSONL → Redis/NATS/JSONL-over-HTTP
- Actor registry: local map → distributed registry
- Cross-machine A2A communication
```

### P3: Industry Integration

```
- WebhookAdapter for external services
- TeamAdapter (composite Actor)
- Capability marketplace (Actors advertise what they can do)
```

## Relationship to Kernel Model

| Kernel Syscall | Actor Model Expression |
|---|---|
| `submit_task` | `Router.Send(MsgTask)` |
| `query_state` | `Actor.Status()` + mailbox peek |
| `spawn` | `Router.Send(MsgDelegate)` to new Actor |
| `yield` | Actor returns `ActorBusy`, task re-queued |
| `introspect` | Actor reads own mailbox + state |

## Non-Goals (P0)

- No distributed consensus
- No Actor persistence across restarts (stateless adapters, state in mailbox)
- No capability negotiation protocol (static assignment in P0)
- No encryption of mailbox content

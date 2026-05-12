# Unified Actor & Communication Layer Requirements

**Status**: Planned
**Decision**: Homogeneous Actor model (人机同构)

## Summary

All participants in Axis — Agents, humans, external services — implement the same Actor interface. Communication between Actors uses a unified message protocol delivered through persistent mailboxes. The Kernel does not distinguish Actor identity; it only observes interface behavior.

## Functional Requirements

### FR1: Actor Interface

- All participants implement `Actor` with: `ID()`, `Receive(msg)`, `Status()`
- Status: Ready / Busy / Offline
- No type-specific dispatch paths in Kernel

### FR2: Message Protocol

- Universal Message struct: ID, From, To, Type, Payload, Timestamp, ReplyTo
- Message types: task, result, query, notify, delegate
- Messages are self-describing (no out-of-band schema needed)

### FR3: Mailbox

- Every Actor has a persistent mailbox
- Messages delivered async; queued if Actor offline
- Storage: `.axis/comm/<actor-id>.jsonl`
- Operations: Send, Receive (destructive), Peek (non-destructive), Ack

### FR4: Router

- Routes messages between Actors
- Online Actor: deliver directly via `Receive()`
- Offline Actor: queue in mailbox
- No message loss guarantee (at-least-once in P0)

### FR5: CLI Adapter (Human)

- Human is an Actor with ID (default: "human")
- `axis inbox` — list pending messages
- `axis respond <msg-id> "<text>"` — reply
- `axis delegate <msg-id> <actor-id>` — forward to another Actor

### FR6: LLM Adapter (Agent)

- Wraps existing ContractExecutor / AgentExecutor logic
- Receives MsgTask, executes via LLM loop, sends MsgResult back

### FR7: Scheduler Integration

- Scheduler assigns tasks to Actors by capability matching
- No separate code paths for human vs agent dispatch
- SLA timeout applies uniformly to all Actors

## Non-Goals

- No distributed messaging (P0 is local only)
- No Actor discovery protocol (static registry in P0)
- No message encryption
- No multi-machine deployment

## Acceptance Criteria

- `Actor` interface defined and implemented by at least 2 adapters
- Messages can be sent between two Actors via Router
- Mailbox persists messages across process restarts
- `axis inbox` shows pending messages for human Actor
- `go test -race ./internal/actor/... ./internal/comm/...` passes

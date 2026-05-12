# Unified Actor & Communication Layer Tasks

**Status**: Planned
**Implements**: requirements.md + design.md

---

## P0: Core Infrastructure (本周)

### T1: Actor Interface + Message Types

- `internal/actor/actor.go` — Actor interface, ActorStatus, Message, MessageType
- `internal/actor/actor_test.go` — type assertions, message serialization

### T2: Mailbox

- `internal/comm/mailbox.go` — JSONL-backed file mailbox
- `internal/comm/mailbox_test.go` — Send/Receive/Peek/Ack, persistence across restarts

### T3: Router

- `internal/comm/router.go` — in-process message router with online/offline delivery
- `internal/comm/router_test.go` — direct delivery, offline queueing

### T4: LLM Adapter

- `internal/actor/llm_adapter.go` — wraps existing executor logic as Actor
- Tests: receives MsgTask, produces MsgResult

### T5: CLI Adapter + Commands

- `internal/actor/cli_adapter.go` — queues to mailbox
- `cmd/axis/inbox_cmd.go` — axis inbox / axis respond / axis delegate
- Tests: message round-trip human ↔ agent

---

## P1: Migration (下周)

### T6: Dispatcher Unification

- Remove separate `executeAgentTask` / `executeHumanTask` paths
- Single path: resolve Actor → Router.Send(MsgTask) → wait for MsgResult

### T7: Scheduler Actor-Awareness

- Capability matching: assign task to Actor with matching capabilities
- SLA timeout uniform across all Actor types

---

## P2: Network (后续)

### T8: Network Mailbox Backend

- Replace JSONL with pluggable backend (Redis / NATS / HTTP)
- Actor registry: distributed discovery

### T9: WebhookAdapter

- External services as Actors
- Inbound webhook → message; outbound message → HTTP POST

---

## Definition of Done (P0)

- Actor interface exists with 2+ implementations
- Messages route between Actors via Router
- Mailbox survives process restart
- `axis inbox` works
- `go test -race ./internal/actor/... ./internal/comm/...` passes
- No new external dependencies

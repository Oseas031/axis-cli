# Adaptive Context Assembly Requirements

> 实现 agent-native-first-principles.md P2（Query is Context）


## Summary

Adaptive Context Assembly prepares task-specific, auditable, budgeted context bundles for Agent execution.

It is not memory, not permission, not control, and not a chatbot. It is the situational readiness layer between `AgentTask` and Agent execution.

The first implementation must be additive and non-invasive. It should help Axis answer: given a task goal, what context does the Agent need to act well, why was that context selected, and what was intentionally left out?

## Design Philosophy

### Prepared Context, Not Piled Memory

Context is prepared for the current task. It is not a dump of all available history, files, documents, logs, or memories.

### Relevance Over Volume

A small precise context bundle is better than a large noisy one. The system must avoid context pollution.

### Auditable Assembly

Every selected context packet must have source, reason, relevance, and budget impact.

### Contractual Fuel, Not Control Plane

Context serves `AgentTask` and `AgentContract`. It must not become a second scheduler, permission system, policy engine, or chatbot runtime.

### Competence Tunes Context

Agent competence changes context shape, not hidden authority. Less reliable agents receive more concrete scaffolding; more reliable agents receive more abstract context.

## Users

- Agents preparing to execute Axis tasks
- Developers inspecting why an Agent received certain context
- Future shell/CLI users previewing task readiness
- Future evaluation loops comparing context quality with task outcomes

## Functional Requirements

### FR1: Context bundle model

Axis must define a context bundle concept that can hold selected context packets for a task.

Each packet must include at minimum:

- ID
- type
- source
- summary or content
- reason
- relevance score

### FR2: Assembly trace

Each assembly run must produce an auditable trace explaining:

- task goal used for assembly
- selected packets
- selection reasons
- relevance scores
- budget usage
- important omissions when known

### FR3: Rule-based P0 assembly

P0 must not require vector databases, LLM calls, or external services.

It should use deterministic rules over task input, contract ID, metadata, and known project documentation paths.

### FR4: Budgeted context

Assembly must enforce a small explicit budget.

Budget may initially be expressed as:

- maximum packet count
- maximum bytes

Token-level budget may be added later.

### FR5: Preview before execution

P0 must support previewing the selected context without injecting it into Agent execution.

This keeps the first slice safe and observable.

### FR6: Authority hierarchy

Context ranking should respect source authority:

1. Current user request / task input
2. Current workspace code
3. Current specs
4. Current test results
5. Accepted design principles
6. Recent summaries
7. Long-term memories
8. Historical reports or deprecated drafts

### FR7: Non-invasive execution path

P0 must not change scheduler, dispatcher, provider, or contract executor semantics.

### FR8: Retrieval-backed assembly (P5)

Assembly may use a local TF-IDF index over Axis documents and user project code to improve relevance ranking beyond deterministic rules.

- Index must be built by explicit user action (`axis context index`).
- No background watcher or automatic ingestion.
- Index update uses mtime-based incremental detection.
- Missing index must not break assembly; system falls back to rule-based mode.

### FR9: Hybrid recall and rerank (P5)

Assembly must combine deterministic rule recall with vector similarity rerank:

- Rule-based candidates provide deterministic coverage.
- TF-IDF cosine similarity provides relevance refinement.
- Hybrid trace must show both rule selection and retrieval scores.

### FR10: Local-only embedding (P5)

P5 embedding must be zero-config, pure local, no external service:

- TF-IDF term vectors computed from scanned documents.
- No API key, no network request, no background process.
- P1 may optionally integrate remote embedding via existing provider profiles.

### FR11: Explicit index lifecycle (P5)

Users control index lifecycle through CLI:

- `axis context index --rebuild` builds from scratch.
- `axis context index --update` incrementally updates changed files.
- `axis context index --status` reports index health and coverage.

## Non-Goals

- No vector database
- No persistent memory database
- No automatic permission escalation
- No policy-heavy control plane
- No Web UI or TUI
- No hidden prompt injection into execution
- No automatic full-repo ingestion
- No LLM parser or reranker in P0
- No multi-agent context market
- No external knowledge base ingestion in P5
- No automatic index rebuild on file change
- No background file watcher

## Acceptance Criteria

- Requirements, design, and tasks exist under `docs/specs/adaptive-context-assembly/`.
- The feature is defined as context readiness, not memory or control.
- P0 scope is rule-based and preview-first.
- Context packet and bundle concepts are specified.
- Scheduler/orchestrator/contract semantics remain unchanged in P0.
- P5 retrieval-backed assembly uses pure local TF-IDF with zero external dependencies.
- Hybrid assembler combines rule recall with vector rerank and falls back gracefully when index is missing.
- Explicit index lifecycle CLI (`axis context index --rebuild/--update/--status`) is available.

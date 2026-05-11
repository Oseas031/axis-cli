# Agent-Native Scheduling System Whitepaper

When everyone asks, "What exactly does this project do?", we do not evade the question or pile up concepts. We give a clear answer with the most rigorous logic and the most concise expression: this project aims to build an Agent-native scheduling system that provides AI Agents with unified scheduling capabilities across the digital and physical worlds, achieving Agent-led full-process autonomy.

## I. Pain Points of the Era: Why Do We Need an Agent-Native Scheduling System?

The large-scale adoption of AI Agent applications (development, operations, creative work, enterprise processes, etc.) is no longer bottlenecked by model capabilities, but by **scheduling capabilities**. Existing Agent tools lack a unified scheduling system and suffer from three insurmountable shortcomings:

### 1. Lack of Scheduling Capability
Agents can only execute passively, unable to autonomously schedule other Agents, tools, or humans, becoming "isolated single-point intelligences."

### 2. Capability Boundary Limitations
Agents excel at efficient computation in the digital world but cannot reach execution and decision-making in the physical world, creating a "digital-physical gap."

### 3. Low Collaboration Efficiency
Multi-Agent collaboration lacks unified scheduling, making it impossible to achieve parallel orchestration, dependency management, and exception fallback, leading to difficulties in large-scale Agent adoption.

Existing AI Agent tools are either "single-point executors" or "human-assistance tools," and none solve the core pain point of "unified scheduling + capability fusion" — this is the fundamental motivation for building an Agent-native scheduling system.

## II. Project Definition: What Exactly Is an Agent-Native Scheduling System?

The core of an Agent-native scheduling system is a **unified scheduling platform designed for AI Agents**, analogous to an operating system scheduler. It has the core capabilities of task orchestration, resource scheduling, and context management, and its essence is:

**To provide AI Agents with a unified scheduling entry point, supporting Agents in scheduling other Agents, tools, and human tasks, achieving Agent-led "digital + physical" full-process autonomy.**

It is not an upgrade of CLI tools, but a brand-new system architecture:

- **Traditional Agent tools**: Single-point execution, unable to schedule other Agents
- **Agent-native scheduling system**: Unified scheduling, supporting Agent inter-calling, tool scheduling, and human task scheduling

In short: it is the "operating system kernel" for Agents, the "unified scheduling platform" connecting the digital and physical worlds — this is the entire core of this project.

**CLI Positioning**: The CLI is just **one client** of the scheduling system, not the core. The core is the scheduling system itself; the CLI is merely one of the entry points for human interaction.

## III. Core Architecture: How to Achieve Agent-Native Scheduling?

The Agent-native scheduling system is **Agent-centric** and adopts a three-layer decoupled architecture:

### 3.1 Core Kernel (Scheduling Layer): Agent Scheduling Engine

Bears the core functions of task orchestration, resource scheduling, and exception handling for Agents:

- **Task orchestration**: Supports DAG orchestration, dependency management, and parallel scheduling
- **Resource scheduling**: Unified scheduling of Agents, tools, and human tasks
- **Context management**: Cross-task, cross-session context passing
- **Exception handling**: Task failure retry, exception fallback, and human intervention

### 3.2 Core Capability (Execution Layer): Human Task Queue

Incorporates human tasks into the scheduling system, opening up the execution chain in the physical world:

- **Human task queue**: Agents submit human tasks, and humans execute them on demand
- **Task protocol**: Standardized task definition, status tracking, and result return
- **Lightweight terminal**: Humans do not need technical knowledge; they receive tasks, execute them, and fill in results

### 3.3 Client Layer (Interaction Layer): Multiple Clients

Provides multiple clients for different usage scenarios:

- **CLI client**: Command-line tool for human interaction
- **API client**: HTTP API for Agent invocation
- **SDK client**: Programming language SDK for easy integration

**Note**: The CLI is just one of the clients, not the core. The core is the scheduling system itself.

## IV. Core Value: What Core Problems Does It Solve?

### 4.1 Underlying Value (Irreplaceable in Final State)

Provides Agents with unified scheduling capabilities, breaking through Agent capability boundaries and achieving true multi-Agent collaboration — this is the key infrastructure for Agents to evolve from "single-point intelligence" to "system intelligence."

### 4.2 Business Value (Verifiable in Practice)

Solves the scheduling bottleneck for large-scale Agent adoption, allowing a single user to manage dozens or even hundreds of Agent tasks in parallel, significantly reducing enterprise AI adoption costs.

### 4.3 Ecosystem Value (Self-Bootstrapping and Sustainable)

Provides APIs and SDKs to facilitate third-party integration and build an Agent scheduling ecosystem.

## V. Evolution Path: From Starting Point to Final State

### Starting Point: Basic Agent Scheduling
- Implement basic task scheduling (FIFO)
- Implement simple task orchestration
- Implement CLI client

### Evolution Path: Gradual Enhancement
- Add DAG parallel scheduling
- Add human task queue
- Add API client
- Add SDK client

### Final State: Agent-Native Scheduling System
- Complete task orchestration capabilities
- Complete resource scheduling capabilities
- Complete exception handling capabilities
- Multi-client support (CLI, API, SDK)

## VI. Summary

In response to the question, "What exactly does this project do?", our answer is concise and firm:

**An Agent-native scheduling system is a unified scheduling platform designed for AI Agents. Its core is to provide Agents with task orchestration, resource scheduling, and context management capabilities, supporting Agents in scheduling other Agents, tools, and human tasks, achieving Agent-led full-process autonomy.**

The CLI is just one client of the scheduling system, not the core. The core is the scheduling system itself.

The wave of the era has arrived, and the Agent-native scheduling system has already set sail.

# System Architecture Visualization

## 1. Overall System Architecture

```mermaid
graph TB
    subgraph Client Layer
        CLI[CLI Client]
    end

    subgraph Core Kernel
        Scheduler[Scheduler<br/>FIFO + Dependency Management]
        Dispatcher[Dispatcher]
        Lifecycle[Lifecycle Management]
        StateStore[State Store]
    end

    subgraph Contract Layer
        ContractExecutor[Contract Executor<br/>Input/Output Validation]
    end

    subgraph Human Layer
        HumanExecutor[Human Executor<br/>Task Queue]
    end

    CLI --> Scheduler
    Scheduler --> Dispatcher
    Dispatcher --> ContractExecutor
    Dispatcher --> HumanExecutor
    Scheduler --> StateStore
    Scheduler --> Lifecycle
```

## 2. Task Scheduling Flow

```mermaid
sequenceDiagram
    participant CLI
    participant Scheduler
    participant Dispatcher
    participant Executor
    participant StateStore

    CLI->>Scheduler: Submit task
    Scheduler->>StateStore: Save task status
    Scheduler->>Scheduler: Check dependencies
    Scheduler->>Dispatcher: Dispatch task
    Dispatcher->>Executor: Execute task
    Executor-->>Dispatcher: Return result
    Dispatcher-->>Scheduler: Return result
    Scheduler->>StateStore: Update task status
    Scheduler-->>CLI: Return result
```

## 3. Contract Validation Flow

```mermaid
graph LR
    A[Input Data] --> B[Input Schema Validation]
    B --> C{Passed?}
    C -->|No| D[Return Error]
    C -->|Yes| E[Execute Task]
    E --> F[Output Data]
    F --> G[Output Schema Validation]
    G --> H{Passed?}
    H -->|No| I[Return Error]
    H -->|Yes| J[Return Result]
```

## 4. Milestone 1 Core Workflow

```mermaid
graph TB
    A[Task Submission] --> B[FIFO Queue]
    B --> C{Check Dependencies}
    C -->|Dependencies Incomplete| D[Wait]
    C -->|Dependencies Complete| E[Execute Task]
    E --> F[Input Validation]
    F --> G{Passed?}
    G -->|No| H[Fail]
    G -->|Yes| I[Execute]
    I --> J[Output Validation]
    J --> K{Passed?}
    K -->|No| H
    K -->|Yes| L[Complete]
    L --> M[Update Status]
```

## 5. System Module Relationships

```mermaid
graph LR
    Scheduler[Scheduler] --> Dispatcher[Dispatcher]
    Scheduler --> StateStore[State Store]
    Scheduler --> Lifecycle[Lifecycle]
    Dispatcher --> ContractExecutor[Contract Executor]
    Dispatcher --> HumanExecutor[Human Executor]
```

## 6. Milestone Evolution Path

```mermaid
graph LR
    M1[Milestone 1<br/>Basic Scheduling] --> M2[Milestone 2<br/>Parallel Scheduling]
    M2 --> M3[Milestone 3<br/>Ecosystem Maturity]

    M1 --> M1F[FIFO Scheduling<br/>Simple Dependencies<br/>Input/Output Validation]
    M2 --> M2F[DAG Parallel<br/>Admission Rules<br/>SLA Agreements]
    M3 --> M3F[Tool Invocation<br/>Exception Handling<br/>Multi-Client]
```

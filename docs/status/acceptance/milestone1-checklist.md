# Milestone 1 Technical Admission Checklist (Occam's Razor Simplified Version)

**[Chinese version / 中文版](../../zh/status/acceptance/milestone1-checklist.md)**

## Milestone Goal
Focus on verifying Agent scheduling core capabilities, implementing a minimal closed loop of basic task scheduling, simple dependency management, and input/output validation.

**Important Notes**:
- **End State**: Agent-native scheduling system
- **CLI Positioning**: CLI is just a client of the scheduling system, not the core
- **Milestone 1 Scope**: Basic Agent scheduling (FIFO + simple dependency management + I/O validation)
- **Design Principle**: Occam's Razor - minimum viable, only implement the minimal feature set needed to verify core concepts

## Technical Admission Standards (Simplified)

### 1. Basic Task Scheduling (Core Capability)

#### 1.1 Task Queue Verification
- [ ] FIFO task queue implementation
- [ ] Task submission/consumption working normally
- [ ] Task queue capacity 1000 non-blocking
- [ ] Task status tracking (pending/running/completed/failed)

#### 1.2 Scheduling Strategy Verification
- [ ] FIFO scheduling strategy verification passed
- [ ] Serial task execution verification passed

### 2. Simple Task Orchestration (Core Capability)

#### 2.1 Task Dependency Management
- [ ] Task dependency definition verification passed
- [ ] Dependent tasks execute only after dependencies complete
- [ ] Circular dependency detection verification passed

#### 2.2 Task Orchestration Execution
- [ ] Serial task orchestration verification passed
- [ ] Orchestration result return verification passed

### 3. Contract Input/Output Validation (Core Capability - Simplified)

#### 3.1 Input Schema Validation
- [ ] Input Schema definition verification passed
- [ ] Input field type validation passed
- [ ] Required input field validation passed
- [ ] Enum value validation passed

#### 3.2 Output Schema Validation
- [ ] Output Schema definition verification passed
- [ ] Output field type validation passed
- [ ] Required output field validation passed

### 4. Basic State Storage (Core Capability)

#### 4.1 State Storage Verification
- [ ] Task state save verification passed
- [ ] Task state query verification passed
- [ ] In-memory state storage verification passed

### 5. CLI Client (Compatibility Layer)

#### 5.1 Basic CLI Verification
- [ ] Basic CLI implemented using cobra framework
- [ ] Basic command parsing verification passed
- [ ] Signal handling (Ctrl+C) verification passed

### 6. End-to-End Closed Loop Verification

#### 6.1 Basic Closed Loop Verification
- [ ] Task submit → schedule → execute → result return closed loop verification passed
- [ ] Basic exception handling verification passed
- [ ] End-to-end success rate ≥ 80%

### 7. Test Coverage Verification (Simplified)

#### 7.1 Unit Tests
- [ ] Task scheduling unit test coverage ≥ 60%
- [ ] Task orchestration unit test coverage ≥ 60%
- [ ] Schema validation unit test coverage ≥ 60%

#### 7.2 Integration Tests
- [ ] Basic end-to-end integration test verification passed

### 8. Build Verification (Simplified)

#### 8.1 Build Verification
- [ ] Go compilation with no warnings or errors
- [ ] Static binary generation verification passed
- [ ] Windows platform build verification passed

## Acceptance Process (Simplified)

### Phase 1: Self-Check
- Complete all check item self-checks
- Generate self-check report

### Phase 2: Automated Testing
- Run test suite
- Generate test report

### Phase 3: Manual Acceptance
- Basic functionality verification
- Milestone 1 formal admission

## Failure Handling

### Check Item Failure Handling
- Single item failure: Fix and re-verify
- Blocking failure: Pause admission, root cause analysis

## Milestone 1 Completion Markers

- [ ] All core check items passed
- [ ] Technical admission standards met
- [ ] Self-check report passed
- [ ] Automated tests passed
- [ ] Manual acceptance passed

## Next Step: Milestone 2 Preparation

After Milestone 1 admission, immediately start Milestone 2 preparation:
- DAG parallel scheduling design
- Contract admission rules (local/remote validation) design
- SLA constraints (timeout, retry, circuit breaker) design
- Error code system design
- Contract version management design

## Milestone 1 Scope Summary

**Included**:
- ✅ FIFO task scheduling
- ✅ Simple dependency management (dependencies must complete before execution)
- ✅ Input/output Schema validation
- ✅ Basic state storage
- ✅ Basic CLI

**Not Included (Deferred to Milestone 2+)**:
- ❌ DAG parallel scheduling
- ❌ Contract admission rules (local/remote validation)
- ❌ SLA constraints (timeout, retry, circuit breaker)
- ❌ Error code system
- ❌ Contract version management
- ❌ Tool invocation layer
- ❌ Global event bus
- ❌ Controller architecture

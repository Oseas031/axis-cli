---
description: Minimal entry routing from user tasks to project workflows
---

# Workflow Entry

This file is the first entry point when an Agent receives a task. It does one thing: route user tasks to the minimal necessary workflow combination.

## 1. Core Principles

1. **Read the entry first, then act**: Do not bypass `workflow/entry.md` and make large changes directly.
2. **Minimal combination**: Prioritize reusing existing workflows; do not add a new workflow for a single task.
3. **Docs first**: New features must have `requirements.md`, `design.md`, `tasks.md`, and `workflow-binding.md`.
4. **Honest verification**: If it has not been run or tools are missing, write "not verified"; do not write "passed".
5. **Scope restraint**: Grand designs can update direction, but must not automatically expand the current milestone scope.
6. **Design sovereignty**: When the user has transferred design sovereignty, the Agent actively organizes the design route and writes documentation; only ask for confirmation before destructive or high-risk operations.

---

## 2. Four Route Categories

### 1. Feature Implementation / Bug Fix

```text
wf-pr-check + wf-ci + wf-doc-006
```

Execution requirements: locate root cause -> minimal implementation -> test verification -> sync docs.  
CLI / Shell behavior changes must be accompanied by behavioral tests or executable verification commands.

### 2. New Feature / New Spec

```text
wf-doc-004 + wf-pr-check + wf-ci + wf-doc-006
```

Execution requirements: write or confirm specs and `workflow-binding.md` first, then implement.  
If binding is missing, add binding first; do not code directly.

### 3. Documentation / Design / Workflow Adjustment

```text
wf-doc-004 + wf-doc-006
```

Execution requirements: update core design or reports first, then update entry docs and handover status.  
When the project's core positioning changes, entry documents must be checked as a group:

```text
README.md
docs/README.md
docs/guides/QUICKSTART.md
docs/product/WHITEPAPER.md
docs/status/current-progress.md
HANDOVER.md
```

### 4. CI/CD / Release / Security / Monitoring

```text
wf-ci or wf-pr-check or wf-security or wf-cd or wf-monitoring + wf-doc-004
```

Execution requirements: check `.github/config/registry.yml` first, then modify the corresponding workflow.  
Build, test, and security may be used as hard gates; experience-based checks default to reminders only.

---

## 3. Retrospective Rules

Work retrospectives use:

```text
wf-doc-004 + wf-doc-006
```

Requirements:

1. Each work item is assigned to only one upstream workflow.
2. For each category, distill successful practices, root causes of problems, temporary solutions, and blockers.
3. Evaluate by keep / correct / remove /沉淀 (沉淀 means distill into permanent rules).
4. Only feed back into workflow the rules that are executable, not overly controlling, and consistent with Occam's Razor.

---

## 4. Default Execution Order

```text
1. Read workflow/entry.md
2. Select the minimal upstream workflow combination
3. Read related workflow docs or registry
4. Write/update spec, plan, or report
5. Implement the minimal change
6. Run necessary verification
7. Update current-progress / HANDOVER / report
```

---

## 5. Prohibited Actions

- Do not make large changes without reading the workflow first.
- Do not add an independent workflow for a single feature.
- Do not upgrade advisory guidelines to hard gates.
- Do not expand the current milestone scope without a Spec-RDT.
- Do not handle existing entry documents by deleting and recreating them.
- For the full list of prohibitions, see `CLAUDE.md` section 1.


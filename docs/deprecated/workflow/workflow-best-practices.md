# Workflow Best Practices

**[Chinese version / 中文版](../zh/workflow/workflow-best-practices.md)**

## Overview

This document records best practices for developing and maintaining GitHub Actions workflows in the Axis project, based on actual experience.

## 1. Workflow Trigger Design

### 1.1 Precise Path Filtering
Use path filters to avoid unnecessary workflow execution:

```yaml
on:
  push:
    paths:
      - '**.go'           # Only trigger on Go file changes
      - 'go.mod'
      - 'go.sum'
      - '.github/config/registry.yml'
```

### 1.2 Event Type Combinations
Correctly combine different event types:

```yaml
on:
  push:
    branches: [ main, develop ]
  pull_request:
    branches: [ main, develop ]
  workflow_run:
    workflows: [CI, PR Quality Check]
    types: [completed]
```

## 2. Conditional Execution Patterns

### 2.1 Event Type Check Pattern
When accessing specific event properties, check the event type first:

```yaml
jobs:
  validate-registry:
    if: |
      github.event_name == 'push' && contains(github.event.head_commit.modified, '.github/config/registry.yml') ||
      github.event_name == 'pull_request' && contains(github.event.pull_request.changed_files, '.github/config/registry.yml')
```

### 2.2 Workflow Run Conditions
Only run workflows under specific conditions:

```yaml
on:
  workflow_run:
    workflows: [CI]
    types: [completed]

jobs:
  metrics:
    if: github.event_name == 'workflow_run' && github.event.workflow_run.conclusion == 'success'
```

## 3. Data Validation Patterns

### 3.1 File Existence Check
Handle files that may not exist:

```yaml
- name: Check if benchmark results exist
  id: check-benchmark
  run: |
    if [ -s benchmark.txt ] && grep -q "Benchmark" benchmark.txt; then
      echo "has_benchmark=true" >> $GITHUB_OUTPUT
    else
      echo "has_benchmark=false" >> $GITHUB_OUTPUT
    fi

- name: Process benchmark
  if: steps.check-benchmark.outputs.has_benchmark == 'true'
  run: process_benchmark
```

### 3.2 Data Integrity Validation
Validate data format and content:

```python
# Python validation example
with open('data.json', 'r') as f:
    data = json.load(f)
    if not isinstance(data, dict):
        print("Invalid data format")
        sys.exit(1)
    if 'required_field' not in data:
        print("Missing required field")
        sys.exit(1)
```

## 4. Error Handling Strategies

### 4.1 Continue-on-Error Pattern
Use continue-on-error for non-critical steps:

```yaml
- name: Optional check
  id: optional-check
  continue-on-error: true
  run: some-check

- name: Handle failure
  if: steps.optional-check.outcome == 'failure'
  run: echo "Optional check failed, continuing..."
```

### 4.2 Dependency Failure Handling
Use needs conditions to handle dependency failures:

```yaml
jobs:
  job1:
    runs-on: ubuntu-latest
  job2:
    runs-on: ubuntu-latest
    needs: job1
    if: needs.job1.result == 'success'
```

### 4.3 Always-Run Summary
Use always() to ensure summary steps always run:

```yaml
jobs:
  summary:
    runs-on: ubuntu-latest
    needs: [job1, job2]
    if: always()
    steps:
      - name: Generate summary
        run: echo "Summary of results"
```

## 5. Context Variables

### 5.1 Common Context Variables
```yaml
- name: Show context
  run: |
    echo "Event: ${{ github.event_name }}"
    echo "Branch: ${{ github.ref_name }}"
    echo "Actor: ${{ github.actor }}"
    echo "Base ref: ${{ github.base_ref }}"
    echo "Head ref: ${{ github.head_ref }}"
```

### 5.2 PR-Specific Variables
Use PR-specific variables in PR workflows:

```yaml
- name: PR checks
  if: github.event_name == 'pull_request'
  run: |
    echo "PR number: ${{ github.event.pull_request.number }}"
    echo "Base branch: ${{ github.base_ref }}"
    echo "Head branch: ${{ github.head_ref }}"
    git diff --stat origin/${{ github.base_ref }}...HEAD
```

## 6. Script Patterns

### 6.1 Python Heredoc Pattern
Correctly use Python heredoc:

```yaml
- name: Python script
  run: |
    python3 << 'EOF'
    import sys
    try:
        # Python code here
        print("result=value")
    except Exception as e:
        print(f"Error: {e}")
        sys.exit(1)
    EOF

    echo "Processing result: $result"
```

### 6.2 JavaScript Optional Chaining Pattern
Use optional chaining in github-script:

```javascript
const workflowId = context.event?.workflow_run?.workflow_id;
if (!workflowId) {
  console.log('No workflow_run event data available');
  return;
}
```

### 6.3 Subprocess Exception Handling
Handle subprocess exceptions in Python scripts:

```python
try:
    result = subprocess.run(['git', 'command'], capture_output=True, text=True, check=True)
except subprocess.CalledProcessError as e:
    print(f"Error: {e}")
    sys.exit(1)
```

## 7. Permission Management

### 7.1 Least Privilege Principle
Grant only the minimum permissions required:

```yaml
permissions:
  contents: read
  pull-requests: read
  issues: write  # Only when creating issues is needed
```

### 7.2 Write Permission Configuration
Be explicit when write permissions are needed:

```yaml
# Note: This step requires contents: write permission
# Configure in Repository Settings > Actions > General
- name: Commit and push
  if: github.ref == 'refs/heads/main'
  run: |
    git config --local user.email "action@github.com"
    git config --local user.name "GitHub Action"
    git add file
    git commit -m "message"
    git push
```

## 8. Caching Strategies

### 8.1 Dependency Caching
Cache Go modules and build artifacts:

```yaml
- name: Cache Go modules
  uses: actions/cache@v4
  with:
    path: |
      ~/.cache/go-build
      ~/go/pkg/mod
    key: ${{ runner.os }}-go-${{ hashFiles('**/go.sum') }}
    restore-keys: |
      ${{ runner.os }}-go-
```

### 8.2 Build Caching
Cache build artifacts to speed up subsequent builds:

```yaml
- name: Cache build
  uses: actions/cache@v4
  with:
    path: build/
    key: ${{ runner.os }}-build-${{ github.sha }}
```

## 9. Workflow Organization

### 9.1 Separation of Concerns
Separate different responsibilities into different jobs:

```yaml
jobs:
  validate:
    # Data validation
  build:
    needs: validate
    # Build
  test:
    needs: build
    # Test
  deploy:
    needs: test
    # Deploy
```

### 9.2 Matrix Strategy
Use matrix strategy to test multiple configurations:

```yaml
strategy:
  matrix:
    os: [ubuntu-latest, windows-latest, macos-latest]
    go-version: ['1.26', '1.27']
```

## 10. Monitoring and Observability

### 10.1 Step Summary
Use GITHUB_STEP_SUMMARY for output summaries:

```yaml
- name: Generate summary
  run: |
    echo "# Test Results" >> $GITHUB_STEP_SUMMARY
    echo "- Total tests: 100" >> $GITHUB_STEP_SUMMARY
    echo "- Passed: 95" >> $GITHUB_STEP_SUMMARY
    echo "- Failed: 5" >> $GITHUB_STEP_SUMMARY
```

### 10.2 Artifact Upload
Upload build artifacts and test results:

```yaml
- name: Upload artifacts
  uses: actions/upload-artifact@v4
  with:
    name: test-results
    path: test-results/
    retention-days: 30
```

### 10.3 Failure Notification
Send notifications on critical failures:

```yaml
- name: Notify on failure
  if: failure()
  uses: actions/github-script@v7
  with:
    script: |
      github.rest.issues.create({
        owner: context.repo.owner,
        repo: context.repo.repo,
        title: 'Workflow failed',
        body: 'Please check the workflow logs.',
        labels: ['workflow-failure']
      })
```

## 11. Performance Optimization

### 11.1 Parallel Execution
Use parallel execution to speed up workflows:

```yaml
jobs:
  job1:
    runs-on: ubuntu-latest
  job2:
    runs-on: ubuntu-latest
  job3:
    runs-on: ubuntu-latest
    needs: [job1, job2]  # Execute job1 and job2 in parallel
```

### 11.2 Conditional Skip
Skip unnecessary steps:

```yaml
- name: Skip if no changes
  id: check-changes
  run: |
    if git diff --quiet; then
      echo "changed=false" >> $GITHUB_OUTPUT
    else
      echo "changed=true" >> $GITHUB_OUTPUT
    fi

- name: Process changes
  if: steps.check-changes.outputs.changed == 'true'
  run: process_changes
```

## 12. Security Practices

### 12.1 Secret Usage
Use secrets correctly:

```yaml
- name: Use secret
  env:
    API_KEY: ${{ secrets.API_KEY }}
  run: echo "Using API key"
```

### 12.2 Dependency Scanning
Regularly scan for dependency vulnerabilities:

```yaml
on:
  schedule:
    - cron: '0 0 * * 0'  # Every Sunday

jobs:
  security:
    runs-on: ubuntu-latest
    steps:
      - name: Run govulncheck
        run: govulncheck ./...
```

## 13. Debugging Tips

### 13.1 Enable Debug Logging
Enable debug logs when needed:

```yaml
- name: Enable debug logging
  run: |
    echo "::debug::Detailed debug information"
    echo "::warning::Warning message"
    echo "::error::Error message"
```

### 13.2 Preserve Build Artifacts on Failure
Keep build artifacts when failure occurs:

```yaml
- name: Upload build artifacts on failure
  if: failure()
  uses: actions/upload-artifact@v4
  with:
    name: build-artifacts
    path: build/
```

## 14. Workflow Documentation

### 14.1 Add Comments
Add clear comments in workflows:

```yaml
# Validate registry.yml structure and file references
# Triggered on push/PR when registry.yml changes
- name: Validate registry
  run: validate_registry
```

### 14.2 Update Handover Document
Update HANDOVER.md after workflow changes.

## 15. Testing Workflows

### 15.1 Local Testing
Use the act tool to test workflows locally:

```bash
act push -j validate-registry
```

### 15.2 Dry Run
Perform dry run checks before committing:

```bash
# Check workflow syntax
# Check path filters are correct
# Check conditional logic
```

## Summary

Following these best practices helps:
- Improve workflow reliability
- Reduce debugging time
- Improve maintainability
- Enhance security
- Optimize performance

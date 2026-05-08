#!/usr/bin/env python3
# Pre-commit hook to validate registry.yml
import subprocess
import sys
import os
import yaml
import glob

# Check if registry.yml is being modified
try:
    result = subprocess.run(['git', 'diff', '--cached', '--name-only'], capture_output=True, text=True, check=True)
except subprocess.CalledProcessError as e:
    print(f"[ERR] Error running git command: {e}")
    sys.exit(1)

registry_changed = any(
    f.endswith('registry.yml') or f.endswith('registry.yaml')
    for f in result.stdout.split('\n')
)
if not registry_changed:
    sys.exit(0)

# Find the actual registry file being modified
registry_file = '.github/config/registry.yml'
for line in result.stdout.split('\n'):
    if line.endswith('registry.yml'):
        registry_file = line
        break

print(f"Validating {registry_file}...")

try:
    with open(registry_file, 'r') as f:
        registry = yaml.safe_load(f)

    required_fields = ['id', 'name', 'version', 'status', 'category', 'file']
    workflow_ids = set()
    errors = []

    for workflow in registry.get('workflows', []):
        for field in required_fields:
            if field not in workflow:
                errors.append(f"[ERR] Workflow {workflow.get('id', 'unknown')} missing required field: {field}")

        if 'id' not in workflow or 'file' not in workflow:
            continue

        if workflow['id'] in workflow_ids:
            errors.append(f"[ERR] Duplicate workflow ID: {workflow['id']}")
        workflow_ids.add(workflow['id'])

        if not os.path.exists(workflow['file']):
            errors.append(f"[ERR] Workflow file does not exist: {workflow['file']}")
            if workflow['file'].startswith('docs/deprecated/'):
                suggestions = glob.glob('docs/deprecated/workflows/*.md')
                if suggestions:
                    errors.append(f"[HINT] Did you mean one of these?")
                    for s in suggestions[:3]:
                        errors.append(f"   - {s}")
            elif workflow['file'].startswith('.github/workflows/'):
                actual_files = glob.glob('.github/workflows/*.yml')
                if actual_files:
                    errors.append(f"[HINT] Available workflow files:")
                    for f in actual_files[:5]:
                        errors.append(f"   - {f}")

        if 'documentation' in workflow and not os.path.exists(workflow['documentation']):
            errors.append(f"[ERR] Documentation file does not exist: {workflow['documentation']}")
            if workflow['documentation'].startswith('/'):
                errors.append(f"[HINT] Absolute path detected. Use relative path from repository root.")
            if 'worklow' in workflow['documentation']:
                errors.append(f"[HINT] Typo detected: 'worklow' should be 'workflow'")

    if errors:
        print("[FAIL] Registry validation failed:")
        for error in errors:
            print(error)
        print("[FAIL] Pre-commit validation failed. Commit aborted.")
        sys.exit(1)

    print("[OK] Registry structure validation passed")
    print(f"[OK] Validated {len(workflow_ids)} workflows")
except Exception as e:
    print(f"[ERR] Error during validation: {e}")
    sys.exit(1)

sys.exit(0)

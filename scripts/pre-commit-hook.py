#!/usr/bin/env python3
# Pre-commit hook to validate registry.yml
import subprocess
import sys
import os
import yaml
import glob

# Check if registry.yml is being modified
result = subprocess.run(['git', 'diff', '--cached', '--name-only'], capture_output=True, text=True)
if '.github/registry.yml' not in result.stdout:
    sys.exit(0)

print("Validating registry.yml...")

try:
    with open('.github/registry.yml', 'r') as f:
        registry = yaml.safe_load(f)

    # Check required fields
    required_fields = ['id', 'name', 'version', 'status', 'category', 'file']
    workflow_ids = set()
    errors = []

    for workflow in registry.get('workflows', []):
        # Check required fields
        for field in required_fields:
            if field not in workflow:
                errors.append(f"❌ Workflow {workflow.get('id', 'unknown')} missing required field: {field}")

        # Skip further validation if required fields are missing
        if 'id' not in workflow or 'file' not in workflow:
            continue

        # Check for duplicate IDs
        if workflow['id'] in workflow_ids:
            errors.append(f"❌ Duplicate workflow ID: {workflow['id']}")
        workflow_ids.add(workflow['id'])

        # Check if file exists with better error messages
        if not os.path.exists(workflow['file']):
            errors.append(f"❌ Workflow file does not exist: {workflow['file']}")
            # Suggest similar files
            if workflow['file'].startswith('docs/deprecated/'):
                suggestions = glob.glob('docs/deprecated/workflows/*.md')
                if suggestions:
                    errors.append(f"💡 Did you mean one of these?")
                    for s in suggestions[:3]:
                        errors.append(f"   - {s}")
            elif workflow['file'].startswith('.github/workflows/'):
                actual_files = glob.glob('.github/workflows/*.yml')
                if actual_files:
                    errors.append(f"💡 Available workflow files:")
                    for f in actual_files[:5]:
                        errors.append(f"   - {f}")

        # Check if documentation exists with better error messages
        if 'documentation' in workflow and not os.path.exists(workflow['documentation']):
            errors.append(f"❌ Documentation file does not exist: {workflow['documentation']}")
            # Check for common typos
            if workflow['documentation'].startswith('/'):
                errors.append(f"💡 Absolute path detected. Use relative path from repository root.")
            if 'worklow' in workflow['documentation']:
                errors.append(f"💡 Typo detected: 'worklow' should be 'workflow'")

    if errors:
        print("❌ Registry validation failed:")
        for error in errors:
            print(error)
        print("❌ Pre-commit validation failed. Commit aborted.")
        print("   Fix the errors and try again.")
        sys.exit(1)

    print("✅ Registry structure validation passed")
    print(f"✅ Validated {len(workflow_ids)} workflows")
except Exception as e:
    print(f"❌ Error during validation: {e}")
    sys.exit(1)

sys.exit(0)

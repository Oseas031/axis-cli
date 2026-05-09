#!/bin/bash
# Install pre-commit hooks for the repository

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
HOOKS_DIR="$REPO_ROOT/.git/hooks"

echo "Installing pre-commit hooks..."

# Create hooks directory if it doesn't exist
mkdir -p "$HOOKS_DIR"

# Copy the pre-commit hook
cp "$SCRIPT_DIR/pre-commit-hook.py" "$HOOKS_DIR/pre-commit"

# Make it executable
chmod +x "$HOOKS_DIR/pre-commit"

echo "✅ Pre-commit hook installed successfully"
echo "   Location: $HOOKS_DIR/pre-commit"

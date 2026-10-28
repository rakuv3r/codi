#!/bin/bash
set -e

echo "Installing Git Hooks..."

# Configure Git to use hooks from project directory
git config core.hooksPath hooks

# Make hooks executable
chmod +x hooks/pre-commit

echo "✅ Git Hooks installed successfully"
echo "   Git configured to use hooks/ directory"
echo "   pre-commit: runs make lint"

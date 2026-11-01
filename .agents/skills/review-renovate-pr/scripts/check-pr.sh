#!/usr/bin/env bash
# Helper script to inspect and verify a Renovate PR locally
# Usage: ./check-pr.sh <pr-number>

set -euo pipefail

if [ $# -lt 1 ]; then
  echo "Usage: $0 <pr-number>" >&2
  exit 1
fi

PR="$1"

echo "=== 1. Inspecting PR #$PR ==="
gh pr view "$PR" --json number,title,author,headRefName,url,state \
  --template 'PR: #{{.number}} {{.title}}
Author: {{.author.login}}
Branch: {{.headRefName}}
State: {{.state}}
URL: {{.url}}
'

echo "=== 2. Changed Files ==="
gh pr diff "$PR" --name-only

echo ""
echo "=== 3. GitHub Checks ==="
gh pr checks "$PR" || true

echo ""
echo "=== 4. Diff Summary in go.mod / Workflows ==="
gh pr diff "$PR" | grep -E '^\+[^+]|^\-[^-]' | grep -E '(go\.mod|\.github|\+require|\-require|\+ *[a-zA-Z0-9\.\/_-]+ v|\- *[a-zA-Z0-9\.\/_-]+ v)' || true

echo ""
echo "To test this PR locally, run:"
echo "  gh pr checkout $PR"
echo "  go mod verify && go vet ./... && go test -v ./... && go build ./cmd/vger"

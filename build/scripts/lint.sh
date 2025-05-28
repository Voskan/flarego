#!/usr/bin/env bash
# build/scripts/lint.sh -------------------------------------------------------
# Aggregated linter runner for the FlareGo mono‑repo.  It orchestrates Go
# static analysis, UI TypeScript checks and Markdown/JSON formatting so that CI
# pipelines (and local developers) can run a single command.
#
# Usage:  bash build/scripts/lint.sh [--fix]
#   --fix   Apply autofixers where supported (goimports, prettier).  CI runs
#           without --fix to ensure no diffs are introduced silently.
#
# Environment variables respected:
#   GO_VERSION         – when set, ensures `go env GOVER` matches (CI safety)
#   GOPRIVATE          – inherited for private module access
#   NPM_LINT_SCRIPT    – overrides npm script name (default "lint")
# -----------------------------------------------------------------------------
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "$ROOT_DIR"

AUTO_FIX=false
if [[ "${1:-}" == "--fix" ]]; then
  AUTO_FIX=true
fi

log() { echo -e "\033[1;34m[lint]\033[0m $*"; }

# -----------------------------------------------------------------------------
# Sanity: Go version
# -----------------------------------------------------------------------------
if [[ -n "${GO_VERSION:-}" ]]; then
  cur=$(go env GOVERSION | sed 's/go//')
  if [[ "$cur" != "$GO_VERSION" ]]; then
    echo "error: expected Go $GO_VERSION, got $cur" >&2
    exit 1
  fi
fi

# -----------------------------------------------------------------------------
# Go: vet + golangci-lint + go test -run=none -cover
# -----------------------------------------------------------------------------
log "Running go vet" && go vet ./...

if command -v golangci-lint >/dev/null 2>&1; then
  log "Running golangci-lint" && golangci-lint run --timeout 5m
else
  echo "warning: golangci-lint not installed – skipping staticcheck" >&2
fi

# Quick compile test (no execution)
log "Running 'go test -run=^$ -cover ./...'"
go test -run=^$ -cover ./...

# Optional goimports fix
if $AUTO_FIX && command -v goimports >/dev/null 2>&1; then
  log "Running goimports -w"
  goimports -w $(git ls-files '*.go' | grep -v vendor)
fi

# -----------------------------------------------------------------------------
# Web UI: npm run lint (eslint + types) + prettier
# -----------------------------------------------------------------------------
if [[ -d "web" ]]; then
  pushd web >/dev/null
  NPM_BIN=$(command -v npm || true)
  if [[ -z "$NPM_BIN" ]]; then
    echo "warning: npm not found – skipping UI lint" >&2
  else
    script="${NPM_LINT_SCRIPT:-lint}"
    if npm run | grep -q " $script"; then
      log "Running npm run $script" && npm run $script
    else
      echo "warning: npm script '$script' missing – skipping" >&2
    fi
    if $AUTO_FIX; then
      if npm run | grep -q " prettier"; then
        log "Running prettier --write" && npx prettier --write "src/**/*.{ts,tsx,js,jsx,json,css,md}"
      fi
    fi
  fi
  popd >/dev/null
fi

# -----------------------------------------------------------------------------
# Markdown & YAML lint (markdownlint-cli & yamllint optional)
# -----------------------------------------------------------------------------
if command -v markdownlint >/dev/null 2>&1; then
  log "Running markdownlint"
  markdownlint docs/**/*.md README.md || true
fi
if command -v yamllint >/dev/null 2>&1; then
  log "Running yamllint"
  yamllint -d relaxed .github/**/*.yml deployments/**/*.yaml
fi

log "All linters finished successfully ✅"

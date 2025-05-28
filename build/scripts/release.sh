#!/usr/bin/env bash
# build/scripts/release.sh ----------------------------------------------------
# Automated release helper for the FlareGo project.  It wraps goreleaser,
# docker buildx and git tagging so maintainers can cut versions with a single
# command while CI (GitHub Actions release.yml) uses the same logic.
#
# Usage:
#   bash build/scripts/release.sh v1.2.3          # full release
#   bash build/scripts/release.sh --dry-run v1.2.3 # snapshot (no publish)
#
# Requirements:
#   • goreleaser installed (v1.23+) and on $PATH
#   • docker buildx (for multi‐arch images)
#   • $GITHUB_TOKEN set with repo:write & packages:write scopes (for goreleaser)
#   • $DOCKERHUB_USER / $DOCKERHUB_PASS for docker login (optional)
#
# The script performs:
#   1. Sanity checks (clean git tree, semver tag format).
#   2. Runs `go test`.
#   3. goreleaser (snapshot or actual) with provenance.
#   4. docker buildx build & push for agent + gateway.
#   5. git tag & push (unless --no-tag).
# -----------------------------------------------------------------------------
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "$ROOT_DIR"

DRY_RUN=false
NO_TAG=false
VERSION=""

while [[ $# -gt 0 ]]; do
  case $1 in
    --dry-run)
      DRY_RUN=true
      shift
      ;;
    --no-tag)
      NO_TAG=true
      shift
      ;;
    v*)
      VERSION="$1"
      shift
      ;;
    *)
      echo "Unknown arg: $1" >&2; exit 1;;
  esac
done

if [[ -z "$VERSION" ]]; then
  echo "usage: release.sh [--dry-run] [--no-tag] vX.Y.Z" >&2
  exit 1
fi

# Verify semver-ish
if ! [[ $VERSION =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
  echo "error: version must match vMAJOR.MINOR.PATCH" >&2; exit 1
fi

# Ensure clean git state
if [[ -n $(git status --porcelain) ]]; then
  echo "error: git tree is dirty" >&2; exit 1
fi

# Run tests
echo "[release] Running unit tests"
go test ./...

echo "[release] Running goreleaser"
if $DRY_RUN; then
  goreleaser release --snapshot --clean --skip-publish --rm-dist
else
  : "${GITHUB_TOKEN:?GITHUB_TOKEN not set}"
  goreleaser release --clean --rm-dist
fi

# Docker images
PLAT="linux/amd64,linux/arm64"
for img in agent gateway; do
  NAME="flarego/${img}:${VERSION#v}"
  echo "[release] Building Docker image $NAME"
  if $DRY_RUN; then
    docker buildx build --platform $PLAT -f build/Dockerfile.${img} --tag $NAME --load .
  else
    docker buildx build --platform $PLAT -f build/Dockerfile.${img} --tag $NAME --push .
  fi
done

# Git tag
if ! $NO_TAG && ! $DRY_RUN; then
  if git rev-parse "$VERSION" >/dev/null 2>&1; then
    echo "Tag $VERSION already exists, skipping"
  else
    echo "[release] Creating git tag $VERSION"
    git tag -a "$VERSION" -m "Release $VERSION"
    git push origin "$VERSION"
  fi
fi

echo "[release] Done ✅"

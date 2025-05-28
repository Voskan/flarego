#!/usr/bin/env bash
# build/scripts/generate-proto.sh
# -----------------------------------------------------------------------------
# Regenerates Go sources from all .proto files in internal/proto.
#
# Usage (from repo root):
#     bash build/scripts/generate-proto.sh
#
# Prerequisites:
#   • protoc                — https://github.com/protocolbuffers/protobuf
#   • protoc-gen-go         — google.golang.org/protobuf/cmd/protoc-gen-go (v1.32+)
#   • protoc-gen-go-grpc    — google.golang.org/grpc/cmd/protoc-gen-go-grpc (v1.3+)
#
# The script is idempotent and safe to run in CI; it exits with non‑zero status
# if generation fails or required tools are missing.  Generated *.pb.go files
# are placed next to their source .proto with source_relative paths so import
# paths remain stable.
# -----------------------------------------------------------------------------
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
PROTO_DIR="${ROOT_DIR}/internal/proto"
OUT_DIR="${PROTO_DIR}"

PROTOC_BIN="${PROTOC:-protoc}"
GO_PLUGIN_BIN="${PROTOC_GEN_GO:-protoc-gen-go}"
GRPC_PLUGIN_BIN="${PROTOC_GEN_GO_GRPC:-protoc-gen-go-grpc}"

# -----------------------------------------------------------------------------
# Sanity‑check toolchain.
# -----------------------------------------------------------------------------
for tool in "$PROTOC_BIN" "$GO_PLUGIN_BIN" "$GRPC_PLUGIN_BIN"; do
    if ! command -v "$tool" >/dev/null 2>&1; then
        echo "error: required tool '$tool' not found in PATH" >&2
        exit 1
    fi
done

# -----------------------------------------------------------------------------
# Generate code for each proto file.
# -----------------------------------------------------------------------------
find "$PROTO_DIR" -maxdepth 1 -name '*.proto' | while read -r proto; do
    rel="${proto#$ROOT_DIR/}"
    echo "[proto] Generating $rel"
    "$PROTOC_BIN" \
        --proto_path="$PROTO_DIR" \
        --go_out="$OUT_DIR" --go_opt=paths=source_relative \
        --go-grpc_out="$OUT_DIR" --go-grpc_opt=paths=source_relative \
        "$proto"
    echo "[proto] Done $rel"
done

echo "✅ Protobuf generation completed"

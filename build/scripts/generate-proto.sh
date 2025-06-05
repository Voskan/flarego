#!/usr/bin/env bash
# build/scripts/generate-proto.sh
# -----------------------------------------------------------------------------
# Regenerates Go and TypeScript sources from all .proto files in internal/proto.
#
# Usage (from repo root):
#     bash build/scripts/generate-proto.sh
#
# Prerequisites:
#   • protoc                — https://github.com/protocolbuffers/protobuf
#   • protoc-gen-go         — google.golang.org/protobuf/cmd/protoc-gen-go (v1.32+)
#   • protoc-gen-go-grpc    — google.golang.org/grpc/cmd/protoc-gen-go-grpc (v1.3+)
#   • @bufbuild/protoc-gen-es
#   • @bufbuild/protoc-gen-connect-es
#
# The script is idempotent and safe to run in CI; it exits with non‑zero status
# if generation fails or required tools are missing.  Generated files are placed
# next to their source .proto with source_relative paths so import paths remain
# stable.
# -----------------------------------------------------------------------------
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
PROTO_DIR="${ROOT_DIR}/internal/proto"
GO_OUT_DIR="${ROOT_DIR}/internal/proto"
TS_OUT_DIR="${ROOT_DIR}/web/src/gen"

echo "root dir: $ROOT_DIR"
echo "proto dir: $PROTO_DIR"
echo "go out dir: $GO_OUT_DIR"
echo "ts out dir: $TS_OUT_DIR"

PROTOC_BIN="${PROTOC:-protoc}"
GO_PLUGIN_BIN="${PROTOC_GEN_GO:-protoc-gen-go}"
GRPC_PLUGIN_BIN="${PROTOC_GEN_GO_GRPC:-protoc-gen-go-grpc}"
ES_PLUGIN_BIN="${ROOT_DIR}/node_modules/.bin/protoc-gen-es"
CONNECT_ES_PLUGIN_BIN="${ROOT_DIR}/node_modules/.bin/protoc-gen-connect-es"

# -----------------------------------------------------------------------------
# Sanity‑check toolchain.
# -----------------------------------------------------------------------------
for tool in "$PROTOC_BIN" "$GO_PLUGIN_BIN" "$GRPC_PLUGIN_BIN"; do
    if ! command -v "$tool" >/dev/null 2>&1; then
        echo "error: required tool '$tool' not found in PATH" >&2
        exit 1
    fi
done

if [ ! -x "$ES_PLUGIN_BIN" ]; then
    echo "error: required tool '$ES_PLUGIN_BIN' not found or not executable" >&2
    exit 1
fi

if [ ! -x "$CONNECT_ES_PLUGIN_BIN" ]; then
    echo "error: required tool '$CONNECT_ES_PLUGIN_BIN' not found or not executable" >&2
    exit 1
fi

# -----------------------------------------------------------------------------
# Generate code for each proto file.
# -----------------------------------------------------------------------------
find "$PROTO_DIR" -maxdepth 1 -name '*.proto' | while read -r proto; do
    rel="${proto#$ROOT_DIR/}"
    echo "[proto] Generating $rel"
    
    # Generate Go code
    "$PROTOC_BIN" \
        --proto_path="$PROTO_DIR" \
        --go_out="$GO_OUT_DIR" --go_opt=paths=source_relative \
        --go-grpc_out="$GO_OUT_DIR" --go-grpc_opt=paths=source_relative \
        "$proto"
    
    # Generate TypeScript code
    "$PROTOC_BIN" \
        --proto_path="$PROTO_DIR" \
        --plugin=protoc-gen-es="$ES_PLUGIN_BIN" \
        --plugin=protoc-gen-connect-es="$CONNECT_ES_PLUGIN_BIN" \
        --es_out=import_extension=ts:"$TS_OUT_DIR" \
        --connect-es_out=import_extension=ts:"$TS_OUT_DIR" \
        "$proto"
    
    echo "[proto] Done $rel"
done

echo "✅ Protobuf generation completed"

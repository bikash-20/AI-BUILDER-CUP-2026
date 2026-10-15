#!/usr/bin/env bash
# gen_proto.sh — generate Go + TypeScript types from .proto files.
#
# Currently a stub: FinGuard AI uses REST (OpenAPI 3.0) instead of gRPC.
# This script is kept as a placeholder for future gRPC services.
# When the first .proto is added:
#   1. Add it under proto/
#   2. Install buf: https://buf.build/docs/installation
#   3. Wire `buf generate` here for Go and TS targets.
#
# Example for future implementation:
#
#   buf generate \
#     --template buf.gen.yaml \
#     --path proto/fraud/v1 \
#     --include-imports

set -euo pipefail

PROTO_DIR="$(dirname "$0")/../proto"

if [ ! -d "$PROTO_DIR" ]; then
  echo "No proto/ directory yet. Skipping."
  exit 0
fi

if ! command -v buf >/dev/null 2>&1; then
  echo "buf not installed. Install: https://buf.build/docs/installation"
  exit 0
fi

cd "$PROTO_DIR"
buf generate
echo "✓ Generated code from protos."

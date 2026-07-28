#!/usr/bin/env bash
# Builds the specguard binary and the UI, then serves the example project as a
# single origin on :8138 — the deterministic target for the Playwright e2e run.
set -euo pipefail

web="$(cd "$(dirname "$0")/.." && pwd)"
repo="$(cd "$web/.." && pwd)"
# Use whatever `go` is on PATH (CI provides it); fall back to a common install.
command -v go >/dev/null 2>&1 || export PATH="/usr/local/go/bin:$PATH"

go build -o "$web/.e2e-specguard" "$repo/cmd/specguard"
npm --prefix "$web" run build >/dev/null

exec "$web/.e2e-specguard" serve -C "$repo/example" -web "$web/dist" -addr :8138

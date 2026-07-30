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

# Also build the single-file UI and generate a static report from the freshly
# built template, so report.spec.ts exercises the real `specguard report`
# pipeline end to end (not the committed embed).
npm --prefix "$web" run build:single >/dev/null
"$web/.e2e-specguard" report -C "$repo/example" \
  -web "$web/dist-single/index.html" \
  -branch e2e-branch -commit 0123456789abcdef -repo clems4ever/specguard \
  -o "$web/.e2e-report.html"

exec "$web/.e2e-specguard" serve -C "$repo/example" -web "$web/dist" -addr :8138

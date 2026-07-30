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

# A second report WITH an ingested test run: the example's real `go test -json`
# (all pass) plus a committed Playwright fixture that fails one spec, so
# report-results.spec.ts exercises the covered-but-failing state end to end.
(cd "$repo/example" && go test -json ./... > "$web/.e2e-go-results.json" 2>/dev/null) || true
"$web/.e2e-specguard" report -C "$repo/example" \
  -web "$web/dist-single/index.html" \
  -branch e2e-branch -commit 0123456789abcdef -repo clems4ever/specguard \
  -results "$web/.e2e-go-results.json,$web/e2e/fixtures/playwright-results.json" \
  -o "$web/.e2e-report-results.html"

# A third report WITH a screenshot attachment, so report-artifacts.spec.ts
# exercises the per-spec gallery end to end (real report + real image copied to
# an assets dir, referenced by relative URL). A committed PNG stands in for a
# Playwright screenshot; the JSON is written here so its path is absolute.
cat > "$web/.e2e-shots.json" <<JSON
{"suites":[{"specs":[{"title":"logs in","tags":["@spec:auth-login"],
  "tests":[{"results":[{"status":"passed","attachments":[
    {"name":"login screen","contentType":"image/png","path":"$web/e2e/fixtures/login.png"}
  ]}]}]}]}]}
JSON
"$web/.e2e-specguard" report -C "$repo/example" \
  -web "$web/dist-single/index.html" \
  -branch e2e-branch -commit 0123456789abcdef -repo clems4ever/specguard \
  -results "$web/.e2e-shots.json" \
  -assets "$web/.e2e-assets" -assets-base ".e2e-assets" \
  -o "$web/.e2e-report-artifacts.html"

exec "$web/.e2e-specguard" serve -C "$repo/example" -web "$web/dist" -addr :8138

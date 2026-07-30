---
id: lint-ref-locations
title: Each reference records its file and line
parent: lint-traceability
covers:
  - internal/lint
---
## Behaviour

Every `spec:<id>` reference is recorded with its file and 1-based line, and a Go reference also carries the test function it sits above.

## Why

Precise locations are what let the report deep-link to the exact covering test and correlate `go test` results back to a spec.

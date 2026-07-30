---
id: results-go-func
title: Go results map by test function
parent: results-overlay
covers:
  - internal/results
---
## Behaviour

A `go test -json` stream maps each test's outcome to a spec via the test function the `// spec:<id>` comment sits above.

## Why

Go's machine output is keyed by function name, so the comment-to-function association is what closes the loop to a spec.

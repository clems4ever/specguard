---
id: results-format-autodetect
title: Result formats are auto-detected
covers:
  - internal/results
---
## Behaviour

Each results file is classified as Playwright JSON or `go test -json` by its shape, so a mixed set can be passed together.

## Why

One flag that 'just works' across runners is friendlier than making the user declare each file's format.

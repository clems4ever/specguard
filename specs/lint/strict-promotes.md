---
id: lint-strict-promotes
title: Strict mode promotes warnings to errors
covers:
  - internal/lint
---
## Behaviour

Under `-strict`, warnings such as `covers-unmatched` become errors and fail the run.

## Why

Some projects want zero tolerance; strict mode lets the same rules gate a build hard.

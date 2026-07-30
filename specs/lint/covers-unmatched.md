---
id: lint-covers-unmatched
title: A covers entry matching no file warns
covers:
  - internal/lint
---
## Behaviour

A `covers:` glob that matches no file in the tree is a warning (promoted to an error under `-strict`).

## Why

A covers entry pointing at code that no longer exists is stale traceability worth surfacing.

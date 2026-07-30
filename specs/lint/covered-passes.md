---
id: lint-covered-passes
title: A spec referenced by a test is covered
covers:
  - internal/lint
---
## Behaviour

When at least one test carries a `spec:<id>` reference for a spec, that spec is marked covered and the run passes.

## Why

Coverage is the core signal: a documented behaviour is only trustworthy once a test is pinned to it.

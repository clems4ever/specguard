---
id: lint-duplicate-id
title: Duplicate spec ids are rejected
covers:
  - internal/lint
---
## Behaviour

Two spec files sharing the same `id` produce a `duplicate-id` error.

## Why

Ids are the join key between specs and tests; a collision makes coverage ambiguous.

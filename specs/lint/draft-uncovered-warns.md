---
id: lint-draft-uncovered-warns
title: A draft spec may be uncovered
covers:
  - internal/lint
---
## Behaviour

A spec with `status: draft` that has no covering test yet is a warning, not a failure.

## Why

Drafts capture planned behaviour before the test exists; they should be visible without blocking the build.

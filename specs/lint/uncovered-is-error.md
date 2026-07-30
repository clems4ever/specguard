---
id: lint-uncovered-is-error
title: An uncovered spec fails the build
parent: lint-traceability
covers:
  - internal/lint
---
## Behaviour

A non-draft spec that no test references is reported as an `uncovered-spec` error, and the run exits non-zero.

## Why

An uncovered spec is a promise with nothing holding it true; the build should refuse to go green on it.

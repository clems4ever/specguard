---
id: diff-regression
title: The delta flags coverage regressions
covers:
  - internal/diff
---
## Behaviour

Diffed against a base ref, a spec that arrives uncovered (or loses its coverage) is reported as a regression.

## Why

A reviewer should be able to trust one number — regressions — instead of re-reading the whole spec set on every change.

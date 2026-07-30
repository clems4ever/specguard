---
id: lint-traceability
title: Traceability is enforced
covers:
  - internal/lint
---
## Behaviour

Every spec must be proven by a real test and every reference must resolve to a
defined spec, or the build fails. The specs below are the individual rules that
enforce this.

## Why

Traceability is the product's core promise: if it isn't mechanically enforced,
specs and code drift apart and the catalog stops being trustworthy.

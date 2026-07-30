---
id: lint-undefined-reference
title: A reference to an unknown spec is an error
covers:
  - internal/lint
---
## Behaviour

A `spec:<id>` reference in a test that resolves to no defined spec is an `undefined-reference` error.

## Why

A dangling reference means a test claims to cover a behaviour that isn't documented — usually a typo or a deleted spec.

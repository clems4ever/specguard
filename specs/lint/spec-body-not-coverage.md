---
id: lint-spec-body-not-coverage
title: A token in a spec body is not coverage
covers:
  - internal/lint
---
## Behaviour

A `spec:<id>` token appearing inside a spec's own markdown body does not count as a covering reference.

## Why

Prose about a spec shouldn't be mistaken for a test that enforces it, or every spec could trivially 'cover' itself.

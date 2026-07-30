---
id: spec-hierarchy
title: Specs derive from one another
parent: spec-model
covers:
  - internal/spec/spec.go
  - internal/lint/lint.go
---
## Behaviour

A spec may declare a `parent`, forming a tree from broad intents down to leaf
behaviours. A leaf must be covered by a real test; a parent's coverage is
*derived* — it needs no test of its own and is satisfied when all its children
are. An undefined parent or a cycle in the parent chain fails the build.

## Why

A flat list captures leaf assertions but not intent. Deriving specs lets a
high-level capability be the thing a human or agent reasons about, with the
concrete tests as its proof — and makes it clear why a low-level behaviour
exists and what breaks if it changes.

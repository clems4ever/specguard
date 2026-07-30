---
id: ui-area-overview
title: Each area can describe itself
parent: ui-report
covers:
  - internal/spec/spec.go
  - web/src/components/Dashboard.tsx
---
## Behaviour

An area of specs can carry a human overview — a title and a one-line
description — written in an optional `specs/<area>/_area.md` file. When present,
the report shows it beneath the area heading, so a reader learns what a group of
specs is about without opening any of them. The overview file is not a spec and
never affects the build.

## Why

Grouping specs by area only helps if the reader knows what each area is for. A
bare directory name (`lint`, `results`) is opaque; a sentence of context turns
the grouping into a table of contents a newcomer can follow.

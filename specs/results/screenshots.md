---
id: results-screenshots
title: Screenshots attach to their spec
covers:
  - internal/results
---
## Behaviour

Image attachments on a `@spec:<id>`-tagged Playwright test are collected as that spec's artifacts; non-image attachments are ignored.

## Why

A screenshot is visual proof of a behaviour — attaching it to the spec turns the report into something a PM can look at.

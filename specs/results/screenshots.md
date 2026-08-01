---
id: results-screenshots
title: Screenshots and videos attach to their spec
parent: results-overlay
covers:
  - internal/results
---
## Behaviour

Image and video attachments on a `@spec:<id>`-tagged Playwright test are
collected as that spec's artifacts, in order; other attachment types (e.g. a
trace zip) are ignored.

## Why

A screenshot or a short clip is visual proof of a behaviour — attaching it to the
spec turns the report into something a PM can watch, not just read.

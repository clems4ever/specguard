---
id: results-playwright-tag
title: Playwright results map by tag
covers:
  - internal/results
---
## Behaviour

A Playwright JSON report maps each test's outcome to a spec via its `@spec:<id>` tag.

## Why

The tag is already the coverage link; reusing it means test outcomes attach to specs with no extra wiring.

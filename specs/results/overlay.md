---
id: results-overlay
title: A real test run overlays onto the specs
covers:
  - internal/results
---
## Behaviour

Given the output of a test run, specguard maps each result back to the spec it
proves, so a spec reads pass or fail. The specs below are the parts of that
mapping.

## Why

Coverage says a test *exists*; overlaying its result says the behaviour is
actually *proven right now* — the difference a PM cares about.

---
id: diff-review
title: A change can be reviewed in isolation
covers:
  - internal/diff
---
## Behaviour

Given a base git ref, specguard reports only the specs a change touched — what
gained or lost coverage, and which covered code changed — so review focuses on
the delta rather than the whole catalog. The specs below are the parts of that.

## Why

Re-reading every spec after each change does not scale; showing just what moved
is what makes the check usable on real pull requests.

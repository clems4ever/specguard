---
id: diff-impl-changed
title: A spec whose covered code changed is flagged
parent: diff-review
covers:
  - internal/diff
---
## Behaviour

When a change touches a file a spec `covers:`, that spec is marked impl-changed even if its definition and coverage are untouched.

## Why

The behaviour's implementation moved under it — worth a reviewer's eyes even though traceability is intact.

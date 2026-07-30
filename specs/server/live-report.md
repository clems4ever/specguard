---
id: server-live-report
title: The server returns a live report
parent: server-live
covers:
  - internal/server
---
## Behaviour

`specguard serve` recomputes and returns the report as JSON on every request, so edits on disk show up on refresh.

## Why

The live UI is only useful if it reflects the working tree as it is now, not a stale snapshot.

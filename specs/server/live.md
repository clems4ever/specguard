---
id: server-live
title: The report can be served live
covers:
  - internal/server
---
## Behaviour

`specguard serve` exposes the report over HTTP while you work: the live report,
a coverage badge, and a single-page-app fallback so deep links resolve on a hard
refresh. The specs below are those endpoints and behaviours.

## Why

A live server turns the report from a snapshot into a dashboard you can keep
open and refresh as the repo changes.

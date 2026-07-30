---
id: report-catalog
title: A self-contained report is produced
covers:
  - internal/report
---
## Behaviour

`specguard report` turns a run into a single, shareable HTML catalog: the specs,
their data embedded so it works offline, screenshots published beside it, and
branch-safe slugs. The specs below are the guarantees that make that report
correct and safe.

## Why

A report is only trustworthy if it is complete, self-contained, and can't be
corrupted by the very data it renders.

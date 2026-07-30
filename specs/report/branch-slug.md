---
id: report-branch-slug
title: Branch names slug without collisions
covers:
  - internal/report
---
## Behaviour

A branch name is encoded to a filesystem-safe directory segment reversibly, so distinct branches never map to the same per-branch report.

## Why

Per-branch publishing must never silently overwrite one branch's report with another's.

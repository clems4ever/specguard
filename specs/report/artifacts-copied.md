---
id: report-artifacts-copied
title: Screenshots are published beside the report
covers:
  - internal/report
---
## Behaviour

With `-assets`, each spec's screenshots are copied under `assets/specs/<id>/` and referenced by relative URL; a missing source is skipped, not fatal.

## Why

Turning the single file into a bundle is how galleries reach GitHub Pages without inlining megabytes of images.

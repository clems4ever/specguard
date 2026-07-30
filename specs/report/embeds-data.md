---
id: report-embeds-data
title: The report embeds its data
covers:
  - internal/report
---
## Behaviour

`specguard report` splices the lint report into the UI template as JSON on `window`, so the page renders offline with no server.

## Why

A self-contained file can be committed, emailed, or published to Pages and still work — that's the whole point of the report.

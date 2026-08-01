---
id: ui-live-preview
title: A spec deep-links to the live feature
parent: ui-report
covers:
  - web/src/links.ts
  - web/src/components/SpecDetail.tsx
---
## Behaviour

A spec can declare a `preview` path where its behaviour runs. When the report is
generated with a preview base (a per-PR deploy), each such spec shows a "Try it
live" link that opens the running feature at that path — so a PM jumps straight
from the spec to the behaviour to review it, without touching code.

## Why

A PM reviews by *using* the feature, not by reading a diff. Binding a spec to the
exact screen in a live preview is what makes that review one click.

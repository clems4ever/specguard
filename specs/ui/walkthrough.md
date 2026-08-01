---
id: ui-walkthrough
title: Captured evidence reads as a captioned walkthrough
parent: ui-report
covers:
  - web/src/components/Gallery.tsx
---
## Behaviour

A spec's captured artifacts render as an ordered walkthrough: each capture shows
its attachment name as a caption, so `Given… / When… / Then…` steps read as a
concrete example. Screenshots open in a lightbox; a video attachment renders as
an inline player.

## Why

Prose says what a behaviour should do; a captioned sequence of the real screens
(and a clip) shows what it actually does — the context a reviewer needs to judge
it without reading code.

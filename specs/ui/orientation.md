---
id: ui-orientation
title: The report orients a first-time reader
parent: ui-report
covers:
  - web/src/components/Intro.tsx
  - web/src/components/Dashboard.tsx
---
## Behaviour

The dashboard opens with a short explanation of what specguard is and a legend
for the status badges, so a reader who has never seen the product understands
what a row means. Specs are grouped into areas that can be collapsed and that
carry a count and a roll-up status, and a table-of-contents lets the reader jump
to an area — so the catalog can be scanned by section, not read row by row.

## Why

A wall of identical rows is overwhelming and forces item-by-item reading. An
orientation layer — what this is, how to read it, and the shape of the whole —
lets a newcomer grasp the report at a glance before diving in.

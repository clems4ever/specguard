---
id: results-worst-wins
title: A spec fails if any covering test fails
covers:
  - internal/results
---
## Behaviour

A spec's aggregate result is worst-wins: it is failing if any of its covering tests failed, passing only when all ran and none failed.

## Why

'Covered but failing' is the most important state to surface; one red test must not be hidden behind a green sibling.

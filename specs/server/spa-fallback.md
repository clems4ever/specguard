---
id: server-spa-fallback
title: Unknown routes fall back to the app
covers:
  - internal/server
---
## Behaviour

A request for an unknown path serves the built `index.html`, so a hard refresh on a client-side route (e.g. `/spec/<id>`) still loads the app.

## Why

Deep links into a spec must survive a refresh for the report to be shareable.

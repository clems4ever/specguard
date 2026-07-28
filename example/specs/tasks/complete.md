---
id: tasks-complete
title: Toggling a task persists its completed state
status: active
covers:
  - server/tasks.go
  - web/e2e/tasks.spec.ts
---

## Behaviour

Checking a task marks it complete and the state survives a reload. Unchecking
reverts it. The change is optimistic in the UI but confirmed by the server.

## Why

A checkbox that silently forgets on reload destroys trust in the whole app.

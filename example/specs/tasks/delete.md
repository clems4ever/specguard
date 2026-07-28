---
id: tasks-delete
title: Deleting a task asks for confirmation
status: active
covers:
  - server/tasks.go
  - web/e2e/tasks.spec.ts
---

## Behaviour

Delete opens a confirmation; cancelling keeps the task, confirming removes it
for good. Only the task's owner may delete it.

## Why

Deletion is irreversible; a stray click must not lose work, and ownership stops
one user erasing another's task.

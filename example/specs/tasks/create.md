---
id: tasks-create
title: A user can create a task with a title
status: active
covers:
  - server/tasks.go
  - web/e2e/tasks.spec.ts
---

## Behaviour

Submitting a non-empty title creates a task owned by the current user and shows
it at the top of the list. An empty title is rejected inline.

## Why

Creation is the core loop; an empty-title task is noise that the spec forbids.

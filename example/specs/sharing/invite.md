---
id: sharing-invite
title: A task owner can invite another user by email
status: active
covers:
  - server/sharing.go
  - web/e2e/sharing.spec.ts
---

## Behaviour

The owner enters an email; that user gains read access and sees the task in
their list. Inviting an unknown email queues a pending invite rather than failing.

## Why

Sharing is the reason Taskflow beats a private list; a hard failure on unknown
emails would block the most common invite path.

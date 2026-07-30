---
id: sharing-permissions
title: Shared collaborators get read-only access unless promoted to editor
status: draft
covers:
  - server/sharing.go
---

## Behaviour

An invited collaborator can view a shared task but cannot edit or delete it until
the owner promotes them to editor. Editors may change the title and completed
state but still cannot delete.

## Why

Sharing without a permission model means any collaborator can wipe the owner's
task. This behaviour is specified but not yet built — it is a **draft**, so
specguard reports it as planned work rather than failing the build.

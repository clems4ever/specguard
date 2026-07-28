---
id: auth-logout
title: Logging out invalidates the session server-side
status: active
covers:
  - server/auth.go
  - web/e2e/auth.spec.ts
---

## Behaviour

Logout clears the cookie and marks the session revoked on the server, so a
replayed cookie is rejected even before it would have expired.

## Why

Client-side cookie deletion alone is not logout — a captured cookie must stop
working the instant the user logs out.

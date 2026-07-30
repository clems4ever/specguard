---
id: auth-session-expiry
title: Sessions expire after 24h of inactivity
parent: auth-access
status: active
covers:
  - server/auth.go
---

## Behaviour

A session unused for 24 hours is rejected and the user is sent back to login.
Activity slides the window forward.

## Why

Bounds the blast radius of a leaked cookie. Verified server-side because the
clock the client reports cannot be trusted.

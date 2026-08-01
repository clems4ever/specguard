---
id: auth-login
title: A user can log in with email and password
parent: auth-access
preview: /login
status: active
covers:
  - server/auth.go
  - web/e2e/auth.spec.ts
---

## Behaviour

Posting valid credentials returns a session cookie and redirects to the task
list. Invalid credentials return `401` and never reveal whether the email exists.

## Why

Login is the front door. A vague error that distinguishes "unknown email" from
"wrong password" leaks account existence; the spec pins the non-disclosure.

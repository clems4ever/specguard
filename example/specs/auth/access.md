---
id: auth-access
title: Users can access their account
covers:
  - server/auth.go
---
## Behaviour

The account is gated by authentication: a user signs in, stays signed in for a
session, and can sign out. The specs below are the concrete behaviours that make
that true.

## Why

"Access to the account" is the capability a product owner cares about; the login,
session and logout specs are how it is proven.

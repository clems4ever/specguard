---
id: spec-acceptance
title: A behaviour is accepted once, then guarded
covers:
  - internal/lint/acceptance.go
---
## Behaviour

Each spec has a *fingerprint* of its expectation — its intent plus the source of
its covering tests. A PM records acceptance against that fingerprint once. While
the fingerprint is unchanged the acceptance holds and the covering tests guard
non-regression automatically; a pure code refactor never disturbs it. Changing
the intent or a covering test changes the fingerprint, so that one spec becomes
`stale` and routes back for re-review. `specguard accept --check` fails while any
implemented spec is unaccepted at its current fingerprint.

## Why

The PM's judgement is expensive and must not be re-spent on every change. Binding
a one-time behavioural sign-off to a fingerprint — and letting the tests carry it
forward — is what makes review scale as agents do most of the work.

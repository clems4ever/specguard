---
id: spec-frontmatter-invalid
title: Malformed frontmatter is rejected
covers:
  - internal/spec
---
## Behaviour

A spec missing its `id` (or with a malformed id, bad status, or an unterminated frontmatter block) is a parse error rather than a silently-empty spec.

## Why

A spec that can't be identified can't be traced to a test — failing loudly beats a spec that quietly covers nothing.

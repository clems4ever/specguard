---
id: spec-model
title: A spec is a parsed markdown file
covers:
  - internal/spec
---
## Behaviour

A spec is plain markdown with frontmatter declaring a stable id, a title, the
code it governs and (optionally) a parent. specguard parses it, rejects
malformed frontmatter, and derives specs from one another. The specs below are
the parts of that model.

## Why

The whole tool rests on a simple, dependency-free file format a human can write
and an agent can read — so the format's rules must themselves be specified.

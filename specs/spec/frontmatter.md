---
id: spec-frontmatter
title: A spec is parsed from frontmatter
covers:
  - internal/spec
---
## Behaviour

A spec file is `id`, `title`, optional `status` and `covers` in a leading `---` block, followed by a free-form markdown body. All of it is parsed and the body is carried through verbatim.

## Why

The spec file is the human-readable half of the contract; the report and the linter both rely on this shape being parsed faithfully.

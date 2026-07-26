# specguard

A tiny, dependency-free traceability linter for **specification-by-example**.

You keep a directory of short spec files — one stable idea each, written for
humans and agents. Your tests link back to a spec by id. `specguard` enforces
the link both ways:

- every spec is referenced by **at least one test** (nothing is merely
  documented but unverified);
- every `spec:<id>` reference in a test points at a **real spec** (no dangling
  links after a rename).

That gives an autonomous agent a durable, machine-checkable map of *what the
repo is supposed to do* — even in the face of a contradictory new request — and
gives a human a one-glance overview of which behaviours are pinned by tests.

It is language-agnostic: a spec is claimed by dropping the token `spec:<id>`
anywhere in a test file — a Playwright tag (`@spec:skills-edit`) or a Go comment
(`// spec:skills-edit`) both work.

## The two artifacts

A **spec** is markdown with frontmatter. The body is prose; specguard only reads
the frontmatter.

```markdown
---
id: skills-edit
title: Editing a SKILL.md persists to disk and updates the catalog
status: active
covers:
  - internal/skill
  - web/src/skills
  - web/e2e/skills.spec.ts
---

## Behaviour
Opening a skill, editing its `SKILL.md`, and saving writes the file to disk and
re-reads it, so the change survives a reload and the new description shows on the
catalog row.

## Why
Skills are the unit an agent reuses; a silent failure to persist an edit means an
agent keeps running stale instructions.
```

A **test** claims the spec by id:

```ts
test('edits the SKILL.md and persists the change', { tag: '@spec:skills-edit' }, async ({ page }) => { … })
```

```go
// spec:skills-edit
func TestCatalogSave(t *testing.T) { … }
```

`covers` is optional; when present, each entry must match a real file, so a spec
can't quietly point at code that was deleted or moved.

## Usage

```
specguard            # lint ./ using ./.specguard.yml
specguard -C ../repo # lint another tree
specguard -json      # machine-readable report (feed a spec-overview UI)
specguard -strict    # treat warnings (e.g. covers-unmatched) as errors
```

Exit code is `0` when the check passes and `1` when it fails, so it drops
straight into CI.

## Config — `.specguard.yml`

```yaml
specsDir: specs
tests:
  - "internal/**/*_test.go"
  - "web/e2e/**/*.spec.ts"
  - "web/src/**/*.test.ts"
```

Defaults (used when no config is present): `specsDir: specs` and a broad set of
`*_test.go` / `*.spec.ts` / `*.test.ts(x)` globs.

## Checks

| rule | severity | meaning |
|------|----------|---------|
| `uncovered-spec`      | error   | a spec no test references |
| `undefined-reference` | error   | a test references a `spec:<id>` that doesn't exist |
| `duplicate-id`        | error   | two spec files share an id |
| `parse-error`         | error   | a spec's frontmatter is malformed |
| `covers-unmatched`    | warning | a `covers` entry matches no file (error under `-strict`) |

## Build

```
go build -o specguard ./cmd/specguard
```

Zero third-party dependencies; only the Go standard library.

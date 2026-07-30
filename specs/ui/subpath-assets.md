---
id: ui-subpath-assets
title: Screenshots load when the report is hosted under a subpath
covers:
  - web/src/base.ts
---
## Behaviour

The report can be hosted at a site root or under a project subpath (e.g. GitHub
project Pages at `/<repo>/`). On any route — including a hard refresh straight to
a spec-detail deep link like `/<repo>/spec/<id>` — a spec's screenshots resolve
against the app's base and load, rather than 404'ing.

## Why

The screenshot galleries are the visual proof a spec is really implemented. If
they break the moment the report is published under a repo subpath, that proof
is lost exactly where a reviewer looks for it.

---
id: report-escapes-breakout
title: Embedded data can't break out of the script
covers:
  - internal/report
---
## Behaviour

The embedded JSON is escaped so a spec body containing `</script>` or an HTML comment cannot break out of the `<script>` tag.

## Why

The report renders untrusted spec text; an injection that breaks the page (or worse) would undermine the trust artifact.

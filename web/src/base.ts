// Where the report is hosted varies: a site root (`specguard serve`, or a
// user/organisation GitHub Pages site), a project-Pages subpath like
// `/specguard/`, or a local file opened with `file://`. The single-file report
// is built once and served in all of these, so we resolve the base at runtime
// from the current URL rather than baking a base in at build time.

// basePath returns the directory the app is served from, always ending in `/`.
// A spec-detail deep link is `<base>spec/<id>`; when the current path ends in
// that route we strip it to recover the base, otherwise we drop any trailing
// file name (e.g. `index.html`). Examples:
//   `/specguard/`                    -> `/specguard/`
//   `/specguard/spec/ui-code-links`  -> `/specguard/`
//   `/`                              -> `/`
//   `/spec/auth-login`               -> `/`
//   `/home/u/report.html` (file://)  -> `/home/u/`
export function basePath(pathname: string = window.location.pathname): string {
  const route = pathname.match(/^(.*\/)spec\/[^/]+\/?$/);
  if (route) return route[1];
  return pathname.replace(/[^/]*$/, '');
}

// assetUrl resolves a report-relative asset path (e.g. `assets/specs/…/0.png`)
// against the base, so a screenshot loads from the app root no matter which
// route the browser is on. Without this a relative `assets/…` on a deep route
// like `/specguard/spec/<id>` would resolve to `/specguard/spec/assets/…` and
// 404. Already-absolute paths and full URLs are returned unchanged.
export function assetUrl(path: string, pathname?: string): string {
  if (path.startsWith('/') || /^[a-z][a-z0-9+.-]*:/i.test(path)) return path;
  return basePath(pathname) + path;
}

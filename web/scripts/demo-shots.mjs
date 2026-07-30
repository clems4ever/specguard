// Generate real screenshots for the published demo report. The example project
// (Taskflow) has no running app, so instead of a full e2e we render a few mock
// screens with Playwright, screenshot them, and emit a Playwright-shaped results
// file tagged `@spec:<id>` plus the PNGs. `specguard report -results ... -assets`
// then attaches them to the matching specs as per-spec galleries.
//
// Usage: node scripts/demo-shots.mjs <out-dir>
// Writes <out-dir>/<id>.png and <out-dir>/results.json (absolute paths inside).
import { chromium } from '@playwright/test';
import { mkdirSync, writeFileSync } from 'node:fs';
import path from 'node:path';

const outDir = path.resolve(process.argv[2] || 'demo-shots');
mkdirSync(outDir, { recursive: true });

const screen = (title, body) => `<!doctype html><html><body style="margin:0;font-family:system-ui,sans-serif;background:#0d1117;color:#e6edf3">
  <div style="padding:28px 32px;border-bottom:1px solid #30363d;font-weight:600">◉ Taskflow</div>
  <div style="padding:32px"><h1 style="margin:0 0 18px">${title}</h1>${body}</div>
</body></html>`;

const card = (t) => `<div style="border:1px solid #30363d;border-radius:10px;padding:16px 18px;margin:10px 0;background:#161b22">${t}</div>`;

// Each entry maps a mock screen to a real example spec id.
const shots = [
  { id: 'auth-login', title: 'Sign in', body:
    `<input placeholder="Email" style="display:block;width:320px;padding:10px;margin:8px 0;border-radius:8px;border:1px solid #30363d;background:#0d1117;color:#e6edf3">
     <input placeholder="Password" type="password" style="display:block;width:320px;padding:10px;margin:8px 0;border-radius:8px;border:1px solid #30363d;background:#0d1117;color:#e6edf3">
     <button style="padding:10px 18px;border-radius:8px;border:0;background:#238636;color:#fff;font-weight:600">Log in</button>` },
  { id: 'tasks-create', title: 'New task', body:
    `<input value="Ship the spec report" style="display:block;width:420px;padding:10px;margin:8px 0;border-radius:8px;border:1px solid #30363d;background:#0d1117;color:#e6edf3">
     <button style="padding:10px 18px;border-radius:8px;border:0;background:#238636;color:#fff;font-weight:600">Create</button>` },
  { id: 'tasks-complete', title: 'My tasks', body:
    card('☑ <s style="color:#8b949e">Draft the specs</s>') + card('☐ Wire up CI') + card('☐ Publish the report') },
  { id: 'sharing-invite', title: 'Share this board', body:
    `<input value="teammate@example.com" style="display:block;width:360px;padding:10px;margin:8px 0;border-radius:8px;border:1px solid #30363d;background:#0d1117;color:#e6edf3">
     <button style="padding:10px 18px;border-radius:8px;border:0;background:#1f6feb;color:#fff;font-weight:600">Invite</button>` },
];

const browser = await chromium.launch();
const page = await browser.newPage({ viewport: { width: 720, height: 420 } });
const specs = [];
for (const s of shots) {
  await page.setContent(screen(s.title, s.body));
  const file = path.join(outDir, `${s.id}.png`);
  await page.screenshot({ path: file });
  specs.push({
    title: s.title,
    tags: [`@spec:${s.id}`],
    tests: [{ results: [{ status: 'passed', attachments: [{ name: s.title, contentType: 'image/png', path: file }] }] }],
  });
}
await browser.close();

writeFileSync(path.join(outDir, 'results.json'), JSON.stringify({ suites: [{ specs }] }, null, 2));
console.log(`demo-shots: wrote ${shots.length} screenshots + results.json to ${outDir}`);

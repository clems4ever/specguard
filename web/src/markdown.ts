import { marked } from 'marked';

// Spec bodies are authored inside the repository (a trusted source), but we
// still strip <script>/<style> and inline event handlers as defense in depth
// before rendering the produced HTML.
export function renderMarkdown(md: string): string {
  const html = marked.parse(md ?? '', { async: false }) as string;
  return sanitize(html);
}

function sanitize(html: string): string {
  return html
    .replace(/<\/?(script|style)[^>]*>/gi, '')
    .replace(/\son\w+="[^"]*"/gi, '')
    .replace(/\son\w+='[^']*'/gi, '')
    .replace(/javascript:/gi, '');
}

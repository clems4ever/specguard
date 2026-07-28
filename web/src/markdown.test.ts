import { describe, it, expect } from 'vitest';
import { renderMarkdown } from './markdown';

describe('renderMarkdown', () => {
  it('renders headings as real elements', () => {
    expect(renderMarkdown('## Behaviour')).toContain('<h2');
    expect(renderMarkdown('## Behaviour')).toContain('Behaviour');
  });
  it('renders inline code', () => {
    expect(renderMarkdown('use `spec:id`')).toContain('<code>spec:id</code>');
  });
  it('strips <script> tags', () => {
    const html = renderMarkdown('hi\n\n<script>alert(1)</script>');
    expect(html).not.toContain('<script');
  });
  it('strips inline event handlers and javascript: urls', () => {
    const html = renderMarkdown('<a href="javascript:alert(1)" onclick="x()">x</a>');
    expect(html).not.toContain('onclick=');
    expect(html.toLowerCase()).not.toContain('javascript:');
  });
  it('tolerates empty input', () => {
    expect(renderMarkdown('')).toBe('');
  });
});

import { renderMarkdown } from '../markdown';

export function Markdown({ source }: { source: string }) {
  return (
    <div
      className="markdown"
      data-testid="markdown"
      dangerouslySetInnerHTML={{ __html: renderMarkdown(source) }}
    />
  );
}

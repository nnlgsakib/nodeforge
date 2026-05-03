import type { MonologueMessage } from '../hooks/useWebSocket';

function escapeMarkdown(text: string): string {
  return text.replace(/[\\*_#\[\]()]/g, '\\$&');
}

export function buildMarkdownContent(messages: MonologueMessage[], sessionId?: string): string {
  if (messages.length === 0) {
    return '# LLM Inner Monologue\n\n**Session:** ' + (sessionId || 'unknown') + '\n\nNo monologue messages recorded.\n';
  }

  const lines: string[] = [];
  lines.push('# LLM Inner Monologue');
  lines.push('');
  lines.push('**Session:** ' + (sessionId || 'unknown'));
  lines.push('**Generated on:** ' + new Date().toLocaleString());
  lines.push('');
  lines.push('---');
  lines.push('');

  for (const msg of messages) {
    const time = new Date(msg.timestamp).toLocaleString();
    lines.push(`## [${time}]`);
    lines.push('');
    lines.push(escapeMarkdown(msg.text));
    lines.push('');
  }

  return lines.join('\n');
}

export function exportMonologueAsMarkdown(messages: MonologueMessage[], sessionId?: string): void {
  const md = buildMarkdownContent(messages, sessionId);
  const blob = new Blob([md], { type: 'text/markdown' });
  const url = URL.createObjectURL(blob);
  const a = document.createElement('a');
  a.href = url;
  const dateStr = new Date().toISOString().slice(0, 10);
  a.download = `monologue-${dateStr}.md`;
  a.click();
  URL.revokeObjectURL(url);
}

import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { exportMonologueAsMarkdown, buildMarkdownContent } from './monologue-export';

describe('monologue-export', () => {
  beforeEach(() => {
    // Mock document.createElement
    const mockAnchor = {
      href: '',
      download: '',
      click: vi.fn(),
    };
    vi.spyOn(document, 'createElement').mockReturnValue(mockAnchor as any);
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  const sampleMessages = [
    { id: '1', text: 'Thinking about the goal...', timestamp: 1714500000000 },
    { id: '2', text: 'Analyzing node graph structure...', timestamp: 1714500001000 },
  ];

  describe('buildMarkdownContent', () => {
    it('should build valid markdown from monologue messages', () => {
      const md = buildMarkdownContent(sampleMessages, 'test-session-1');
      expect(md).toContain('# LLM Inner Monologue');
      expect(md).toContain('**Session:** test-session-1');
      expect(md).toContain('Thinking about the goal...');
      expect(md).toContain('Analyzing node graph structure...');
    });

    it('should return empty content marker when no messages', () => {
      const md = buildMarkdownContent([], 'test-session');
      expect(md).toContain('No monologue messages recorded');
    });
  });

  describe('exportMonologueAsMarkdown', () => {
    it('should use .md extension not .txt', () => {
      const anchor = { href: '', download: '', click: vi.fn() } as any;
      vi.spyOn(document, 'createElement').mockReturnValue(anchor);

      exportMonologueAsMarkdown(sampleMessages, 'session-123');

      expect(anchor.download).toMatch(/\.md$/);
      expect(anchor.download).not.toMatch(/\.txt$/);
      expect(anchor.click).toHaveBeenCalled();
    });
  });
});

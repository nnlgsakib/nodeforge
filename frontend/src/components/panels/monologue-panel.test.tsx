import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { MonologuePanel } from './monologue-panel';

const mockMessages = [
  { id: '1', text: 'Thinking...', timestamp: 1000 },
];

describe('MonologuePanel', () => {
  const defaultProps = {
    collapsed: true,
    onToggleCollapse: vi.fn(),
    messages: [] as any[],
    isStreaming: false,
    onClear: vi.fn(),
    sessionId: 'test-session',
  };

  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('should render toggle button when collapsed', () => {
    render(<MonologuePanel {...defaultProps} />);
    expect(screen.getByTitle('Toggle Monologue Panel (M)')).toBeTruthy();
  });

  it('should call onToggleCollapse when toggle button clicked', () => {
    render(<MonologuePanel {...defaultProps} />);
    const btn = screen.getByTitle('Toggle Monologue Panel (M)');
    fireEvent.click(btn);
    expect(defaultProps.onToggleCollapse).toHaveBeenCalled();
  });

  it('should show empty state when no messages and expanded', () => {
    render(<MonologuePanel {...defaultProps} collapsed={false} />);
    expect(screen.getByText(/LLM thoughts will appear/)).toBeTruthy();
  });

  it('should display messages when provided and expanded', () => {
    render(
      <MonologuePanel {...defaultProps} collapsed={false} messages={mockMessages} />
    );
    expect(screen.getByText('Thinking...')).toBeTruthy();
  });

  it('should show Export and Clear buttons when expanded', () => {
    render(<MonologuePanel {...defaultProps} collapsed={false} />);
    expect(screen.getByTitle('Export as Markdown')).toBeTruthy();
    expect(screen.getByTitle('Clear history')).toBeTruthy();
  });
});

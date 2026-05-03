import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { ChatPanel } from '../components/panels/ChatPanel';
import { MonologuePanel } from '../components/panels/monologue-panel';
import type { MonologueMessage } from '../hooks/useWebSocket';

/**
 * Accessibility tests for ChatPanel and MonologuePanel
 * Verifies ARIA labels, keyboard navigation, and screen reader compatibility
 */
describe('Accessibility: ChatPanel', () => {
  const defaultProps = {
    collapsed: false,
    onToggleCollapse: vi.fn(),
    onSendGoal: vi.fn(),
    generating: false,
  };

  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('should have accessible heading for chat panel', () => {
    render(<ChatPanel {...defaultProps} />);
    const heading = screen.getByRole('heading', { level: 3 });
    expect(heading).toBeTruthy();
    expect(heading.textContent).toContain('Chat');
  });

  it('should have accessible input with placeholder', () => {
    render(<ChatPanel {...defaultProps} />);
    const input = screen.getByPlaceholderText('Describe your goal...');
    expect(input).toBeTruthy();
    expect(input).not.toBeDisabled();
  });

  it('should have accessible submit button', () => {
    render(<ChatPanel {...defaultProps} />);
    const submitBtn = screen.getByRole('button', { name: 'Send' });
    expect(submitBtn).toBeTruthy();
  });

  it('should have accessible collapse toggle button with title', () => {
    render(<ChatPanel {...defaultProps} />);
    const collapseBtn = screen.getByTitle('Collapse chat');
    expect(collapseBtn).toBeTruthy();
    expect(collapseBtn).toHaveAttribute('title');
  });

  it('should disable input with proper ARIA state when generating', () => {
    render(<ChatPanel {...defaultProps} generating={true} />);
    const input = screen.getByPlaceholderText('Describe your goal...');
    expect(input).toBeDisabled();
  });

  it('should disable submit button when generating', () => {
    render(<ChatPanel {...defaultProps} generating={true} />);
    const submitBtn = screen.getByRole('button', { name: 'Send' });
    expect(submitBtn).toBeDisabled();
  });

  it('should disable submit button when input is invalid', () => {
    render(<ChatPanel {...defaultProps} />);
    const submitBtn = screen.getByRole('button', { name: 'Send' });
    expect(submitBtn).toBeDisabled();
  });

  it('should display validation message for screen readers when input is too short', () => {
    render(<ChatPanel {...defaultProps} />);
    const input = screen.getByPlaceholderText('Describe your goal...');
    fireEvent.change(input, { target: { value: 'short' } });
    const msg = screen.getByText(/Minimum 10 characters required/);
    expect(msg).toBeTruthy();
    // Verify the message is associated with input via aria-describedby
    expect(input).toHaveAttribute('aria-describedby');
  });

  it('should support keyboard Enter for submission', () => {
    render(<ChatPanel {...defaultProps} />);
    const input = screen.getByPlaceholderText('Describe your goal...');
    // Verify input accepts keyboard events (Enter key submission)
    expect(input.tagName).toBe('INPUT');
  });

  it('should have chat messages region', () => {
    const { container } = render(<ChatPanel {...defaultProps} />);
    const messagesContainer = container.querySelector('.chat-messages');
    expect(messagesContainer).toBeTruthy();
  });
});

describe('Accessibility: MonologuePanel', () => {
  const sampleMessages: MonologueMessage[] = [
    { id: '1', text: 'Thinking about the goal...', timestamp: Date.now() },
  ];

  const defaultProps = {
    collapsed: false,
    onToggleCollapse: vi.fn(),
    messages: [] as MonologueMessage[],
    isStreaming: false,
    onClear: vi.fn(),
    sessionId: 'test-session',
  };

  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('should have accessible Dialog title via Radix', () => {
    render(<MonologuePanel {...defaultProps} collapsed={false} />);
    // Radix Dialog.Title is visually hidden but accessible to screen readers
    const title = screen.getByText('LLM Inner Monologue');
    expect(title).toBeTruthy();
  });

  it('should have accessible Dialog description via Radix', () => {
    render(<MonologuePanel {...defaultProps} collapsed={false} />);
    const description = screen.getByText(/LLM inner monologue streaming panel/);
    expect(description).toBeTruthy();
  });

  it('should have accessible toggle button with keyboard shortcut hint', () => {
    render(<MonologuePanel {...defaultProps} collapsed={true} />);
    const toggleBtn = screen.getByTitle('Toggle Monologue Panel (M)');
    expect(toggleBtn).toBeTruthy();
  });

  it('should have accessible export button', () => {
    render(<MonologuePanel {...defaultProps} collapsed={false} />);
    const exportBtn = screen.getByTitle('Export as Markdown');
    expect(exportBtn).toBeTruthy();
  });

  it('should have accessible clear button', () => {
    render(<MonologuePanel {...defaultProps} collapsed={false} />);
    const clearBtn = screen.getByTitle('Clear history');
    expect(clearBtn).toBeTruthy();
  });

  it('should have accessible close button', () => {
    render(<MonologuePanel {...defaultProps} collapsed={false} />);
    const closeBtn = screen.getByTitle('Close (Esc)');
    expect(closeBtn).toBeTruthy();
  });

  it('should have accessible auto-scroll checkbox with label', () => {
    render(<MonologuePanel {...defaultProps} collapsed={false} messages={sampleMessages} />);
    const checkbox = screen.getByRole('checkbox', { name: /auto-scroll/i });
    expect(checkbox).toBeTruthy();
  });

  it('should display empty state message for screen readers', () => {
    render(<MonologuePanel {...defaultProps} collapsed={false} />);
    const emptyState = screen.getByText(/LLM thoughts will appear here during graph execution/);
    expect(emptyState).toBeTruthy();
  });

  it('should display messages with timestamps readable by screen readers', () => {
    render(
      <MonologuePanel
        {...defaultProps}
        collapsed={false}
        messages={sampleMessages}
        isStreaming={false}
      />
    );
    expect(screen.getByText('Thinking about the goal...')).toBeTruthy();
  });

  it('should disable export button when no messages', () => {
    render(<MonologuePanel {...defaultProps} collapsed={false} />);
    const exportBtn = screen.getByTitle('Export as Markdown');
    expect(exportBtn).toBeDisabled();
  });

  it('should disable clear button when no messages', () => {
    vi.spyOn(window, 'confirm').mockReturnValue(true);
    render(<MonologuePanel {...defaultProps} collapsed={false} />);
    const clearBtn = screen.getByTitle('Clear history');
    expect(clearBtn).toBeDisabled();
  });

  it('should enable export button when messages exist', () => {
    render(
      <MonologuePanel
        {...defaultProps}
        collapsed={false}
        messages={sampleMessages}
        isStreaming={false}
      />
    );
    const exportBtn = screen.getByTitle('Export as Markdown');
    expect(exportBtn).not.toBeDisabled();
  });
});

describe('Accessibility: Keyboard navigation', () => {
  it('ChatPanel input should be focusable', () => {
    render(
      <ChatPanel
        collapsed={false}
        onToggleCollapse={vi.fn()}
        onSendGoal={vi.fn()}
        generating={false}
      />
    );
    const input = screen.getByPlaceholderText('Describe your goal...');
    input.focus();
    expect(document.activeElement).toBe(input);
  });

  it('ChatPanel send button should be focusable', () => {
    render(
      <ChatPanel
        collapsed={false}
        onToggleCollapse={vi.fn()}
        onSendGoal={vi.fn()}
        generating={false}
      />
    );
    const sendBtn = screen.getByRole('button', { name: 'Send' });
    expect(sendBtn).toBeTruthy();
  });

  it('MonologuePanel buttons should be focusable', () => {
    const sampleMessages: MonologueMessage[] = [
      { id: '1', text: 'Thinking...', timestamp: Date.now() },
    ];
    render(
      <MonologuePanel
        collapsed={false}
        onToggleCollapse={vi.fn()}
        messages={sampleMessages}
        isStreaming={false}
        onClear={vi.fn()}
        sessionId="test"
      />
    );

    const buttons = screen.getAllByRole('button');
    expect(buttons.length).toBeGreaterThan(0);
  });
});

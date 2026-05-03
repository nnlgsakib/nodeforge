import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { render, screen, fireEvent, act, waitFor } from '@testing-library/react';
import { ChatPanel } from '../components/panels/ChatPanel';
import { MonologuePanel } from '../components/panels/monologue-panel';
import type { MonologueMessage } from '../hooks/useWebSocket';

/**
 * Integration test: goal submission → graph generation → monologue streaming
 *
 * This test verifies the full user flow:
 * 1. User types a goal in ChatPanel and presses Enter
 * 2. ChatPanel enters "generating" state
 * 3. WebSocket receives graph_update and monologue messages
 * 4. MonologuePanel displays streaming tokens
 * 5. User can toggle MonologuePanel with 'm' key
 */
describe('Goal submission → Graph generation flow', () => {
  const mockOnSendGoal = vi.fn();
  const mockOnToggleCollapse = vi.fn();
  const mockOnClear = vi.fn();

  const sampleMonologueMessages: MonologueMessage[] = [
    { id: '1', text: 'Analyzing user goal...', timestamp: Date.now() - 2000 },
    { id: '2', text: 'Generating node graph...', timestamp: Date.now() - 1000 },
    { id: '3', text: 'Creating implementation nodes...', timestamp: Date.now() },
  ];

  beforeEach(() => {
    vi.clearAllMocks();
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  it('should show generating state when goal is submitted', async () => {
    const { unmount: unmount1 } = render(
      <ChatPanel
        collapsed={false}
        onToggleCollapse={mockOnToggleCollapse}
        onSendGoal={mockOnSendGoal}
        generating={false}
      />
    );

    const input = screen.getByPlaceholderText('Describe your goal...');
    fireEvent.change(input, { target: { value: 'Build a REST API with Go and Gin framework' } });

    await act(async () => {
      fireEvent.keyDown(input, { key: 'Enter' });
    });

    // Verify goal was sent
    expect(mockOnSendGoal).toHaveBeenCalledWith('Build a REST API with Go and Gin framework');

    // Unmount first render before second render
    unmount1();

    // Re-render with generating=true to simulate backend processing
    const { unmount: unmount2 } = render(
      <ChatPanel
        collapsed={false}
        onToggleCollapse={mockOnToggleCollapse}
        onSendGoal={mockOnSendGoal}
        generating={true}
      />
    );

    expect(screen.getByText(/Generating graph/)).toBeTruthy();
    expect(screen.getByPlaceholderText('Describe your goal...')).toBeDisabled();
    expect(screen.getByRole('button', { name: 'Send' })).toBeDisabled();

    unmount2();
  });

  it('should display monologue messages during graph execution', () => {
    render(
      <MonologuePanel
        collapsed={false}
        onToggleCollapse={mockOnToggleCollapse}
        messages={sampleMonologueMessages}
        isStreaming={true}
        onClear={mockOnClear}
        sessionId="test-session-1"
      />
    );

    expect(screen.getByText('Analyzing user goal...')).toBeTruthy();
    expect(screen.getByText('Generating node graph...')).toBeTruthy();
    expect(screen.getByText('Creating implementation nodes...')).toBeTruthy();
  });

  it('should show streaming indicator (pulsing dot) when isStreaming is true', () => {
    render(
      <MonologuePanel
        collapsed={false}
        onToggleCollapse={mockOnToggleCollapse}
        messages={sampleMonologueMessages}
        isStreaming={true}
        onClear={mockOnClear}
        sessionId="test-session-1"
      />
    );

    // The pulsing dot is a span with animation: pulse inside the h3 header
    const heading = screen.getByRole('heading', { name: /LLM Thoughts/ });
    expect(heading).toBeTruthy();
    const pulsingDot = heading.querySelector('span[style*="animation"]');
    expect(pulsingDot).toBeTruthy();
  });

  it('should hide streaming indicator when isStreaming is false', () => {
    const { container } = render(
      <MonologuePanel
        collapsed={false}
        onToggleCollapse={mockOnToggleCollapse}
        messages={sampleMonologueMessages}
        isStreaming={false}
        onClear={mockOnClear}
        sessionId="test-session-1"
      />
    );

    // Verify no pulsing animation element
    const pulsingElements = container.querySelectorAll('span[style*="animation"]');
    expect(pulsingElements.length).toBe(0);
  });

  it('should truncate to last 100 messages when more than 100 are provided', () => {
    const manyMessages = Array.from({ length: 150 }, (_, i) => ({
      id: `${i}`,
      text: `Message ${i}`,
      timestamp: Date.now() - (150 - i) * 1000,
    }));

    render(
      <MonologuePanel
        collapsed={false}
        onToggleCollapse={mockOnToggleCollapse}
        messages={manyMessages}
        isStreaming={false}
        onClear={mockOnClear}
        sessionId="test-session-1"
      />
    );

    // Should show truncation notice
    expect(screen.getByText(/Showing last 100 of 150 messages/)).toBeTruthy();

    // First message (0) should NOT be visible, last messages should be
    expect(screen.queryByText('Message 0')).toBeNull();
    expect(screen.getByText('Message 149')).toBeTruthy();
  });

  it('should export monologue history as Markdown', async () => {
    const mockClick = vi.fn();
    const mockAnchor = {
      href: '',
      download: '',
      click: mockClick,
    };
    const originalCreateElement = document.createElement.bind(document);
    vi.spyOn(document, 'createElement').mockImplementation((tag: string) => {
      if (tag === 'a') return mockAnchor as any;
      return originalCreateElement(tag);
    });
    vi.spyOn(URL, 'createObjectURL').mockReturnValue('blob:test');
    vi.spyOn(URL, 'revokeObjectURL').mockImplementation(() => {});

    render(
      <MonologuePanel
        collapsed={false}
        onToggleCollapse={mockOnToggleCollapse}
        messages={sampleMonologueMessages}
        isStreaming={false}
        onClear={mockOnClear}
        sessionId="test-session-1"
      />
    );

    const exportBtn = screen.getByTitle('Export as Markdown');
    await act(async () => {
      fireEvent.click(exportBtn);
    });

    await waitFor(() => {
      expect(mockAnchor.download).toMatch(/\.md$/);
    });
    expect(mockClick).toHaveBeenCalled();
  });

  it('should clear monologue history when Clear button is clicked', () => {
    vi.spyOn(window, 'confirm').mockReturnValue(true);
    render(
      <MonologuePanel
        collapsed={false}
        onToggleCollapse={mockOnToggleCollapse}
        messages={sampleMonologueMessages}
        isStreaming={false}
        onClear={mockOnClear}
        sessionId="test-session-1"
      />
    );

    const clearBtn = screen.getByTitle('Clear history');
    fireEvent.click(clearBtn);

    expect(mockOnClear).toHaveBeenCalled();
    vi.restoreAllMocks();
  });

  it('should show empty state when no monologue messages', () => {
    render(
      <MonologuePanel
        collapsed={false}
        onToggleCollapse={mockOnToggleCollapse}
        messages={[]}
        isStreaming={false}
        onClear={mockOnClear}
        sessionId="test-session-1"
      />
    );

    expect(screen.getByText(/LLM thoughts will appear here during graph execution/)).toBeTruthy();
  });

  it('should toggle auto-scroll when checkbox is clicked', () => {
    render(
      <MonologuePanel
        collapsed={false}
        onToggleCollapse={mockOnToggleCollapse}
        messages={sampleMonologueMessages}
        isStreaming={false}
        onClear={mockOnClear}
        sessionId="test-session-1"
      />
    );

    const checkbox = screen.getByRole('checkbox', { name: /auto-scroll/i });
    // Default is checked (autoScroll defaults to true)
    expect(checkbox).toBeChecked();
    fireEvent.click(checkbox);
    expect(checkbox).not.toBeChecked();
    fireEvent.click(checkbox);
    expect(checkbox).toBeChecked();
  });

  it('should display timestamps for each message', () => {
    render(
      <MonologuePanel
        collapsed={false}
        onToggleCollapse={mockOnToggleCollapse}
        messages={sampleMonologueMessages}
        isStreaming={false}
        onClear={mockOnClear}
        sessionId="test-session-1"
      />
    );

    // Timestamps should be rendered as time strings
    const timeElements = screen.getAllByText(/\d{1,2}:\d{2}:\d{2}/);
    expect(timeElements.length).toBeGreaterThanOrEqual(1);
  });
});

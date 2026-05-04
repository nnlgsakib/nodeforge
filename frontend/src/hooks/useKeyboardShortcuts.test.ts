import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook } from '@testing-library/react';
import { useKeyboardShortcuts } from './useKeyboardShortcuts';

describe('useKeyboardShortcuts', () => {
  const mockHandlers = {
    onToggleMonologue: vi.fn(),
    onTogglePause: vi.fn(),
    onSkipNode: vi.fn(),
    onForkSession: vi.fn(),
    onRetryNode: vi.fn(),
    sendMessage: vi.fn(),
  };

  beforeEach(() => {
    vi.restoreAllMocks();
  });

  it('registers keydown listener on mount', () => {
    const addEventListenerSpy = vi.spyOn(window, 'addEventListener');
    renderHook(() =>
      useKeyboardShortcuts({
        ...mockHandlers,
        isPaused: false,
      })
    );
    expect(addEventListenerSpy).toHaveBeenCalledWith('keydown', expect.any(Function));
    addEventListenerSpy.mockRestore();
  });

  it('removes keydown listener on unmount', () => {
    const removeEventListenerSpy = vi.spyOn(window, 'removeEventListener');
    const { unmount } = renderHook(() =>
      useKeyboardShortcuts({
        ...mockHandlers,
        isPaused: false,
      })
    );
    unmount();
    expect(removeEventListenerSpy).toHaveBeenCalledWith('keydown', expect.any(Function));
    removeEventListenerSpy.mockRestore();
  });

  it('calls onToggleMonologue when m key pressed', () => {
    const handler = vi.fn();
    renderHook(() =>
      useKeyboardShortcuts({
        ...mockHandlers,
        onToggleMonologue: handler,
        isPaused: false,
      })
    );
    const event = new KeyboardEvent('keydown', { key: 'm' });
    window.dispatchEvent(event);
    expect(handler).toHaveBeenCalled();
  });

  it('calls onTogglePause when p key pressed', () => {
    renderHook(() =>
      useKeyboardShortcuts({
        ...mockHandlers,
        isPaused: false,
      })
    );
    const event = new KeyboardEvent('keydown', { key: 'p' });
    window.dispatchEvent(event);
    expect(mockHandlers.onTogglePause).toHaveBeenCalledWith(true);
    expect(mockHandlers.sendMessage).toHaveBeenCalledWith({ type: 'pause', paused: true });
  });

  it('calls onTogglePause when space key pressed', () => {
    renderHook(() =>
      useKeyboardShortcuts({
        ...mockHandlers,
        isPaused: false,
      })
    );
    const event = new KeyboardEvent('keydown', { key: ' ' });
    window.dispatchEvent(event);
    expect(mockHandlers.onTogglePause).toHaveBeenCalledWith(true);
    expect(mockHandlers.sendMessage).toHaveBeenCalledWith({ type: 'pause', paused: true });
  });

  it('calls onSkipNode when s key pressed', () => {
    renderHook(() =>
      useKeyboardShortcuts({
        ...mockHandlers,
        isPaused: false,
      })
    );
    const event = new KeyboardEvent('keydown', { key: 's' });
    window.dispatchEvent(event);
    expect(mockHandlers.onSkipNode).toHaveBeenCalled();
    expect(mockHandlers.sendMessage).toHaveBeenCalledWith({ type: 'skip_node' });
  });

  it('calls onForkSession when f key pressed', () => {
    renderHook(() =>
      useKeyboardShortcuts({
        ...mockHandlers,
        isPaused: false,
      })
    );
    const event = new KeyboardEvent('keydown', { key: 'f' });
    window.dispatchEvent(event);
    expect(mockHandlers.onForkSession).toHaveBeenCalled();
    expect(mockHandlers.sendMessage).toHaveBeenCalledWith({ type: 'fork_session' });
  });

  it('calls onRetryNode when r key pressed', () => {
    renderHook(() =>
      useKeyboardShortcuts({
        ...mockHandlers,
        isPaused: false,
      })
    );
    const event = new KeyboardEvent('keydown', { key: 'r' });
    window.dispatchEvent(event);
    expect(mockHandlers.onRetryNode).toHaveBeenCalled();
    expect(mockHandlers.sendMessage).toHaveBeenCalledWith({ type: 'retry_node' });
  });
});

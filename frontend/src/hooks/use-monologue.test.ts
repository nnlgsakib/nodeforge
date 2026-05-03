import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { renderHook, act } from '@testing-library/react';
import { useMonologue } from './use-monologue';

describe('useMonologue', () => {
  beforeEach(() => {
    vi.useFakeTimers();
  });

  afterEach(() => {
    vi.useRealTimers();
    vi.restoreAllMocks();
  });

  it('should initialize with empty messages by default', () => {
    const { result } = renderHook(() => useMonologue());
    expect(result.current.messages).toEqual([]);
    expect(result.current.isStreaming).toBe(false);
    expect(result.current.autoScroll).toBe(true);
  });

  it('should initialize with provided messages', () => {
    const initial = [
      { id: '1', text: 'hello', timestamp: 1000 },
    ];
    const { result } = renderHook(() => useMonologue(initial));
    expect(result.current.messages).toEqual(initial);
  });

  it('should add a message with addMessage', () => {
    const { result } = renderHook(() => useMonologue());
    act(() => {
      vi.setSystemTime(new Date(2024, 0, 1));
      result.current.addMessage('test thought');
    });
    expect(result.current.messages).toHaveLength(1);
    expect(result.current.messages[0].text).toBe('test thought');
    expect(result.current.messages[0].id).toMatch(/^msg-\d+$/);
  });

  it('should clear messages with clearMessages', () => {
    const initial = [
      { id: '1', text: 'hello', timestamp: 1000 },
    ];
    const { result } = renderHook(() => useMonologue(initial));
    act(() => {
      result.current.clearMessages();
    });
    expect(result.current.messages).toEqual([]);
  });

  it('should update autoScroll state', () => {
    const { result } = renderHook(() => useMonologue());
    act(() => {
      result.current.setAutoScroll(false);
    });
    expect(result.current.autoScroll).toBe(false);
  });

  it('should update isStreaming state', () => {
    const { result } = renderHook(() => useMonologue());
    act(() => {
      result.current.setStreaming(true);
    });
    expect(result.current.isStreaming).toBe(true);
  });
});

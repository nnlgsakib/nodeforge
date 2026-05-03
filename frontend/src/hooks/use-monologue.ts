import { useState, useCallback } from 'react';
import type { MonologueMessage } from './useWebSocket';

export interface UseMonologueReturn {
  messages: MonologueMessage[];
  isStreaming: boolean;
  autoScroll: boolean;
  addMessage: (text: string) => void;
  clearMessages: () => void;
  setAutoScroll: (value: boolean) => void;
  setStreaming: (value: boolean) => void;
}

export function useMonologue(initialMessages: MonologueMessage[] = []): UseMonologueReturn {
  const [messages, setMessages] = useState<MonologueMessage[]>(initialMessages);
  const [isStreaming, setStreaming] = useState(false);
  const [autoScroll, setAutoScroll] = useState(true);

  const addMessage = useCallback((text: string) => {
    setMessages((prev) => [
      ...prev,
      {
        id: `msg-${Date.now()}`,
        text,
        timestamp: Date.now(),
      },
    ]);
  }, []);

  const clearMessages = useCallback(() => {
    setMessages([]);
  }, []);

  return {
    messages,
    isStreaming,
    autoScroll,
    addMessage,
    clearMessages,
    setAutoScroll,
    setStreaming,
  };
}

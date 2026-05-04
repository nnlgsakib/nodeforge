import { useEffect } from 'react';
import type { WebSocketMessage } from './useWebSocket';

interface KeyboardShortcutsOptions {
  onToggleMonologue: () => void;
  onTogglePause: (paused: boolean) => void;
  isPaused: boolean;
  onSkipNode: () => void;
  onForkSession: () => void;
  onRetryNode: () => void;
  sendMessage: (msg: WebSocketMessage) => void;
  connected?: boolean;
}

export function useKeyboardShortcuts({
  onToggleMonologue,
  onTogglePause,
  isPaused,
  onSkipNode,
  onForkSession,
  onRetryNode,
  sendMessage,
  connected,
}: KeyboardShortcutsOptions) {
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.repeat) return;
      const target = e.target as HTMLElement;
      if (target.tagName === 'INPUT' || target.tagName === 'TEXTAREA') return;

      switch (e.key) {
        case 'm':
          if (!e.ctrlKey && !e.metaKey) {
            e.preventDefault();
            onToggleMonologue();
          }
          break;
        case 'p':
        case ' ':
          if (!e.ctrlKey && !e.metaKey) {
            e.preventDefault();
            const newPaused = !isPaused;
            onTogglePause(newPaused);
            sendMessage({ type: 'pause', paused: newPaused });
          }
          break;
        case 's':
          if (!e.ctrlKey && !e.metaKey) {
            e.preventDefault();
            onSkipNode();
            if (connected) sendMessage({ type: 'skip_node' });
          }
          break;
        case 'f':
          if (!e.ctrlKey && !e.metaKey) {
            e.preventDefault();
            onForkSession();
            if (connected) sendMessage({ type: 'fork_session' });
          }
          break;
        case 'r':
          if (!e.ctrlKey && !e.metaKey) {
            e.preventDefault();
            onRetryNode();
            if (connected) sendMessage({ type: 'retry_node' });
          }
          break;
      }
    };

    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  }, [onToggleMonologue, onTogglePause, isPaused, onSkipNode, onForkSession, onRetryNode, sendMessage]);
}

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
}

export function useKeyboardShortcuts({
  onToggleMonologue,
  onTogglePause,
  isPaused,
  onSkipNode,
  onForkSession,
  onRetryNode,
  sendMessage,
}: KeyboardShortcutsOptions) {
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
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
            sendMessage({ type: 'skip_node' });
          }
          break;
        case 'f':
          if (!e.ctrlKey && !e.metaKey) {
            e.preventDefault();
            onForkSession();
            sendMessage({ type: 'fork_session' });
          }
          break;
        case 'r':
          if (!e.ctrlKey && !e.metaKey) {
            e.preventDefault();
            onRetryNode();
            sendMessage({ type: 'retry_node' });
          }
          break;
      }
    };

    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  }, [onToggleMonologue, onTogglePause, isPaused, onSkipNode, onForkSession, onRetryNode, sendMessage]);
}

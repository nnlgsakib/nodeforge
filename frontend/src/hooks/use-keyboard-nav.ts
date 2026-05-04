import { useEffect, useCallback, useRef } from 'react';

interface UseKeyboardNavOptions {
  /** Called when vim/emacs navigation should pan the canvas */
  onPan?: (dx: number, dy: number) => void;
  /** Called for one-key node controls: p=pause, r=retry, f=fork, s=skip, m=monologue */
  onKeyAction?: (action: string) => void;
  /** Array of node IDs for Tab cycling */
  nodeIds?: string[];
  /** Called when a node should be selected via keyboard */
  onSelectNode?: (nodeId: string) => void;
  /** ID of currently selected node */
  selectedNodeId?: string | null;
}

/**
 * Hook for Vim/Emacs keybindings and one-key controls (AC:4, Subtasks 4.1-4.7)
 * Handles: h/j/k/l pan, Ctrl-f/b/n/p pan, p/r/f/s/m one-key controls, Tab node cycling, Escape close
 */
export function useKeyboardNav({
  onPan,
  onKeyAction,
  nodeIds,
  onSelectNode,
  selectedNodeId,
}: UseKeyboardNavOptions) {
  const currentIndexRef = useRef(-1);

  const handleKeyDown = useCallback(
    (e: KeyboardEvent) => {
      const target = e.target as HTMLElement;
      if (
        target.tagName === 'INPUT' ||
        target.tagName === 'TEXTAREA' ||
        target.isContentEditable ||
        target.closest('[data-ignore-keyboard]')
      ) {
        return;
      }

      // Escape key to close panels (Subtask 4.5)
      if (e.key === 'Escape') {
        onKeyAction?.('escape');
        return;
      }

      // Vim keys: only on bare keys (no modifiers)
      if (!e.ctrlKey && !e.metaKey && !e.altKey && !e.shiftKey) {
        const panStep = 50;
        switch (e.key) {
          case 'h':
            e.preventDefault();
            onPan?.(-panStep, 0);
            return;
          case 'j':
            e.preventDefault();
            onPan?.(0, panStep);
            return;
          case 'k':
            e.preventDefault();
            onPan?.(0, -panStep);
            return;
          case 'l':
            e.preventDefault();
            onPan?.(panStep, 0);
            return;
        }

        // One-key node controls (Subtask 4.3)
        switch (e.key) {
          case 'p':
            e.preventDefault();
            onKeyAction?.('pause');
            return;
          case 'r':
            e.preventDefault();
            onKeyAction?.('retry');
            return;
          case 'f':
            e.preventDefault();
            onKeyAction?.('fork');
            return;
          case 's':
            e.preventDefault();
            onKeyAction?.('skip');
            return;
          case 'm':
            e.preventDefault();
            onKeyAction?.('monologue');
            return;
          case ' ':
            e.preventDefault();
            onKeyAction?.('pause');
            return;
        }

        return;
      }

      // Emacs keys: Ctrl-f/b/n/p (Subtask 4.2)
      if (e.ctrlKey) {
        const panStep = 50;
        switch (e.key) {
          case 'f':
            e.preventDefault();
            onPan?.(panStep, 0);
            return;
          case 'b':
            e.preventDefault();
            onPan?.(-panStep, 0);
            return;
          case 'n':
            e.preventDefault();
            onPan?.(0, panStep);
            return;
          case 'p':
            e.preventDefault();
            onPan?.(0, -panStep);
            return;
        }
      }

      // Tab cycles through nodes (Subtask 4.4)
      if (e.key === 'Tab' && nodeIds && nodeIds.length > 0) {
        e.preventDefault();
        if (selectedNodeId) {
          const currentIdx = nodeIds.indexOf(selectedNodeId);
          const nextIdx = e.shiftKey
            ? (currentIdx - 1 + nodeIds.length) % nodeIds.length
            : (currentIdx + 1) % nodeIds.length;
          onSelectNode?.(nodeIds[nextIdx]);
        } else {
          onSelectNode?.(nodeIds[0]);
        }
        currentIndexRef.current = nodeIds.indexOf(selectedNodeId || nodeIds[0]);
        return;
      }

      // Enter activates selected node
      if (e.key === 'Enter' && selectedNodeId) {
        e.preventDefault();
        onKeyAction?.('activate');
        return;
      }
    },
    [onPan, onKeyAction, nodeIds, onSelectNode, selectedNodeId]
  );

  useEffect(() => {
    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  }, [handleKeyDown]);
}

import { useCallback, useRef, useEffect } from 'react';

const LAYOUT_TIMEOUT_MS = 5000;

interface LayoutNode {
  id: string;
  width?: number;
  height?: number;
}

interface LayoutEdge {
  source: string;
  target: string;
}

interface LayoutConfig {
  rankdir?: 'TB' | 'BT' | 'LR' | 'RL';
  nodeWidth?: number;
  nodeHeight?: number;
  ranksep?: number;
  nodesep?: number;
}

interface LayoutMetrics {
  layoutTimeMS: number;
  nodeCount: number;
  edgeCount: number;
}

interface PendingRequest {
  resolve: (v: { positions: Record<string, { x: number; y: number }>; metrics?: LayoutMetrics }) => void;
  reject: (e: Error) => void;
  timer: ReturnType<typeof setTimeout>;
}

interface UseLayoutWorkerReturn {
  runLayout: (nodes: LayoutNode[], edges: LayoutEdge[], config?: LayoutConfig) => Promise<{ positions: Record<string, { x: number; y: number }>; metrics?: LayoutMetrics }>;
}

export function useLayoutWorker(): UseLayoutWorkerReturn {
  const workerRef = useRef<Worker | null>(null);
  const pendingRef = useRef<PendingRequest | null>(null);

  const getWorker = useCallback(() => {
    if (!workerRef.current) {
      const worker = new Worker(
        new URL('../workers/layout.worker.ts', import.meta.url),
        { type: 'module' }
      );

      // L6: Single persistent message listener
      worker.addEventListener('message', (event: MessageEvent) => {
        const pending = pendingRef.current;
        if (!pending) return;

        if (event.data.type === 'layout-done') {
          clearTimeout(pending.timer);
          pendingRef.current = null;
          pending.resolve({ positions: event.data.positions, metrics: event.data.metrics });
        } else if (event.data.type === 'layout-error') {
          clearTimeout(pending.timer);
          pendingRef.current = null;
          pending.reject(new Error(event.data.error));
        }
        // Other message types are ignored silently
      });

      // Handle worker runtime errors
      worker.addEventListener('error', () => {
        const pending = pendingRef.current;
        if (pending) {
          clearTimeout(pending.timer);
          pendingRef.current = null;
          pending.reject(new Error('Web Worker encountered a runtime error'));
        }
      });

      workerRef.current = worker;
    }
    return workerRef.current;
  }, []);

  // Cleanup worker on unmount
  useEffect(() => {
    return () => {
      if (pendingRef.current) {
        clearTimeout(pendingRef.current.timer);
        pendingRef.current.reject(new Error('Component unmounted'));
        pendingRef.current = null;
      }
      if (workerRef.current) {
        workerRef.current.terminate();
        workerRef.current = null;
      }
    };
  }, []);

  const runLayout = useCallback(
    (nodes: LayoutNode[], edges: LayoutEdge[], config?: LayoutConfig) => {
      return new Promise<{ positions: Record<string, { x: number; y: number }>; metrics?: LayoutMetrics }>((resolve, reject) => {
        const worker = getWorker();

        // Cancel any pending request
        if (pendingRef.current) {
          clearTimeout(pendingRef.current.timer);
          pendingRef.current.reject(new Error('Cancelled by newer layout request'));
        }

        // L1: Timeout guard — reject if worker doesn't respond within 5s
        const timer = setTimeout(() => {
          pendingRef.current = null;
          reject(new Error('Layout worker timed out'));
        }, LAYOUT_TIMEOUT_MS);

        pendingRef.current = { resolve, reject, timer };
        worker.postMessage({ type: 'layout', nodes, edges, config });
      });
    },
    [getWorker]
  );

  return { runLayout };
}

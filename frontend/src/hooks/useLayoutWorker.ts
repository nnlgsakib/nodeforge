import { useCallback, useRef } from 'react';

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

interface UseLayoutWorkerReturn {
  runLayout: (nodes: LayoutNode[], edges: LayoutEdge[], config?: LayoutConfig) => Promise<Record<string, { x: number; y: number }>>;
}

export function useLayoutWorker(): UseLayoutWorkerReturn {
  const workerRef = useRef<Worker | null>(null);
  const requestIdRef = useRef(0);
  const pendingRef = useRef<{ id: number; resolve: (v: any) => void; reject: (e: Error) => void } | null>(null);

  const getWorker = useCallback(() => {
    if (!workerRef.current) {
      workerRef.current = new Worker(
        new URL('../workers/layout.worker.ts', import.meta.url),
        { type: 'module' }
      );
    }
    return workerRef.current;
  }, []);

  const runLayout = useCallback(
    (nodes: LayoutNode[], edges: LayoutEdge[], config?: LayoutConfig) => {
      return new Promise<Record<string, { x: number; y: number }>>((resolve, reject) => {
        const worker = getWorker();

        // Cancel any pending request
        if (pendingRef.current) {
          pendingRef.current.reject(new Error('Cancelled by newer layout request'));
          pendingRef.current = null;
        }

        const requestId = ++requestIdRef.current;
        pendingRef.current = { id: requestId, resolve, reject };

        const handler = (event: MessageEvent) => {
          if (pendingRef.current?.id !== requestId) {
            // Stale response, ignore
            return;
          }
          worker.removeEventListener('message', handler);
          pendingRef.current = null;

          if (event.data.type === 'layout-done') {
            resolve(event.data.positions);
          } else if (event.data.type === 'layout-error') {
            reject(new Error(event.data.error));
          } else {
            reject(new Error('Unexpected worker message: ' + event.data.type));
          }
        };

        worker.addEventListener('message', handler);
        worker.postMessage({ type: 'layout', nodes, edges, config });
      });
    },
    [getWorker]
  );

  return { runLayout };
}

import { useEffect, useRef, useState, useCallback } from 'react';

export interface WebSocketMessage {
  type: string;
  [key: string]: unknown;
}

export interface MonologueMessage {
  id: string;
  text: string;
  timestamp: number;
}

export interface UseWebSocketReturn {
  connected: boolean;
  monologueMessages: MonologueMessage[];
  isStreaming: boolean;
  sendMessage: (msg: WebSocketMessage) => void;
  lastGraphUpdate: unknown | null;
  lastNodeUpdate: unknown | null;
  lastEdgeUpdate: unknown | null;
}

export function useWebSocket(): UseWebSocketReturn {
  const wsRef = useRef<WebSocket | null>(null);
  const [connected, setConnected] = useState(false);
  const [monologueMessages, setMonologueMessages] = useState<MonologueMessage[]>([]);
  const [isStreaming, setIsStreaming] = useState(false);
  const [lastGraphUpdate, setLastGraphUpdate] = useState<unknown | null>(null);
  const [lastNodeUpdate, setLastNodeUpdate] = useState<unknown | null>(null);
  const [lastEdgeUpdate, setLastEdgeUpdate] = useState<unknown | null>(null);

  const sendMessage = useCallback((msg: WebSocketMessage) => {
    if (wsRef.current && wsRef.current.readyState === WebSocket.OPEN) {
      wsRef.current.send(JSON.stringify(msg));
    }
  }, []);

  useEffect(() => {
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const wsUrl = `${protocol}//${window.location.host}/ws`;
    const ws = new WebSocket(wsUrl);
    wsRef.current = ws;

    ws.onopen = () => {
      setConnected(true);
    };

    ws.onmessage = (event: MessageEvent) => {
      try {
        const data = JSON.parse(event.data as string);
        switch (data.type) {
          case 'graph_update':
            setLastGraphUpdate(data);
            break;
          case 'node_update':
            setLastNodeUpdate(data);
            break;
          case 'edge_update':
            setLastEdgeUpdate(data);
            break;
          case 'llm_chunk':
          case 'monologue':
            setIsStreaming(true);
            setMonologueMessages((prev) => [
              ...prev,
              {
                id: `msg-${Date.now()}`,
                text: data.text || data.token || '',
                timestamp: Date.now(),
              },
            ]);
            // Auto-stop streaming after a brief delay of no messages
            setTimeout(() => setIsStreaming(false), 2000);
            break;
          case 'connected':
            // Connection confirmed
            break;
        }
      } catch {
        // Ignore parse errors
      }
    };

    ws.onclose = () => {
      setConnected(false);
      wsRef.current = null;
    };

    ws.onerror = () => {
      setConnected(false);
    };

    return () => {
      ws.close();
    };
  }, []);

  return {
    connected,
    monologueMessages,
    isStreaming,
    sendMessage,
    lastGraphUpdate,
    lastNodeUpdate,
    lastEdgeUpdate,
  };
}

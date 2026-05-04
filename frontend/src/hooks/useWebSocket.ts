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

export interface SkillInstallMessage {
  skillId: string;
  status: 'installed' | 'failed';
  message?: string;
}

export interface UseWebSocketReturn {
  connected: boolean;
  monologueMessages: MonologueMessage[];
  isStreaming: boolean;
  sendMessage: (msg: WebSocketMessage) => void;
  graphUpdateQueue: unknown[];
  nodeUpdateQueue: unknown[];
  edgeUpdateQueue: unknown[];
  skillInstallMessages: SkillInstallMessage[];
  clearGraphUpdates: () => void;
  clearNodeUpdates: () => void;
  clearEdgeUpdates: () => void;
  clearMonologueMessages: () => void;
  clearSkillInstallMessages: () => void;
  reconnect: () => void;
  sessionResumed: boolean;
}

export interface SessionResumeMessage {
  type: 'session_resume';
  sessionId: string;
  graphJson: string;
  chatLog: string;
}

export function useWebSocket(): UseWebSocketReturn {
  const wsRef = useRef<WebSocket | null>(null);
  const [connected, setConnected] = useState(false);
  const [monologueMessages, setMonologueMessages] = useState<MonologueMessage[]>([]);
  const streamTimeoutRef = useRef<ReturnType<typeof setTimeout> | null>(null);
  const [isStreaming, setIsStreaming] = useState(false);
  const [graphUpdateQueue, setGraphUpdateQueue] = useState<unknown[]>([]);
  const [nodeUpdateQueue, setNodeUpdateQueue] = useState<unknown[]>([]);
  const [edgeUpdateQueue, setEdgeUpdateQueue] = useState<unknown[]>([]);
  const [skillInstallMessages, setSkillInstallMessages] = useState<SkillInstallMessage[]>([]);
  const [sessionResumed, setSessionResumed] = useState(false);
  const reconnectRef = useRef<(() => void) | null>(null);
  const resumeTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null);

  const clearGraphUpdates = useCallback(() => setGraphUpdateQueue([]), []);
  const clearNodeUpdates = useCallback(() => setNodeUpdateQueue([]), []);
  const clearEdgeUpdates = useCallback(() => setEdgeUpdateQueue([]), []);
  const clearMonologueMessages = useCallback(() => setMonologueMessages([]), []);
  const clearSkillInstallMessages = useCallback(() => setSkillInstallMessages([]), []);

  const sendMessage = useCallback((msg: WebSocketMessage) => {
    if (wsRef.current && wsRef.current.readyState === WebSocket.OPEN) {
      wsRef.current.send(JSON.stringify(msg));
    } else {
      console.warn('WebSocket not connected, message not sent:', msg);
    }
  }, []);

  const reconnect = useCallback(() => {
    if (wsRef.current) {
      wsRef.current.close();
      wsRef.current = null;
    }
    setSessionResumed(false);
    if (resumeTimerRef.current) {
      clearTimeout(resumeTimerRef.current);
      resumeTimerRef.current = null;
    }
    // Re-trigger the connection by creating a new WebSocket
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
            setGraphUpdateQueue(prev => [...prev, data]);
            break;
          case 'node_update':
            setNodeUpdateQueue(prev => [...prev, data]);
            break;
          case 'edge_update':
            setEdgeUpdateQueue(prev => [...prev, data]);
            break;
          case 'llm_chunk':
          case 'monologue':
            setIsStreaming(true);
            setMonologueMessages((prev) => [
              ...prev.slice(-499),
              {
                id: `msg-${Date.now()}-${Math.random().toString(36).slice(2, 8)}`,
                text: data.text || data.token || '',
                timestamp: Date.now(),
              },
            ].slice(-500));
            if (streamTimeoutRef.current) clearTimeout(streamTimeoutRef.current);
            streamTimeoutRef.current = setTimeout(() => setIsStreaming(false), 2000);
            break;
          case 'connected':
            break;
          case 'session_resume':
            if (resumeTimerRef.current) clearTimeout(resumeTimerRef.current);
            setSessionResumed(true);
            resumeTimerRef.current = setTimeout(() => setSessionResumed(false), 500);
            break;
          case 'skill_installed':
            setSkillInstallMessages((prev) => [
              ...prev,
              { skillId: data.skillId, status: 'installed' },
            ]);
            break;
          case 'skill_install_failed':
            setSkillInstallMessages((prev) => [
              ...prev,
              { skillId: data.skillId, status: 'failed', message: data.message },
            ]);
            break;
        }
      } catch {
        // Ignore parse errors
      }
    };

    ws.onclose = () => {
      setConnected(false);
      setIsStreaming(false);
      wsRef.current = null;
      if (streamTimeoutRef.current) {
        clearTimeout(streamTimeoutRef.current);
        streamTimeoutRef.current = null;
      }
    };

    ws.onerror = () => {
      setConnected(false);
    };
  }, []);

  reconnectRef.current = reconnect;

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
            setGraphUpdateQueue(prev => [...prev, data]);
            break;
          case 'node_update':
            setNodeUpdateQueue(prev => [...prev, data]);
            break;
          case 'edge_update':
            setEdgeUpdateQueue(prev => [...prev, data]);
            break;
          case 'llm_chunk':
          case 'monologue':
            setIsStreaming(true);
            setMonologueMessages((prev) => [
              ...prev.slice(-499),
              {
                id: `msg-${Date.now()}-${Math.random().toString(36).slice(2, 8)}`,
                text: data.text || data.token || '',
                timestamp: Date.now(),
              },
            ].slice(-500));
            // Reset timeout on new messages
            if (streamTimeoutRef.current) clearTimeout(streamTimeoutRef.current);
            streamTimeoutRef.current = setTimeout(() => setIsStreaming(false), 2000);
            break;
          case 'connected':
            // Connection confirmed
            break;
          case 'session_resume':
            if (resumeTimerRef.current) clearTimeout(resumeTimerRef.current);
            setSessionResumed(true);
            resumeTimerRef.current = setTimeout(() => setSessionResumed(false), 500);
            break;
          case 'skill_installed':
            setSkillInstallMessages((prev) => [
              ...prev,
              { skillId: data.skillId, status: 'installed' },
            ]);
            break;
          case 'skill_install_failed':
            setSkillInstallMessages((prev) => [
              ...prev,
              { skillId: data.skillId, status: 'failed', message: data.message },
            ]);
            break;
        }
      } catch {
        // Ignore parse errors
      }
    };

    ws.onclose = () => {
      setConnected(false);
      setIsStreaming(false);
      wsRef.current = null;
      // Clear stream timeout on disconnect
      if (streamTimeoutRef.current) {
        clearTimeout(streamTimeoutRef.current);
        streamTimeoutRef.current = null;
      }
    };

    ws.onerror = () => {
      setConnected(false);
    };

    return () => {
      // Clear stream timeout on unmount
      if (streamTimeoutRef.current) {
        clearTimeout(streamTimeoutRef.current);
      }
      if (resumeTimerRef.current) {
        clearTimeout(resumeTimerRef.current);
      }
      ws.close();
    };
  }, []);

  return {
    connected,
    monologueMessages,
    isStreaming,
    sendMessage,
    graphUpdateQueue,
    nodeUpdateQueue,
    edgeUpdateQueue,
    skillInstallMessages,
    clearGraphUpdates,
    clearNodeUpdates,
    clearEdgeUpdates,
    clearMonologueMessages,
    clearSkillInstallMessages,
    reconnect,
    sessionResumed,
  };
}

import { useState, useCallback, useEffect, useRef } from 'react';

export interface Session {
  sessionId: string;
  projectName: string;
  status: 'running' | 'complete' | 'failed' | 'paused' | 'zombie';
  goal: string;
  workspace: string;
  createdAt: string;
  lastActive: string;
}

// L3: Basic runtime validation for session objects from API
function isValidSession(obj: unknown): obj is Session {
  if (typeof obj !== 'object' || obj === null) return false;
  const s = obj as Record<string, unknown>;
  return (
    typeof s.sessionId === 'string' &&
    typeof s.projectName === 'string' &&
    typeof s.status === 'string' &&
    typeof s.createdAt === 'string' &&
    typeof s.lastActive === 'string'
  );
}

interface UseSessionReturn {
  sessions: Session[];
  currentSession: Session | null;
  loading: boolean;
  error: string | null;
  createSession: (projectName: string, goal?: string) => Promise<Session | null>;
  listSessions: () => Promise<void>;
  getSession: (sessionId: string) => Promise<Session | null>;
  autoSaveSession: (sessionId: string, data: { graphJson?: string; chatLog?: string; status?: string }) => Promise<Session | null>;
  resumeSession: (sessionId: string) => Promise<Session | null>;
}

export function useSession(): UseSessionReturn {
  const [sessions, setSessions] = useState<Session[]>([]);
  const [currentSession, setCurrentSession] = useState<Session | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const abortRef = useRef<AbortController | null>(null);

  const cleanup = useCallback(() => {
    if (abortRef.current) {
      abortRef.current.abort();
      abortRef.current = null;
    }
  }, []);

  useEffect(() => {
    return cleanup;
  }, [cleanup]);

  const listSessions = useCallback(async () => {
    cleanup();
    const controller = new AbortController();
    abortRef.current = controller;

    setLoading(true);
    setError(null);
    try {
      const response = await fetch('/api/v1/sessions', { signal: controller.signal });
      if (!response.ok) {
        throw new Error(`Failed to list sessions: ${response.statusText}`);
      }
      const data = await response.json();
      const items = Array.isArray(data.data) ? data.data : [];
      // L3: Filter to only valid sessions
      setSessions(items.filter(isValidSession));
    } catch (err: unknown) {
      if (err instanceof Error && err.name !== 'AbortError') {
        setError(err.message);
        setSessions([]);
      }
    } finally {
      setLoading(false);
    }
  }, [cleanup]);

  const createSession = useCallback(async (projectName: string, goal?: string) => {
    if (!projectName || !projectName.trim()) {
      setError('Project name is required');
      return null;
    }
    setLoading(true);
    setError(null);
    try {
      const response = await fetch('/api/v1/sessions', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ projectName: projectName.trim(), goal }),
      });
      if (!response.ok) {
        const err = await response.json();
        throw new Error(err.error || 'Failed to create session');
      }
      const session: unknown = await response.json();
      if (!isValidSession(session)) {
        throw new Error('Invalid session response from server');
      }
      setSessions(prev => [session, ...prev]);
      setCurrentSession(session);
      return session;
    } catch (err: unknown) {
      if (err instanceof Error) {
        setError(err.message);
      }
      return null;
    } finally {
      setLoading(false);
    }
  }, []);

  const getSession = useCallback(async (sessionId: string) => {
    cleanup();
    const controller = new AbortController();
    abortRef.current = controller;

    setLoading(true);
    setError(null);
    try {
      const response = await fetch(`/api/v1/sessions/${sessionId}`, { signal: controller.signal });
      if (!response.ok) {
        const err = await response.json();
        throw new Error(err.error || 'Failed to get session');
      }
      const session: unknown = await response.json();
      if (!isValidSession(session)) {
        throw new Error('Invalid session response from server');
      }
      setCurrentSession(session);
      return session;
    } catch (err: unknown) {
      if (err instanceof Error && err.name !== 'AbortError') {
        setError(err.message);
      }
      return null;
    } finally {
      setLoading(false);
    }
  }, [cleanup]);

  const autoSaveSession = useCallback(async (
    sessionId: string,
    data: { graphJson?: string; chatLog?: string; status?: string }
  ) => {
    cleanup();
    const controller = new AbortController();
    abortRef.current = controller;

    try {
      const response = await fetch(`/api/v1/sessions/${sessionId}/auto-save`, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(data),
        signal: controller.signal,
      });
      if (!response.ok) {
        const err = await response.json();
        throw new Error(err.error || 'Failed to auto-save session');
      }
      const session: unknown = await response.json();
      if (!isValidSession(session)) {
        throw new Error('Invalid session response from server');
      }
      // Update in list
      setSessions(prev => prev.map(s => s.sessionId === sessionId ? session : s));
      if (currentSession?.sessionId === sessionId) {
        setCurrentSession(session);
      }
      return session;
    } catch (err: unknown) {
      if (err instanceof Error && err.name !== 'AbortError') {
        setError(err.message);
      }
      return null;
    }
  }, [cleanup, currentSession]);

  const resumeSession = useCallback(async (sessionId: string) => {
    cleanup();
    const controller = new AbortController();
    abortRef.current = controller;

    setLoading(true);
    setError(null);
    try {
      const response = await fetch(`/api/v1/sessions/${sessionId}/resume`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        signal: controller.signal,
      });
      if (!response.ok) {
        const err = await response.json();
        throw new Error(err.error || 'Failed to resume session');
      }
      const session: unknown = await response.json();
      if (!isValidSession(session)) {
        throw new Error('Invalid session response from server');
      }
      // Update in list
      setSessions(prev => prev.map(s => s.sessionId === sessionId ? session : s));
      setCurrentSession(session);
      return session;
    } catch (err: unknown) {
      if (err instanceof Error && err.name !== 'AbortError') {
        setError(err.message);
      }
      return null;
    } finally {
      setLoading(false);
    }
  }, [cleanup, currentSession]);

  return {
    sessions,
    currentSession,
    loading,
    error,
    createSession,
    listSessions,
    getSession,
    autoSaveSession,
    resumeSession,
  };
}

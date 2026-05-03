import React, { useState, useCallback, useEffect, useMemo } from 'react';

interface Session {
  sessionId: string;
  projectName: string;
  status: 'running' | 'complete' | 'failed' | 'paused';
  createdAt: string;
  lastActive: string;
}

interface SessionExplorerProps {
  onCreateProject: (projectName: string) => void;
  onResumeSession?: (sessionId: string) => void;
  onForkSession?: (sessionId: string) => void;
  onExportSession?: (sessionId: string) => void;
  onSelectSession?: (sessionId: string) => void;
}

type StatusFilter = 'all' | 'running' | 'complete' | 'failed' | 'paused';
type DateFilter = 'all' | 'today' | 'week' | 'month';

const statusColors: Record<Session['status'], string> = {
  running: '#06b6d4',
  complete: '#22c55e',
  failed: '#ef4444',
  paused: '#ff9800',
};

export const SessionExplorer: React.FC<SessionExplorerProps> = ({
  onCreateProject,
  onResumeSession,
  onForkSession,
  onExportSession,
  onSelectSession,
}) => {
  const [projectName, setProjectName] = useState('');
  const [isCreating, setIsCreating] = useState(false);
  const [sessions, setSessions] = useState<Session[]>([]);
  const [loading, setLoading] = useState(false);
  const [searchQuery, setSearchQuery] = useState('');
  const [statusFilter, setStatusFilter] = useState<StatusFilter>('all');
  const [dateFilter, setDateFilter] = useState<DateFilter>('all');

  // Fetch sessions from API
  const fetchSessions = useCallback(async () => {
    setLoading(true);
    try {
      const response = await fetch('/api/v1/sessions');
      if (!response.ok) {
        // API not available yet (Epic 4 not implemented) - use mock data
        console.warn('Sessions API not available, using mock data');
        setSessions(getMockSessions());
        return;
      }
      const data = await response.json();
      setSessions(data.data || []);
    } catch {
      // Fallback to mock data
      console.warn('Failed to fetch sessions, using mock data');
      setSessions(getMockSessions());
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchSessions();
  }, [fetchSessions]);

  const handleCreate = useCallback(async () => {
    const trimmed = projectName.trim();
    if (trimmed && !isCreating) {
      setIsCreating(true);
      try {
        await onCreateProject(trimmed);
      } finally {
        setProjectName('');
        setIsCreating(false);
      }
    }
  }, [projectName, isCreating, onCreateProject]);

  // Filter sessions
  const filteredSessions = useMemo(() => {
    return sessions.filter((session) => {
      // Search filter (case-insensitive project name)
      if (searchQuery && !session.projectName.toLowerCase().includes(searchQuery.toLowerCase())) {
        return false;
      }
      // Status filter
      if (statusFilter !== 'all' && session.status !== statusFilter) {
        return false;
      }
      // Date filter
      if (dateFilter !== 'all') {
        const created = new Date(session.createdAt);
        const now = new Date();
        const todayStart = new Date(now.getFullYear(), now.getMonth(), now.getDate());
        const weekStart = new Date(todayStart);
        weekStart.setDate(weekStart.getDate() - todayStart.getDay());
        const monthStart = new Date(now.getFullYear(), now.getMonth(), 1);

        if (dateFilter === 'today' && created < todayStart) return false;
        if (dateFilter === 'week' && created < weekStart) return false;
        if (dateFilter === 'month' && created < monthStart) return false;
      }
      return true;
    });
  }, [sessions, searchQuery, statusFilter, dateFilter]);

  const formatDate = (dateStr: string) => {
    const date = new Date(dateStr);
    if (isNaN(date.getTime())) return 'Unknown date';
    const now = new Date();
    const diffMs = now.getTime() - date.getTime();
    if (diffMs < 0) return 'Just now';
    const diffMins = Math.floor(diffMs / 60000);
    const diffHours = Math.floor(diffMs / 3600000);
    const diffDays = Math.floor(diffMs / 86400000);

    if (diffMins < 1) return 'Just now';
    if (diffMins < 60) return `${diffMins}m ago`;
    if (diffHours < 24) return `${diffHours}h ago`;
    if (diffDays < 7) return `${diffDays}d ago`;
    return date.toLocaleDateString();
  };

  return (
    <div className="session-explorer" style={{ padding: '12px' }}>
      <h3 style={{ margin: '0 0 12px 0', fontSize: '14px', fontWeight: 600 }}>Workspace</h3>

      {/* New Project Form */}
      <div className="new-project-form" style={{ marginBottom: '16px' }}>
        <input
          type="text"
          placeholder="Project name"
          value={projectName}
          onChange={(e) => setProjectName(e.target.value)}
          onKeyDown={(e) => e.key === 'Enter' && handleCreate()}
          style={{
            width: '100%',
            padding: '8px',
            marginBottom: '8px',
            border: '1px solid var(--bg-tertiary)',
            borderRadius: '6px',
            background: 'var(--bg-primary)',
            color: 'var(--text-primary)',
            fontSize: '13px',
            boxSizing: 'border-box',
          }}
        />
        <button
          onClick={handleCreate}
          disabled={isCreating || !projectName.trim()}
          style={{
            width: '100%',
            padding: '8px',
            background: projectName.trim() ? 'var(--accent)' : 'var(--bg-tertiary)',
            color: 'white',
            border: 'none',
            borderRadius: '6px',
            cursor: projectName.trim() ? 'pointer' : 'not-allowed',
            fontSize: '13px',
            fontWeight: 500,
            transition: 'background-color 200ms',
          }}
        >
          {isCreating ? 'Creating...' : 'New Project'}
        </button>
      </div>

      {/* Session List Header */}
      <h4 style={{ margin: '0 0 8px 0', fontSize: '13px', fontWeight: 600, color: 'var(--text-secondary)' }}>
        Sessions ({filteredSessions.length})
      </h4>

      {/* Search Input */}
      <input
        type="text"
        placeholder="Search by project name..."
        value={searchQuery}
        onChange={(e) => setSearchQuery(e.target.value)}
        style={{
          width: '100%',
          padding: '6px 8px',
          marginBottom: '8px',
          border: '1px solid var(--bg-tertiary)',
          borderRadius: '6px',
          background: 'var(--bg-primary)',
          color: 'var(--text-primary)',
          fontSize: '12px',
          boxSizing: 'border-box',
        }}
      />

      {/* Filters */}
      <div style={{ display: 'flex', gap: '8px', marginBottom: '12px' }}>
        <select
          value={statusFilter}
          onChange={(e) => setStatusFilter(e.target.value as StatusFilter)}
          style={{
            flex: 1,
            padding: '4px 6px',
            border: '1px solid var(--bg-tertiary)',
            borderRadius: '4px',
            background: 'var(--bg-primary)',
            color: 'var(--text-secondary)',
            fontSize: '11px',
          }}
        >
          <option value="all">All Status</option>
          <option value="running">Running</option>
          <option value="complete">Complete</option>
          <option value="failed">Failed</option>
          <option value="paused">Paused</option>
        </select>
        <select
          value={dateFilter}
          onChange={(e) => setDateFilter(e.target.value as DateFilter)}
          style={{
            flex: 1,
            padding: '4px 6px',
            border: '1px solid var(--bg-tertiary)',
            borderRadius: '4px',
            background: 'var(--bg-primary)',
            color: 'var(--text-secondary)',
            fontSize: '11px',
          }}
        >
          <option value="all">All Time</option>
          <option value="today">Today</option>
          <option value="week">This Week</option>
          <option value="month">This Month</option>
        </select>
      </div>

      {/* Session List */}
      {loading ? (
        <div style={{ textAlign: 'center', padding: '16px', color: 'var(--text-secondary)', fontSize: '12px' }}>
          Loading sessions...
        </div>
      ) : filteredSessions.length === 0 ? (
        <div style={{ textAlign: 'center', padding: '16px', color: 'var(--text-secondary)', fontSize: '12px' }}>
          No sessions found
        </div>
      ) : (
        <div style={{ display: 'flex', flexDirection: 'column', gap: '8px' }}>
          {filteredSessions.map((session) => (
            <div
              key={session.sessionId}
              style={{
                background: 'var(--bg-primary)',
                border: '1px solid var(--bg-tertiary)',
                borderRadius: '8px',
                padding: '10px',
                cursor: 'pointer',
                transition: 'border-color 200ms',
              }}
              onClick={() => onSelectSession?.(session.sessionId)}
              onMouseEnter={(e) => {
                e.currentTarget.style.borderColor = 'var(--accent)';
              }}
              onMouseLeave={(e) => {
                e.currentTarget.style.borderColor = 'var(--bg-tertiary)';
              }}
            >
              {/* Session Header */}
              <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '6px' }}>
                <span style={{ fontSize: '13px', fontWeight: 500, color: 'var(--text-primary)' }}>
                  {session.projectName}
                </span>
                <span
                  style={{
                    fontSize: '10px',
                    padding: '2px 6px',
                    borderRadius: '4px',
                    background: `${statusColors[session.status]}20`,
                    color: statusColors[session.status],
                    fontWeight: 500,
                  }}
                >
                  {session.status}
                </span>
              </div>

              {/* Session Meta */}
              <div style={{ fontSize: '11px', color: 'var(--text-secondary)', marginBottom: '8px' }}>
                Created: {formatDate(session.createdAt)}
              </div>

              {/* Action Buttons */}
              <div style={{ display: 'flex', gap: '6px' }}>
                {session.status !== 'running' && (
                  <button
                    onClick={(e) => {
                      e.stopPropagation();
                      onResumeSession?.(session.sessionId);
                    }}
                    style={{
                      flex: 1,
                      padding: '4px 8px',
                      fontSize: '11px',
                      background: 'var(--accent)',
                      color: 'white',
                      border: 'none',
                      borderRadius: '4px',
                      cursor: 'pointer',
                    }}
                  >
                    Resume
                  </button>
                )}
                <button
                  onClick={(e) => {
                    e.stopPropagation();
                    onForkSession?.(session.sessionId);
                  }}
                  style={{
                    flex: 1,
                    padding: '4px 8px',
                    fontSize: '11px',
                    background: 'var(--bg-tertiary)',
                    color: 'var(--text-secondary)',
                    border: '1px solid var(--bg-tertiary)',
                    borderRadius: '4px',
                    cursor: 'pointer',
                    transition: 'color 200ms, border-color 200ms',
                  }}
                  onMouseEnter={(e) => {
                    e.currentTarget.style.color = 'var(--text-primary)';
                    e.currentTarget.style.borderColor = 'var(--accent)';
                  }}
                  onMouseLeave={(e) => {
                    e.currentTarget.style.color = 'var(--text-secondary)';
                    e.currentTarget.style.borderColor = 'var(--bg-tertiary)';
                  }}
                >
                  Fork
                </button>
                <button
                  onClick={(e) => {
                    e.stopPropagation();
                    onExportSession?.(session.sessionId);
                  }}
                  style={{
                    flex: 1,
                    padding: '4px 8px',
                    fontSize: '11px',
                    background: 'var(--bg-tertiary)',
                    color: 'var(--text-secondary)',
                    border: '1px solid var(--bg-tertiary)',
                    borderRadius: '4px',
                    cursor: 'pointer',
                    transition: 'color 200ms, border-color 200ms',
                  }}
                  onMouseEnter={(e) => {
                    e.currentTarget.style.color = 'var(--text-primary)';
                    e.currentTarget.style.borderColor = 'var(--accent)';
                  }}
                  onMouseLeave={(e) => {
                    e.currentTarget.style.color = 'var(--text-secondary)';
                    e.currentTarget.style.borderColor = 'var(--bg-tertiary)';
                  }}
                >
                  Export
                </button>
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
};

// Mock data for when API is not available
function getMockSessions(): Session[] {
  const now = new Date();
  return [
    {
      sessionId: 'mock-1',
      projectName: 'nfv2-auth-module',
      status: 'complete',
      createdAt: new Date(now.getTime() - 2 * 3600000).toISOString(),
      lastActive: new Date(now.getTime() - 1800000).toISOString(),
    },
    {
      sessionId: 'mock-2',
      projectName: 'api-integration',
      status: 'running',
      createdAt: new Date(now.getTime() - 300000).toISOString(),
      lastActive: new Date(now.getTime() - 60000).toISOString(),
    },
    {
      sessionId: 'mock-3',
      projectName: 'dashboard-refactor',
      status: 'failed',
      createdAt: new Date(now.getTime() - 86400000).toISOString(),
      lastActive: new Date(now.getTime() - 82800000).toISOString(),
    },
    {
      sessionId: 'mock-4',
      projectName: 'test-suite',
      status: 'paused',
      createdAt: new Date(now.getTime() - 7 * 86400000).toISOString(),
      lastActive: new Date(now.getTime() - 2 * 86400000).toISOString(),
    },
  ];
}

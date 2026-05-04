import React, { useState, useCallback, useEffect, useMemo } from 'react';
import { useSession, Session } from '../../hooks/useSession';
import { EmptyState } from '../ui/EmptyState';

interface SessionExplorerProps {
  onCreateProject: (projectName: string) => void;
  onResumeSession?: (sessionId: string) => void;
  onForkSession?: (sessionId: string) => void;
  onExportSession?: (sessionId: string) => void;
  onSelectSession?: (sessionId: string) => void;
  onStartChat?: () => void;
}

type StatusFilter = 'all' | 'running' | 'complete' | 'failed' | 'paused' | 'zombie';
type DateFilter = 'all' | 'today' | 'week' | 'month';

const statusColors: Record<Session['status'], string> = {
  running: '#06b6d4',
  complete: '#22c55e',
  failed: '#ef4444',
  paused: '#ff9800',
  zombie: '#8b5cf6',
};

export const SessionExplorer: React.FC<SessionExplorerProps> = ({
  onCreateProject,
  onResumeSession,
  onForkSession,
  onExportSession,
  onSelectSession,
  onStartChat,
}) => {
  const { sessions, loading: sessionsLoading, listSessions } = useSession();
  const [projectName, setProjectName] = useState('');
  const [isCreating, setIsCreating] = useState(false);
  const [searchQuery, setSearchQuery] = useState('');
  const [statusFilter, setStatusFilter] = useState<StatusFilter>('all');
  const [dateFilter, setDateFilter] = useState<DateFilter>('all');

  useEffect(() => {
    listSessions();
  }, [listSessions]);

  const handleCreate = useCallback(async () => {
    const trimmed = projectName.trim();
    if (trimmed && !isCreating) {
      setIsCreating(true);
      try {
        await onCreateProject(trimmed);
        setProjectName('');
        await listSessions();
      } catch (err) {
        console.error('Failed to create project:', err);
      } finally {
        setIsCreating(false);
      }
    }
  }, [projectName, isCreating, onCreateProject, listSessions]);

  // Filter sessions
  const filteredSessions = useMemo(() => {
    const now = new Date();
    const todayStart = new Date(now.getFullYear(), now.getMonth(), now.getDate());
    const weekStart = new Date(todayStart);
    weekStart.setDate(weekStart.getDate() - todayStart.getDay());
    const monthStart = new Date(now.getFullYear(), now.getMonth(), 1);

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
          aria-label="Project name"
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
          aria-label="Create new project"
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
            transition: 'background-color 200ms ease-out',
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
        aria-label="Search sessions"
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
          aria-label="Filter by status"
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
          <option value="zombie">Zombie</option>
        </select>
        <select
          value={dateFilter}
          onChange={(e) => setDateFilter(e.target.value as DateFilter)}
          aria-label="Filter by date"
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
      {sessionsLoading ? (
        <EmptyState
          icon={<span aria-hidden="true" className="animate-spin" style={{ display: 'inline-block', fontSize: '16px' }}>&#8987;</span>}
          title="Loading sessions..."
          animated
        />
      ) : sessions.length === 0 ? (
        <EmptyState
          icon={<span aria-hidden="true">&#128230;</span>}
          title="No sessions yet"
          description="Start a new project to begin your journey"
          actionLabel="Start Chat"
          onAction={onStartChat}
        />
      ) : filteredSessions.length === 0 ? (
        <div style={{ textAlign: 'center', padding: '16px', color: 'var(--text-secondary)', fontSize: '12px' }}>
          No sessions match your filters
        </div>
      ) : (
        <div style={{ display: 'flex', flexDirection: 'column', gap: '8px' }}>
          {filteredSessions.map((session) => (
            <div
              key={session.sessionId}
              className="session-card"
              tabIndex={0}
              role="button"
              aria-label={`Session: ${session.projectName}, status: ${session.status}`}
              onClick={() => onSelectSession?.(session.sessionId)}
              onKeyDown={(e) => { if (e.key === 'Enter' || e.key === ' ') { e.preventDefault(); onSelectSession?.(session.sessionId); } }}
              style={{
                padding: '10px',
                borderRadius: '6px',
                border: '1px solid var(--bg-tertiary)',
                cursor: 'pointer',
                transition: 'border-color 200ms ease-out, box-shadow 200ms ease-out',
              }}
              onMouseEnter={(e) => {
                (e.currentTarget as HTMLDivElement).style.borderColor = 'var(--accent)';
                (e.currentTarget as HTMLDivElement).style.boxShadow = '0 0 0 1px var(--accent)';
              }}
              onMouseLeave={(e) => {
                (e.currentTarget as HTMLDivElement).style.borderColor = 'var(--bg-tertiary)';
                (e.currentTarget as HTMLDivElement).style.boxShadow = 'none';
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
              {session.goal && (
                <div
                  style={{
                    fontSize: '11px',
                    color: 'var(--text-secondary)',
                    marginBottom: '8px',
                    overflow: 'hidden',
                    textOverflow: 'ellipsis',
                    whiteSpace: 'nowrap',
                  }}
                  title={session.goal}
                >
                  Goal: {session.goal}
                </div>
              )}

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
                      background: session.status === 'zombie' ? '#8b5cf6' : 'var(--accent)',
                      color: 'white',
                      border: 'none',
                      borderRadius: '4px',
                      cursor: 'pointer',
                      transition: 'opacity 200ms ease-out',
                    }}
                    onMouseEnter={(e) => {
                      (e.currentTarget as HTMLButtonElement).style.opacity = '0.85';
                    }}
                    onMouseLeave={(e) => {
                      (e.currentTarget as HTMLButtonElement).style.opacity = '1';
                    }}
                  >
                    {session.status === 'zombie' ? 'Clean Up' : 'Resume'}
                  </button>
                )}
                <button
                  onClick={(e) => {
                    e.stopPropagation();
                    onForkSession?.(session.sessionId);
                  }}
                  className="btn-secondary"
                  style={{
                    flex: 1,
                    padding: '4px 8px',
                    fontSize: '11px',
                    borderRadius: '4px',
                    cursor: 'pointer',
                  }}
                >
                  Fork
                </button>
                {/* Export button: visible when session is complete (all nodes green) */}
                {session.status === 'complete' && onExportSession && (
                  <button
                    onClick={(e) => {
                      e.stopPropagation();
                      onExportSession(session.sessionId);
                    }}
                    aria-label={`Export session ${session.projectName}`}
                    style={{
                      flex: 1,
                      padding: '4px 8px',
                      fontSize: '11px',
                      background: 'transparent',
                      color: '#22C55E',
                      border: '1px solid #22C55E',
                      borderRadius: '4px',
                      cursor: 'pointer',
                      transition: 'background-color 200ms ease-out, color 200ms ease-out',
                    }}
                    onMouseEnter={(e) => {
                      (e.currentTarget as HTMLButtonElement).style.background = '#22C55E';
                      (e.currentTarget as HTMLButtonElement).style.color = '#020617';
                    }}
                    onMouseLeave={(e) => {
                      (e.currentTarget as HTMLButtonElement).style.background = 'transparent';
                      (e.currentTarget as HTMLButtonElement).style.color = '#22C55E';
                    }}
                  >
                    Export
                  </button>
                )}
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
};

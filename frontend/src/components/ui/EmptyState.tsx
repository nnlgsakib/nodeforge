import React from 'react';

interface EmptyStateProps {
  icon: React.ReactNode;
  title: string;
  description?: string;
  actionLabel?: string;
  onAction?: () => void;
  animated?: boolean;
}

/**
 * Reusable empty state component for panels.
 * Renders an icon, title, optional description, and optional action button.
 * Supports animated ellipsis for loading/waiting states.
 */
export const EmptyState: React.FC<EmptyStateProps> = ({
  icon,
  title,
  description,
  actionLabel,
  onAction,
  animated = false,
}) => (
  <div
    className="empty-state"
    role="status"
    aria-live="polite"
    style={{
      display: 'flex',
      flexDirection: 'column',
      alignItems: 'center',
      justifyContent: 'center',
      padding: '24px 16px',
      textAlign: 'center',
      minHeight: '120px',
    }}
  >
    <div
      className="empty-state-icon"
      style={{
        fontSize: '32px',
        marginBottom: '12px',
        color: 'var(--text-secondary)',
        opacity: 0.6,
      }}
      aria-hidden="true"
    >
      {icon}
    </div>
    <div
      className="empty-state-title"
      style={{
        fontSize: '13px',
        fontWeight: 500,
        color: 'var(--text-primary)',
        marginBottom: description ? '4px' : '12px',
      }}
    >
      {title}
    </div>
    {description && (
      <div
        className="empty-state-description"
        style={{
          fontSize: '12px',
          color: 'var(--text-secondary)',
          marginBottom: actionLabel ? '16px' : '0',
        }}
      >
        {description}
      </div>
    )}
    {actionLabel && onAction && (
      <button
        className="empty-state-action"
        onClick={onAction}
        style={{
          padding: '6px 16px',
          fontSize: '12px',
          fontWeight: 500,
          background: 'var(--accent)',
          color: 'white',
          border: 'none',
          borderRadius: '6px',
          cursor: 'pointer',
          transition: 'background-color 200ms',
        }}
      >
        {actionLabel}
      </button>
    )}
    {animated && (
      <span
        className="empty-state-animated-ellipsis"
        aria-hidden="true"
        style={{
          fontSize: '18px',
          color: 'var(--text-secondary)',
          marginTop: '8px',
          animation: 'pulse 1.5s infinite',
        }}
      >
        …
      </span>
    )}
  </div>
);

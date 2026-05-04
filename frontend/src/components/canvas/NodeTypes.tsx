/* eslint-disable react-refresh/only-export-components */
import React, { memo, useCallback } from 'react';
import { Handle, Position, type NodeProps } from '@xyflow/react';
import {
  type NodeStatus,
  NODE_COLORS,
  STATUS_COLORS,
} from '../../types/nodes';
import { useTheme } from '../../hooks/use-theme';

interface NodeData {
  label?: string;
  status?: NodeStatus;
  progress?: number;
}

function getNodeData(data: NodeProps['data']): NodeData {
  return (data || {}) as NodeData;
}

// Resolve effective colors: status override wins, otherwise node type colors
function resolveColors(nodeType: string, status?: NodeStatus, isHighContrast?: boolean) {
  const isRunning = status === 'running';

  // High-contrast mode override (AC:1, Subtask 1.6)
  if (isHighContrast) {
    const hcColors: Record<string, string> = {
      goal: '#00ff00',
      spec: '#00aaff',
      plan: '#ff00ff',
      implement: '#ff8800',
      test: '#ffff00',
      review: '#00ffff',
    };
    const bg = hcColors[nodeType] || '#ffffff';
    const statusOverride = status && status !== 'pending';
    const effectiveBg = statusOverride
      ? status === 'failed' ? '#ff0000'
      : status === 'running' ? '#ffff00'
      : status === 'complete' ? '#00ff00'
      : bg
      : bg;
    return {
      background: effectiveBg,
      border: `2px solid ${effectiveBg}`,
      borderColor: effectiveBg,
      boxShadow: `0 0 12px ${effectiveBg}`,
      isRunning,
      isHighContrast,
    };
  }

  const typeColors = NODE_COLORS[nodeType as keyof typeof NODE_COLORS];
  const statusColors = STATUS_COLORS[status || 'pending'];

  if (status && status !== 'pending' && statusColors.background) {
    return {
      background: statusColors.background,
      border: `2px solid ${statusColors.border}`,
      borderColor: statusColors.border,
      boxShadow: `0 4px 12px ${statusColors.glow}`,
      isRunning,
      isHighContrast: false,
    };
  }

  const bg = typeColors?.background ?? '#94a3b8';
  const border = typeColors?.border ?? '#64748b';
  const glow = typeColors?.glow ?? 'rgba(148, 163, 184, 0.33)';

  return {
    background: bg,
    border: `2px solid ${border}`,
    borderColor: border,
    boxShadow: `0 4px 12px ${glow}`,
    isRunning,
    isHighContrast: false,
  };
}

// Shared handle style for DaVinci-style pins
const handleStyle = (color: string): React.CSSProperties => ({
  width: '10px',
  height: '10px',
  background: color,
  border: '1px solid rgba(255,255,255,0.4)',
  borderRadius: '50%',
});

// ARIA live region for screen reader announcements
function StatusAnnouncer({ status, label }: { status?: NodeStatus; label: string }) {
  if (!status || status === 'pending') return null;
  return (
    <span aria-live="polite" className="sr-only">
      {`${label} node status: ${status}`}
    </span>
  );
}

// Progress bar for running nodes
function ProgressBar({ progress }: { progress: number }) {
  const safeProgress = Number.isFinite(progress) ? progress : 0;
  const clamped = Math.max(0, Math.min(1, safeProgress));
  return (
    <div
      style={{
        marginTop: '8px',
        height: '4px',
        background: 'rgba(255,255,255,0.3)',
        borderRadius: '2px',
        overflow: 'hidden',
      }}
      role="progressbar"
      aria-valuenow={Math.round(clamped * 100)}
      aria-valuemin={0}
      aria-valuemax={100}
    >
      <div
        style={{
          width: `${clamped * 100}%`,
          height: '100%',
          background: 'white',
          borderRadius: '2px',
          transition: 'width 0.3s ease',
        }}
      />
    </div>
  );
}

// Goal Node - Green rounded rectangle, input pin only
const GoalNodeComponent: React.FC<NodeProps> = ({ data, selected }) => {
  const nodeData = getNodeData(data);
  const { isHighContrast: hc } = useTheme();
  const colors = resolveColors('goal', nodeData.status, hc);
  const textColor = hc ? '#000000' : (nodeData.status === 'pending' ? '#1a1b1e' : 'white');
  const handleKeyDown = useCallback((e: React.KeyboardEvent) => {
    if (e.key === 'Enter' || e.key === ' ') {
      e.preventDefault();
      (e.target as HTMLElement).click();
    }
  }, []);

  return (
    <div
      style={{
        borderRadius: '12px',
        padding: '16px 24px',
        minWidth: '140px',
        textAlign: 'center',
        color: textColor,
        fontWeight: 600,
        fontSize: '14px',
        outline: selected ? '2px solid white' : 'none',
        outlineOffset: '2px',
        background: colors.background,
        border: colors.isRunning
          ? '3px solid #FFC107'
          : colors.border,
        boxShadow: colors.isRunning
          ? '0 0 8px #FFC107'
          : colors.boxShadow,
        animation: colors.isRunning
          ? (colors.isHighContrast ? 'node-pulse-hc 300ms infinite alternate' : 'node-pulse 300ms infinite alternate')
          : 'none',
      }}
      role="group"
      aria-label={`Node ${nodeData.label || 'Goal'}, status: ${nodeData.status || 'pending'}`}
      tabIndex={0}
      onKeyDown={handleKeyDown}
    >
      <Handle
        type="target"
        position={Position.Top}
        style={handleStyle(colors.borderColor)}
        aria-label="Goal input"
      />
      <div>{nodeData.label || 'Goal'}</div>
      {nodeData.progress !== undefined && nodeData.status === 'running' && (
        <ProgressBar progress={nodeData.progress} />
      )}
      <StatusAnnouncer status={nodeData.status} label="Goal" />
    </div>
  );
};

// Spec Node - Blue diamond shape, input/output pins
const SpecNodeComponent: React.FC<NodeProps> = ({ data, selected }) => {
  const nodeData = getNodeData(data);
  const { isHighContrast: hc } = useTheme();
  const colors = resolveColors('spec', nodeData.status, hc);
  const textColor = hc ? '#000000' : 'white';
  const handleKeyDown = useCallback((e: React.KeyboardEvent) => {
    if (e.key === 'Enter' || e.key === ' ') {
      e.preventDefault();
      (e.target as HTMLElement).click();
    }
  }, []);

  return (
    <div
      style={{
        width: '120px',
        height: '120px',
        transform: 'rotate(45deg)',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        color: textColor,
        fontWeight: 600,
        fontSize: '14px',
        outline: selected ? '2px solid white' : 'none',
        outlineOffset: '4px',
        background: colors.background,
        border: colors.isRunning
          ? '3px solid #FFC107'
          : colors.border,
        boxShadow: colors.isRunning
          ? '0 0 8px #FFC107'
          : colors.boxShadow,
        animation: colors.isRunning
          ? (colors.isHighContrast ? 'node-pulse-hc 300ms infinite alternate' : 'node-pulse 300ms infinite alternate')
          : 'none',
      }}
      role="group"
      aria-label={`Node ${nodeData.label || 'Spec'}, status: ${nodeData.status || 'pending'}`}
      tabIndex={0}
      onKeyDown={handleKeyDown}
    >
      <Handle
        type="target"
        position={Position.Top}
        style={{ ...handleStyle(colors.borderColor), transform: 'rotate(-45deg)' }}
        aria-label="Spec input"
      />
      <div style={{ transform: 'rotate(-45deg)' }}>{nodeData.label || 'Spec'}</div>
      <Handle
        type="source"
        position={Position.Bottom}
        style={{ ...handleStyle(colors.borderColor), transform: 'rotate(-45deg)' }}
        aria-label="Spec output"
      />

      <StatusAnnouncer status={nodeData.status} label="Spec" />
      {nodeData.progress !== undefined && Number.isFinite(nodeData.progress) && nodeData.status === 'running' && (
        <div style={{ position: 'absolute', bottom: '-8px', left: '50%', transform: 'translateX(-50%)', width: '80px' }}>
          <ProgressBar progress={nodeData.progress} />
        </div>
      )}
    </div>
  );
};

// Plan Node - Purple rounded rectangle, input/output pins
const PlanNodeComponent: React.FC<NodeProps> = ({ data, selected }) => {
  const nodeData = getNodeData(data);
  const { isHighContrast: hc } = useTheme();
  const colors = resolveColors('plan', nodeData.status, hc);
  const textColor = hc ? '#000000' : (nodeData.status === 'pending' ? '#1a1b1e' : 'white');
  const handleKeyDown = useCallback((e: React.KeyboardEvent) => {
    if (e.key === 'Enter' || e.key === ' ') {
      e.preventDefault();
      (e.target as HTMLElement).click();
    }
  }, []);

  return (
    <div
      style={{
        borderRadius: '10px',
        padding: '14px 22px',
        minWidth: '130px',
        textAlign: 'center',
        color: textColor,
        fontWeight: 600,
        fontSize: '14px',
        outline: selected ? '2px solid white' : 'none',
        outlineOffset: '2px',
        background: colors.background,
        border: colors.isRunning
          ? '3px solid #FFC107'
          : colors.border,
        boxShadow: colors.isRunning
          ? '0 0 8px #FFC107'
          : colors.boxShadow,
        animation: colors.isRunning
          ? (colors.isHighContrast ? 'node-pulse-hc 300ms infinite alternate' : 'node-pulse 300ms infinite alternate')
          : 'none',
      }}
      role="group"
      aria-label={`Node ${nodeData.label || 'Plan'}, status: ${nodeData.status || 'pending'}`}
      tabIndex={0}
      onKeyDown={handleKeyDown}
    >
      <Handle
        type="target"
        position={Position.Top}
        style={handleStyle(colors.borderColor)}
        aria-label="Plan input"
      />
      <div>{nodeData.label || 'Plan'}</div>
      {nodeData.progress !== undefined && nodeData.status === 'running' && (
        <ProgressBar progress={nodeData.progress} />
      )}
      <Handle
        type="source"
        position={Position.Bottom}
        style={handleStyle(colors.borderColor)}
        aria-label="Plan output"
      />
      <StatusAnnouncer status={nodeData.status} label="Plan" />
    </div>
  );
};

// Implement Node - Orange rectangle, input/output pins
const ImplementNodeComponent: React.FC<NodeProps> = ({ data, selected }) => {
  const nodeData = getNodeData(data);
  const { isHighContrast: hc } = useTheme();
  const colors = resolveColors('implement', nodeData.status, hc);
  const textColor = hc ? '#000000' : (nodeData.status === 'pending' ? '#1a1b1e' : 'white');
  const handleKeyDown = useCallback((e: React.KeyboardEvent) => {
    if (e.key === 'Enter' || e.key === ' ') {
      e.preventDefault();
      (e.target as HTMLElement).click();
    }
  }, []);

  return (
    <div
      style={{
        borderRadius: '6px',
        padding: '14px 22px',
        minWidth: '130px',
        textAlign: 'center',
        color: textColor,
        fontWeight: 600,
        fontSize: '14px',
        outline: selected ? '2px solid white' : 'none',
        outlineOffset: '2px',
        background: colors.background,
        border: colors.isRunning
          ? '3px solid #FFC107'
          : colors.border,
        boxShadow: colors.isRunning
          ? '0 0 8px #FFC107'
          : colors.boxShadow,
        animation: colors.isRunning
          ? (colors.isHighContrast ? 'node-pulse-hc 300ms infinite alternate' : 'node-pulse 300ms infinite alternate')
          : 'none',
      }}
      role="group"
      aria-label={`Node ${nodeData.label || 'Implement'}, status: ${nodeData.status || 'pending'}`}
      tabIndex={0}
      onKeyDown={handleKeyDown}
    >
      <Handle
        type="target"
        position={Position.Top}
        style={handleStyle(colors.borderColor)}
        aria-label="Implement input"
      />
      <div>{nodeData.label || 'Implement'}</div>
      {nodeData.progress !== undefined && nodeData.status === 'running' && (
        <ProgressBar progress={nodeData.progress} />
      )}
      <Handle
        type="source"
        position={Position.Bottom}
        style={handleStyle(colors.borderColor)}
        aria-label="Implement output"
      />
      <StatusAnnouncer status={nodeData.status} label="Implement" />
    </div>
  );
};

// Test Node - Yellow rounded rectangle, input/output pins
const TestNodeComponent: React.FC<NodeProps> = ({ data, selected }) => {
  const nodeData = getNodeData(data);
  const { isHighContrast: hc } = useTheme();
  const colors = resolveColors('test', nodeData.status, hc);
  const textColor = hc ? '#000000' : (nodeData.status === 'pending' || nodeData.status === 'running' ? '#1a1b1e' : 'white');
  const handleKeyDown = useCallback((e: React.KeyboardEvent) => {
    if (e.key === 'Enter' || e.key === ' ') {
      e.preventDefault();
      (e.target as HTMLElement).click();
    }
  }, []);

  return (
    <div
      style={{
        borderRadius: '10px',
        padding: '14px 22px',
        minWidth: '120px',
        textAlign: 'center',
        color: textColor,
        fontWeight: 600,
        fontSize: '14px',
        outline: selected ? '2px solid white' : 'none',
        outlineOffset: '2px',
        background: colors.background,
        border: colors.isRunning
          ? '3px solid #FFC107'
          : colors.border,
        boxShadow: colors.isRunning
          ? '0 0 8px #FFC107'
          : colors.boxShadow,
        animation: colors.isRunning
          ? (colors.isHighContrast ? 'node-pulse-hc 300ms infinite alternate' : 'node-pulse 300ms infinite alternate')
          : 'none',
      }}
      role="group"
      aria-label={`Node ${nodeData.label || 'Test'}, status: ${nodeData.status || 'pending'}`}
      tabIndex={0}
      onKeyDown={handleKeyDown}
    >
      <Handle
        type="target"
        position={Position.Top}
        style={handleStyle(colors.borderColor)}
        aria-label="Test input"
      />
      <div>{nodeData.label || 'Test'}</div>
      {nodeData.progress !== undefined && nodeData.status === 'running' && (
        <ProgressBar progress={nodeData.progress} />
      )}
      <Handle
        type="source"
        position={Position.Bottom}
        style={handleStyle(colors.borderColor)}
        aria-label="Test output"
      />
      <StatusAnnouncer status={nodeData.status} label="Test" />
    </div>
  );
};

// Review Node - Cyan rectangle, input/output pins
const ReviewNodeComponent: React.FC<NodeProps> = ({ data, selected }) => {
  const nodeData = getNodeData(data);
  const { isHighContrast: hc } = useTheme();
  const colors = resolveColors('review', nodeData.status, hc);
  const textColor = hc ? '#000000' : (nodeData.status === 'pending' ? '#1a1b1e' : 'white');
  const handleKeyDown = useCallback((e: React.KeyboardEvent) => {
    if (e.key === 'Enter' || e.key === ' ') {
      e.preventDefault();
      (e.target as HTMLElement).click();
    }
  }, []);

  return (
    <div
      style={{
        borderRadius: '6px',
        padding: '14px 22px',
        minWidth: '120px',
        textAlign: 'center',
        color: textColor,
        fontWeight: 600,
        fontSize: '14px',
        outline: selected ? '2px solid white' : 'none',
        outlineOffset: '2px',
        background: colors.background,
        border: colors.isRunning
          ? '3px solid #FFC107'
          : colors.border,
        boxShadow: colors.isRunning
          ? '0 0 8px #FFC107'
          : colors.boxShadow,
        animation: colors.isRunning
          ? (colors.isHighContrast ? 'node-pulse-hc 300ms infinite alternate' : 'node-pulse 300ms infinite alternate')
          : 'none',
      }}
      role="group"
      aria-label={`Node ${nodeData.label || 'Review'}, status: ${nodeData.status || 'pending'}`}
      tabIndex={0}
      onKeyDown={handleKeyDown}
    >
      <Handle
        type="target"
        position={Position.Top}
        style={handleStyle(colors.borderColor)}
        aria-label="Review input"
      />
      <div>{nodeData.label || 'Review'}</div>
      {nodeData.progress !== undefined && nodeData.status === 'running' && (
        <ProgressBar progress={nodeData.progress} />
      )}
      <Handle
        type="source"
        position={Position.Bottom}
        style={handleStyle(colors.borderColor)}
        aria-label="Review output"
      />
      <StatusAnnouncer status={nodeData.status} label="Review" />
    </div>
  );
};

export const nodeTypes = {
  goal: memo(GoalNodeComponent),
  spec: memo(SpecNodeComponent),
  plan: memo(PlanNodeComponent),
  implement: memo(ImplementNodeComponent),
  test: memo(TestNodeComponent),
  review: memo(ReviewNodeComponent),
};

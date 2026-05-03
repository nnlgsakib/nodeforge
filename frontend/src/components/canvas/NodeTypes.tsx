/* eslint-disable react-refresh/only-export-components */
import React, { memo, useCallback } from 'react';
import { Handle, Position, type NodeProps } from '@xyflow/react';
import {
  type NodeStatus,
  NODE_COLORS,
  STATUS_COLORS,
} from '../../types/nodes';

interface NodeData {
  label?: string;
  status?: NodeStatus;
  progress?: number;
}

function getNodeData(data: NodeProps['data']): NodeData {
  return (data || {}) as NodeData;
}

// Resolve effective colors: status override wins, otherwise node type colors
function resolveColors(nodeType: string, status?: NodeStatus) {
  const typeColors = NODE_COLORS[nodeType as keyof typeof NODE_COLORS];
  const statusColors = STATUS_COLORS[status || 'pending'];

  if (status && status !== 'pending' && statusColors.background) {
    return {
      background: statusColors.background,
      border: `2px solid ${statusColors.border}`,
      borderColor: statusColors.border,
      boxShadow: `0 4px 12px ${statusColors.glow}`,
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
  const clamped = Math.max(0, Math.min(1, progress));
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
  const colors = resolveColors('goal', nodeData.status);
  const textColor = nodeData.status === 'pending' ? '#1a1b1e' : 'white';
  const handleKeyDown = useCallback((e: React.KeyboardEvent) => {
    if (e.key === 'Enter' || e.key === ' ') {
      e.preventDefault();
      (e.target as HTMLElement).click();
    }
  }, []);

  return (
    <div
      style={{
        ...colors,
        borderRadius: '12px',
        padding: '16px 24px',
        minWidth: '140px',
        textAlign: 'center',
        color: textColor,
        fontWeight: 600,
        fontSize: '14px',
        outline: selected ? '2px solid white' : 'none',
        outlineOffset: '2px',
      }}
      role="group"
      aria-label={`Goal node: ${nodeData.label || 'Goal'}`}
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
  const colors = resolveColors('spec', nodeData.status);
  const handleKeyDown = useCallback((e: React.KeyboardEvent) => {
    if (e.key === 'Enter' || e.key === ' ') {
      e.preventDefault();
      (e.target as HTMLElement).click();
    }
  }, []);

  return (
    <div
      style={{
        ...colors,
        width: '120px',
        height: '120px',
        transform: 'rotate(45deg)',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        color: 'white',
        fontWeight: 600,
        fontSize: '14px',
        outline: selected ? '2px solid white' : 'none',
        outlineOffset: '4px',
      }}
      role="group"
      aria-label={`Spec node: ${nodeData.label || 'Spec'}`}
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
    </div>
  );
};

// Plan Node - Purple rounded rectangle, input/output pins
const PlanNodeComponent: React.FC<NodeProps> = ({ data, selected }) => {
  const nodeData = getNodeData(data);
  const colors = resolveColors('plan', nodeData.status);
  const textColor = nodeData.status === 'pending' ? '#1a1b1e' : 'white';
  const handleKeyDown = useCallback((e: React.KeyboardEvent) => {
    if (e.key === 'Enter' || e.key === ' ') {
      e.preventDefault();
      (e.target as HTMLElement).click();
    }
  }, []);

  return (
    <div
      style={{
        ...colors,
        borderRadius: '10px',
        padding: '14px 22px',
        minWidth: '130px',
        textAlign: 'center',
        color: textColor,
        fontWeight: 600,
        fontSize: '14px',
        outline: selected ? '2px solid white' : 'none',
        outlineOffset: '2px',
      }}
      role="group"
      aria-label={`Plan node: ${nodeData.label || 'Plan'}`}
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
  const colors = resolveColors('implement', nodeData.status);
  const textColor = nodeData.status === 'pending' ? '#1a1b1e' : 'white';
  const handleKeyDown = useCallback((e: React.KeyboardEvent) => {
    if (e.key === 'Enter' || e.key === ' ') {
      e.preventDefault();
      (e.target as HTMLElement).click();
    }
  }, []);

  return (
    <div
      style={{
        ...colors,
        borderRadius: '6px',
        padding: '14px 22px',
        minWidth: '130px',
        textAlign: 'center',
        color: textColor,
        fontWeight: 600,
        fontSize: '14px',
        outline: selected ? '2px solid white' : 'none',
        outlineOffset: '2px',
      }}
      role="group"
      aria-label={`Implement node: ${nodeData.label || 'Implement'}`}
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
  const colors = resolveColors('test', nodeData.status);
  const textColor =
    nodeData.status === 'pending' || nodeData.status === 'running' ? '#1a1b1e' : 'white';
  const handleKeyDown = useCallback((e: React.KeyboardEvent) => {
    if (e.key === 'Enter' || e.key === ' ') {
      e.preventDefault();
      (e.target as HTMLElement).click();
    }
  }, []);

  return (
    <div
      style={{
        ...colors,
        borderRadius: '10px',
        padding: '14px 22px',
        minWidth: '120px',
        textAlign: 'center',
        color: textColor,
        fontWeight: 600,
        fontSize: '14px',
        outline: selected ? '2px solid white' : 'none',
        outlineOffset: '2px',
      }}
      role="group"
      aria-label={`Test node: ${nodeData.label || 'Test'}`}
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
  const colors = resolveColors('review', nodeData.status);
  const textColor = nodeData.status === 'pending' ? '#1a1b1e' : 'white';
  const handleKeyDown = useCallback((e: React.KeyboardEvent) => {
    if (e.key === 'Enter' || e.key === ' ') {
      e.preventDefault();
      (e.target as HTMLElement).click();
    }
  }, []);

  return (
    <div
      style={{
        ...colors,
        borderRadius: '6px',
        padding: '14px 22px',
        minWidth: '120px',
        textAlign: 'center',
        color: textColor,
        fontWeight: 600,
        fontSize: '14px',
        outline: selected ? '2px solid white' : 'none',
        outlineOffset: '2px',
      }}
      role="group"
      aria-label={`Review node: ${nodeData.label || 'Review'}`}
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

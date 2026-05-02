import React, { memo } from 'react';
import { Handle, Position, type NodeProps } from '@xyflow/react';

interface CustomNodeData {
  label?: string;
  status?: 'pending' | 'running' | 'complete' | 'failed' | 'skipped';
  progress?: number;
}

function getNodeData(data: NodeProps['data']): CustomNodeData {
  return (data || {}) as CustomNodeData;
}

// Status-based color mapping (AC2: green=complete, red=failed, yellow=running)
function getStatusColors(status: string, defaultBg: string, defaultBorder: string) {
  switch (status) {
    case 'complete':
      return { background: '#4CAF50', border: '2px solid #2E7D32', borderColor: '#2E7D32', boxShadow: '0 4px 12px rgba(76, 175, 80, 0.5)' };
    case 'failed':
      return { background: '#f44336', border: '2px solid #c62828', borderColor: '#c62828', boxShadow: '0 4px 12px rgba(244, 67, 54, 0.5)' };
    case 'running':
      return { background: '#FFC107', border: '2px solid #F57F17', borderColor: '#F57F17', boxShadow: '0 4px 12px rgba(255, 193, 7, 0.5)' };
    case 'skipped':
      return { background: '#9E9E9E', border: '2px solid #616161', borderColor: '#616161', boxShadow: '0 4px 12px rgba(158, 158, 158, 0.3)' };
    default:
      return { background: defaultBg, border: `2px solid ${defaultBorder}`, borderColor: defaultBorder, boxShadow: `0 4px 12px ${defaultBorder}33` };
  }
}

// Goal Node - Green rounded rectangle
const GoalNode: React.FC<NodeProps> = ({ data }) => {
  const nodeData = getNodeData(data);
  const colors = getStatusColors(nodeData.status || 'pending', '#4CAF50', '#388E3C');
  return (
    <div style={{ ...colors, borderRadius: '12px', padding: '16px 24px', minWidth: '140px', textAlign: 'center', color: 'white', fontWeight: 600, fontSize: '14px' }}>
      <Handle type="target" position={Position.Top} style={{ background: colors.borderColor || '#388E3C' }} />
      <div>{nodeData.label || 'Goal'}</div>
      {nodeData.progress !== undefined && nodeData.status === 'running' && (
        <div style={{ marginTop: '8px', height: '4px', background: 'rgba(255,255,255,0.3)', borderRadius: '2px' }}>
          <div style={{ width: `${nodeData.progress * 100}%`, height: '100%', background: 'white', borderRadius: '2px', transition: 'width 0.3s ease' }} />
        </div>
      )}
      <Handle type="source" position={Position.Bottom} style={{ background: colors.borderColor || '#388E3C' }} />
    </div>
  );
};

// Spec Node - Blue diamond shape
const SpecNode: React.FC<NodeProps> = ({ data }) => {
  const nodeData = getNodeData(data);
  const colors = getStatusColors(nodeData.status || 'pending', '#2196F3', '#1976D2');
  return (
    <div style={{ ...colors, width: '120px', height: '120px', transform: 'rotate(45deg)', display: 'flex', alignItems: 'center', justifyContent: 'center', color: 'white', fontWeight: 600, fontSize: '14px' }}>
      <Handle type="target" position={Position.Top} style={{ background: colors.borderColor || '#1976D2', transform: 'rotate(-45deg)' }} />
      <div style={{ transform: 'rotate(-45deg)' }}>{nodeData.label || 'Spec'}</div>
      <Handle type="source" position={Position.Bottom} style={{ background: colors.borderColor || '#1976D2', transform: 'rotate(-45deg)' }} />
    </div>
  );
};

// Plan Node - Purple rounded rectangle
const PlanNode: React.FC<NodeProps> = ({ data }) => {
  const nodeData = getNodeData(data);
  const colors = getStatusColors(nodeData.status || 'pending', '#9C27B0', '#7B1FAF');
  return (
    <div style={{ ...colors, borderRadius: '10px', padding: '14px 22px', minWidth: '130px', textAlign: 'center', color: 'white', fontWeight: 600, fontSize: '14px' }}>
      <Handle type="target" position={Position.Top} style={{ background: colors.borderColor || '#7B1FAF' }} />
      <div>{nodeData.label || 'Plan'}</div>
      {nodeData.progress !== undefined && nodeData.status === 'running' && (
        <div style={{ marginTop: '8px', height: '4px', background: 'rgba(255,255,255,0.3)', borderRadius: '2px' }}>
          <div style={{ width: `${nodeData.progress * 100}%`, height: '100%', background: 'white', borderRadius: '2px', transition: 'width 0.3s ease' }} />
        </div>
      )}
      <Handle type="source" position={Position.Bottom} style={{ background: colors.borderColor || '#7B1FAF' }} />
    </div>
  );
};

// Implement Node - Orange rectangle
const ImplementNode: React.FC<NodeProps> = ({ data }) => {
  const nodeData = getNodeData(data);
  const colors = getStatusColors(nodeData.status || 'pending', '#FF9800', '#F57C00');
  return (
    <div style={{ ...colors, borderRadius: '6px', padding: '14px 22px', minWidth: '130px', textAlign: 'center', color: 'white', fontWeight: 600, fontSize: '14px' }}>
      <Handle type="target" position={Position.Top} style={{ background: colors.borderColor || '#F57C00' }} />
      <div>{nodeData.label || 'Implement'}</div>
      {nodeData.progress !== undefined && nodeData.status === 'running' && (
        <div style={{ marginTop: '8px', height: '4px', background: 'rgba(255,255,255,0.3)', borderRadius: '2px' }}>
          <div style={{ width: `${nodeData.progress * 100}%`, height: '100%', background: 'white', borderRadius: '2px', transition: 'width 0.3s ease' }} />
        </div>
      )}
      <Handle type="source" position={Position.Bottom} style={{ background: colors.borderColor || '#F57C00' }} />
    </div>
  );
};

// Test Node - Yellow rounded rectangle
const TestNode: React.FC<NodeProps> = ({ data }) => {
  const nodeData = getNodeData(data);
  const colors = getStatusColors(nodeData.status || 'pending', '#FFC107', '#FFA000');
  const textColor = nodeData.status === 'pending' ? '#333' : 'white';
  return (
    <div style={{ ...colors, borderRadius: '10px', padding: '14px 22px', minWidth: '120px', textAlign: 'center', color: textColor, fontWeight: 600, fontSize: '14px' }}>
      <Handle type="target" position={Position.Top} style={{ background: colors.borderColor || '#FFA000' }} />
      <div>{nodeData.label || 'Test'}</div>
      {nodeData.progress !== undefined && nodeData.status === 'running' && (
        <div style={{ marginTop: '8px', height: '4px', background: 'rgba(0,0,0,0.1)', borderRadius: '2px' }}>
          <div style={{ width: `${nodeData.progress * 100}%`, height: '100%', background: '#333', borderRadius: '2px', transition: 'width 0.3s ease' }} />
        </div>
      )}
      <Handle type="source" position={Position.Bottom} style={{ background: colors.borderColor || '#FFA000' }} />
    </div>
  );
};

// Review Node - Cyan rectangle
const ReviewNode: React.FC<NodeProps> = ({ data }) => {
  const nodeData = getNodeData(data);
  const colors = getStatusColors(nodeData.status || 'pending', '#00BCD4', '#00ACC1');
  return (
    <div style={{ ...colors, borderRadius: '6px', padding: '14px 22px', minWidth: '120px', textAlign: 'center', color: 'white', fontWeight: 600, fontSize: '14px' }}>
      <Handle type="target" position={Position.Top} style={{ background: colors.borderColor || '#00ACC1' }} />
      <div>{nodeData.label || 'Review'}</div>
      {nodeData.progress !== undefined && nodeData.status === 'running' && (
        <div style={{ marginTop: '8px', height: '4px', background: 'rgba(255,255,255,0.3)', borderRadius: '2px' }}>
          <div style={{ width: `${nodeData.progress * 100}%`, height: '100%', background: 'white', borderRadius: '2px', transition: 'width 0.3s ease' }} />
        </div>
      )}
      <Handle type="source" position={Position.Bottom} style={{ background: colors.borderColor || '#00ACC1' }} />
    </div>
  );
};

export const nodeTypes = {
  goal: memo(GoalNode),
  spec: memo(SpecNode),
  plan: memo(PlanNode),
  implement: memo(ImplementNode),
  test: memo(TestNode),
  review: memo(ReviewNode),
};

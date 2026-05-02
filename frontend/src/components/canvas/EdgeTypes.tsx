import React from 'react';
import {
  BaseEdge,
  getSmoothStepPath,
  type EdgeProps,
  type Edge,
} from '@xyflow/react';

// Default Edge - #94a3b8, 2px stroke, no animation
const DefaultEdge: React.FC<EdgeProps<Edge>> = ({
  id,
  sourceX,
  sourceY,
  targetX,
  targetY,
  sourcePosition,
  targetPosition,
  style = {},
}) => {
  const [edgePath] = getSmoothStepPath({
    sourceX,
    sourceY,
    sourcePosition,
    targetX,
    targetY,
    targetPosition,
  });

  return (
    <BaseEdge
      id={id}
      path={edgePath}
      style={{
        stroke: '#94a3b8',
        strokeWidth: 2,
        ...style,
      }}
    />
  );
};

// Active Edge - #06b6d4, 3px stroke, animated dash flow
const ActiveEdge: React.FC<EdgeProps<Edge>> = ({
  id,
  sourceX,
  sourceY,
  targetX,
  targetY,
  sourcePosition,
  targetPosition,
  style = {},
}) => {
  const [edgePath] = getSmoothStepPath({
    sourceX,
    sourceY,
    sourcePosition,
    targetX,
    targetY,
    targetPosition,
  });

  return (
    <>
      <BaseEdge
        id={id}
        path={edgePath}
        style={{
          stroke: '#06b6d4',
          strokeWidth: 3,
          strokeDasharray: '12 6',
          animation: 'flow 1s linear infinite',
          ...style,
        }}
      />
      <style>{`
        @keyframes flow {
          to {
            stroke-dashoffset: -18;
          }
        }
      `}</style>
    </>
  );
};

// Tension Edge - #ef4444, 4px stroke
const TensionEdge: React.FC<EdgeProps<Edge>> = ({
  id,
  sourceX,
  sourceY,
  targetX,
  targetY,
  sourcePosition,
  targetPosition,
  style = {},
  selected,
}) => {
  const [edgePath] = getSmoothStepPath({
    sourceX,
    sourceY,
    sourcePosition,
    targetX,
    targetY,
    targetPosition,
  });

  return (
    <BaseEdge
      id={id}
      path={edgePath}
      style={{
        stroke: '#ef4444',
        strokeWidth: 4,
        strokeDasharray: selected ? '8 4' : '16 8',
        opacity: selected ? 1 : 0.7,
        transition: 'stroke-dasharray 0.3s ease, opacity 0.3s ease',
        ...style,
      }}
    />
  );
};

// Success Edge - #22c55e, 2px stroke, brief pulse on completion
const SuccessEdge: React.FC<EdgeProps<Edge>> = ({
  id,
  sourceX,
  sourceY,
  targetX,
  targetY,
  sourcePosition,
  targetPosition,
  style = {},
}) => {
  const [edgePath] = getSmoothStepPath({
    sourceX,
    sourceY,
    sourcePosition,
    targetX,
    targetY,
    targetPosition,
  });

  return (
    <>
      <BaseEdge
        id={id}
        path={edgePath}
        style={{
          stroke: '#22c55e',
          strokeWidth: 2,
          animation: 'pulse-success 0.6s ease-out',
          ...style,
        }}
      />
      <style>{`
        @keyframes pulse-success {
          0% {
            stroke-width: 2;
            opacity: 1;
          }
          50% {
            stroke-width: 4;
            opacity: 0.8;
          }
          100% {
            stroke-width: 2;
            opacity: 1;
          }
        }
      `}</style>
    </>
  );
};

export const edgeTypes = {
  default: DefaultEdge,
  active: ActiveEdge,
  tension: TensionEdge,
  success: SuccessEdge,
};

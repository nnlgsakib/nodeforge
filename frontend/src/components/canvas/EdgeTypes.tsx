/* eslint-disable react-refresh/only-export-components */
import React, { useState, useRef, useCallback, useEffect } from 'react';
import {
  BaseEdge,
  getSmoothStepPath,
  type EdgeProps,
} from '@xyflow/react';
import { useTheme } from '../../hooks/use-theme';
import {
  type AppEdgeData,
  LONG_PRESS_DURATION,
} from '../../types/edges';

interface TypedEdgeProps extends EdgeProps {
  data?: AppEdgeData;
}

interface EdgeTooltipProps {
  x: number;
  y: number;
  data: AppEdgeData;
}

// Check if high-contrast mode is active (Subtask 1.7) — replaced by useTheme hook in components

// High-contrast edge colors (AC:1)
const HC_COLORS = {
  default: '#cccccc',
  active: '#00ffff',
  tension: '#ff0000',
  success: '#00ff00',
};

// Tooltip showing edge metadata on hover
const EdgeTooltip: React.FC<EdgeTooltipProps> = ({ x, y, data }) => {
  const { metadata, tension } = data;
  if (!metadata && tension === undefined) return null;

  return (
    <div
      role="tooltip"
      aria-label="Edge metadata"
      style={{
        position: 'absolute',
        left: x,
        top: y,
        background: 'var(--bg-secondary)',
        border: '1px solid var(--bg-tertiary)',
        borderRadius: '6px',
        padding: '8px 12px',
        fontSize: '11px',
        color: 'var(--text-primary)',
        zIndex: 100,
        pointerEvents: 'none',
        whiteSpace: 'nowrap',
        boxShadow: '0 2px 8px rgba(0,0,0,0.4)',
      }}
    >
      {tension !== undefined && <div>Tension: {(tension * 100).toFixed(0)}%</div>}
      {metadata?.latency !== undefined && <div>Latency: {metadata.latency}ms</div>}
      {metadata?.dataFlowRate !== undefined && <div>Flow: {metadata.dataFlowRate} tok/s</div>}
      {metadata?.message && <div style={{ marginTop: '4px', fontStyle: 'italic' }}>{metadata.message}</div>}
    </div>
  );
};

// Metadata bubble shown on long-press (TouchDesigner-style)
const MetadataBubble: React.FC<EdgeTooltipProps> = ({ x, y, data }) => {
  const { metadata, tension, upstreamHealth } = data as AppEdgeData & { upstreamHealth?: number };
  // Early return if there's nothing to show (P9)
  if (
    tension === undefined &&
    metadata?.latency === undefined &&
    metadata?.dataFlowRate === undefined &&
    upstreamHealth === undefined &&
    !metadata?.message
  ) {
    return null;
  }

  return (
    <div
      role="dialog"
      aria-label="Edge metadata bubble"
      style={{
        position: 'absolute',
        left: x,
        top: y,
        background: 'rgba(37, 38, 43, 0.95)',
        border: '1px solid var(--accent)',
        borderRadius: '8px',
        padding: '12px 16px',
        fontSize: '12px',
        color: 'var(--text-primary)',
        zIndex: 200,
        maxWidth: '220px',
        boxShadow: '0 4px 16px rgba(6, 182, 212, 0.3)',
      }}
    >
      <div style={{ fontWeight: 600, marginBottom: '6px', color: 'var(--accent)' }}>Edge Details</div>
      {tension !== undefined && (
        <div style={{ marginBottom: '4px' }}>
          <span style={{ color: 'var(--text-secondary)' }}>Tension: </span>
          <span>{`${(tension * 100).toFixed(0)}%`}</span>
        </div>
      )}
      {metadata?.latency !== undefined && (
        <div style={{ marginBottom: '4px' }}>
          <span style={{ color: 'var(--text-secondary)' }}>Latency: </span>
          {metadata.latency}ms
        </div>
      )}
      {metadata?.dataFlowRate !== undefined && (
        <div style={{ marginBottom: '4px' }}>
          <span style={{ color: 'var(--text-secondary)' }}>Flow Rate: </span>
          {metadata.dataFlowRate} tok/s
        </div>
      )}
      {metadata?.upstreamHealth !== undefined && (
        <div style={{ marginBottom: '4px' }}>
          <span style={{ color: 'var(--text-secondary)' }}>Upstream Health: </span>
          <span>{`${(metadata.upstreamHealth * 100).toFixed(0)}%`}</span>
        </div>
      )}
      {metadata?.message && (
        <div style={{ marginTop: '8px', paddingTop: '8px', borderTop: '1px solid var(--bg-tertiary)', fontStyle: 'italic' }}>
          {metadata.message}
        </div>
      )}
    </div>
  );
};

// Shared hook for edge tooltip/bubble interaction (P12)
function useEdgeInteraction() {
  const [showTooltip, setShowTooltip] = useState(false);
  const [showBubble, setShowBubble] = useState(false);
  const [tooltipPos, setTooltipPos] = useState({ x: 0, y: 0 });
  const [tooltipId] = useState(() => `edge-tooltip-${Math.random().toString(36).slice(2, 9)}`);
  const pressTimer = useRef<ReturnType<typeof setTimeout> | null>(null);

  const handleMouseEnter = useCallback((e: React.MouseEvent) => {
    setShowTooltip(true);
    setTooltipPos({ x: e.clientX, y: e.clientY - 40 });
  }, []);

  const handleMouseMove = useCallback((e: React.MouseEvent) => {
    if (showTooltip) {
      setTooltipPos({ x: e.clientX, y: e.clientY - 40 });
    }
  }, [showTooltip]);

  const handleMouseLeave = useCallback(() => {
    setShowTooltip(false);
    if (pressTimer.current) {
      clearTimeout(pressTimer.current);
      pressTimer.current = null;
    }
  }, []);

  const handlePointerDown = useCallback(() => {
    pressTimer.current = setTimeout(() => {
      setShowBubble(true);
    }, LONG_PRESS_DURATION);
  }, []);

  const handlePointerUp = useCallback((e: React.PointerEvent) => {
    if (pressTimer.current) {
      clearTimeout(pressTimer.current);
      pressTimer.current = null;
    }
    e.stopPropagation(); // Prevent immediate click dismissal (P6)
  }, []);

  // Cleanup timer on unmount (P5)
  useEffect(() => {
    return () => {
      if (pressTimer.current) {
        clearTimeout(pressTimer.current);
      }
    };
  }, []);

  // Dismiss bubble on outside click with deferred timing to avoid race (P6)
  useEffect(() => {
    if (showBubble) {
      const dismiss = () => setShowBubble(false);
      const timer = setTimeout(() => {
        window.addEventListener('click', dismiss, { once: true });
      }, 100);
      return () => {
        clearTimeout(timer);
        window.removeEventListener('click', dismiss);
      };
    }
  }, [showBubble]);

  return {
    showTooltip,
    showBubble,
    tooltipPos,
    tooltipId,
    setTooltipPos,
    handleMouseEnter,
    handleMouseMove,
    handleMouseLeave,
    handlePointerDown,
    handlePointerUp,
  };
}

// Base edge wrapper that handles interaction (P12, P13)
interface EdgeWrapperProps {
  id: string;
  ariaLabel: string;
  children: React.ReactNode;
  showTooltip: boolean;
  showBubble: boolean;
  tooltipPos: { x: number; y: number };
  tooltipId: string;
  data: AppEdgeData | undefined;
  handlers: {
    handleMouseEnter: (e: React.MouseEvent) => void;
    handleMouseMove: (e: React.MouseEvent) => void;
    handleMouseLeave: () => void;
    handlePointerDown: () => void;
    handlePointerUp: (e: React.PointerEvent) => void;
  };
}

const EdgeWrapper: React.FC<EdgeWrapperProps> = ({
  id: _id,
  ariaLabel,
  children,
  showTooltip,
  showBubble,
  tooltipPos,
  tooltipId,
  data,
  handlers,
}) => (
  <>
    <g
      onMouseEnter={handlers.handleMouseEnter}
      onMouseMove={handlers.handleMouseMove}
      onMouseLeave={handlers.handleMouseLeave}
      onPointerDown={handlers.handlePointerDown}
      onPointerUp={handlers.handlePointerUp}
      role="graphics-symbol"
      aria-label={ariaLabel}
      aria-describedby={showTooltip ? tooltipId : undefined}
      tabIndex={0}
    >
      {children}
    </g>
    {showTooltip && (
      <div id={tooltipId} style={{ position: 'fixed' }}>
        <EdgeTooltip x={tooltipPos.x} y={tooltipPos.y} data={data || {}} />
      </div>
    )}
    {showBubble && <MetadataBubble x={tooltipPos.x} y={tooltipPos.y} data={data || {}} />}
  </>
);

// Default Edge - #94a3b8, 2px stroke, no animation
const DefaultEdgeComponent: React.FC<TypedEdgeProps> = (props) => {
  const { id, source, target, sourceX, sourceY, targetX, targetY, sourcePosition, targetPosition, style = {}, data, selected } = props;
  const interaction = useEdgeInteraction();
  const { isHighContrast: hc } = useTheme();

  const [edgePath] = getSmoothStepPath({ sourceX, sourceY, sourcePosition, targetX, targetY, targetPosition });

  return (
    <EdgeWrapper
      id={id}
      ariaLabel={`Edge from ${source} to ${target}, status: default`}
      showTooltip={interaction.showTooltip}
      showBubble={interaction.showBubble}
      tooltipPos={interaction.tooltipPos}
      tooltipId={interaction.tooltipId}
      data={data}
      handlers={interaction}
    >
      <BaseEdge
        id={id}
        path={edgePath}
        style={{ stroke: hc ? HC_COLORS.default : '#94a3b8', strokeWidth: selected ? 3 : 2, ...style }}
      />
    </EdgeWrapper>
  );
};

// Active Edge - #06b6d4, 3px stroke, animated dash flow
const ActiveEdgeComponent: React.FC<TypedEdgeProps> = (props) => {
  const { id, source, target, sourceX, sourceY, targetX, targetY, sourcePosition, targetPosition, style = {}, data, selected } = props;
  const interaction = useEdgeInteraction();
  const { isHighContrast: hc } = useTheme();

  const [edgePath] = getSmoothStepPath({ sourceX, sourceY, sourcePosition, targetX, targetY, targetPosition });

  return (
    <EdgeWrapper
      id={id}
      ariaLabel={`Edge from ${source} to ${target}, status: active (flowing)`}
      showTooltip={interaction.showTooltip}
      showBubble={interaction.showBubble}
      tooltipPos={interaction.tooltipPos}
      tooltipId={interaction.tooltipId}
      data={data}
      handlers={interaction}
    >
      <BaseEdge
        id={id}
        path={edgePath}
        style={{ stroke: hc ? HC_COLORS.active : '#06b6d4', strokeWidth: selected ? 4 : 3, strokeDasharray: '12 6', animation: 'flow 1s linear infinite', ...style }}
      />
      <style>{`@keyframes flow { to { stroke-dashoffset: -18; } }`}</style>
    </EdgeWrapper>
  );
};

// Tension Edge - #ef4444, 4px stroke
const TensionEdgeComponent: React.FC<TypedEdgeProps> = (props) => {
  const { id, source, target, sourceX, sourceY, targetX, targetY, sourcePosition, targetPosition, style = {}, data, selected } = props;
  const interaction = useEdgeInteraction();
  const { isHighContrast: hc } = useTheme();

  const [edgePath] = getSmoothStepPath({ sourceX, sourceY, sourcePosition, targetX, targetY, targetPosition });

  return (
    <EdgeWrapper
      id={id}
      ariaLabel={`Edge from ${source} to ${target}, status: tension (upstream failure)`}
      showTooltip={interaction.showTooltip}
      showBubble={interaction.showBubble}
      tooltipPos={interaction.tooltipPos}
      tooltipId={interaction.tooltipId}
      data={data}
      handlers={interaction}
    >
      <BaseEdge
        id={id}
        path={edgePath}
        style={{ stroke: hc ? HC_COLORS.tension : '#ef4444', strokeWidth: selected ? 5 : 4, strokeDasharray: selected ? '8 4' : '16 8', opacity: selected ? 1 : 0.7, transition: 'stroke-dasharray 0.3s ease, opacity 0.3s ease', ...style }}
      />
    </EdgeWrapper>
  );
};

// Success Edge - #22c55e, 2px stroke, brief pulse on completion
const SuccessEdgeComponent: React.FC<TypedEdgeProps> = (props) => {
  const { id, source, target, sourceX, sourceY, targetX, targetY, sourcePosition, targetPosition, style = {}, data, selected } = props;
  const interaction = useEdgeInteraction();
  const { isHighContrast: hc } = useTheme();

  const [edgePath] = getSmoothStepPath({ sourceX, sourceY, sourcePosition, targetX, targetY, targetPosition });

  return (
    <EdgeWrapper
      id={id}
      ariaLabel={`Edge from ${source} to ${target}, status: success (completed)`}
      showTooltip={interaction.showTooltip}
      showBubble={interaction.showBubble}
      tooltipPos={interaction.tooltipPos}
      tooltipId={interaction.tooltipId}
      data={data}
      handlers={interaction}
    >
      <BaseEdge
        id={id}
        path={edgePath}
        style={{ stroke: hc ? HC_COLORS.success : '#22c55e', strokeWidth: selected ? 3 : 2, animation: 'pulse-success 0.6s ease-out', ...style }}
      />
      <style>{`@keyframes pulse-success { 0% { stroke-width: 2; opacity: 1; } 50% { stroke-width: 4; opacity: 0.8; } 100% { stroke-width: 2; opacity: 1; } }`}</style>
    </EdgeWrapper>
  );
};

export const edgeTypes = {
  default: DefaultEdgeComponent,
  active: ActiveEdgeComponent,
  tension: TensionEdgeComponent,
  success: SuccessEdgeComponent,
};

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
  const rafRef = useRef<number | null>(null); // M15: throttle mouse-move updates

  const handleMouseEnter = useCallback((e: React.MouseEvent) => {
    setShowTooltip(true);
    setTooltipPos({ x: e.clientX, y: e.clientY - 40 });
  }, []);

  // M15: Throttle tooltip position updates with requestAnimationFrame
  const handleMouseMove = useCallback((e: React.MouseEvent) => {
    if (showTooltip) {
      if (rafRef.current !== null) {
        cancelAnimationFrame(rafRef.current);
      }
      const x = e.clientX;
      const y = e.clientY - 40;
      rafRef.current = requestAnimationFrame(() => {
        setTooltipPos({ x, y });
        rafRef.current = null;
      });
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
    if (pressTimer.current) {
      clearTimeout(pressTimer.current);
    }
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
      if (rafRef.current !== null) {
        cancelAnimationFrame(rafRef.current);
      }
    };
  }, []);

  // M12: Dismiss bubble on Escape key
  useEffect(() => {
    if (showBubble) {
      const handleEscape = (e: KeyboardEvent) => {
        if (e.key === 'Escape') {
          setShowBubble(false);
        }
      };
      window.addEventListener('keydown', handleEscape);
      return () => window.removeEventListener('keydown', handleEscape);
    }
  }, [showBubble]);

  // Dismiss bubble on outside click with deferred timing to avoid race (P6)
  useEffect(() => {
    if (showBubble) {
      let listenerAdded = false;
      const dismiss = () => {
        listenerAdded = false;
        setShowBubble(false);
      };
      const timer = setTimeout(() => {
        listenerAdded = true;
        window.addEventListener('click', dismiss, { once: true });
      }, 100);
      return () => {
        clearTimeout(timer);
        if (listenerAdded) {
          window.removeEventListener('click', dismiss);
        }
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

// Active Edge - #06b6d4, 3px stroke, heartbeat ECG double-pulse
const ActiveEdgeComponent: React.FC<TypedEdgeProps> = (props) => {
  const { id, source, target, sourceX, sourceY, targetX, targetY, sourcePosition, targetPosition, style = {}, data, selected } = props;
  const interaction = useEdgeInteraction();
  const { isHighContrast: hc } = useTheme();

  // Heartbeat: ECG double-pulse rate scales with tension (higher tension = faster pulse)
  const rawTension = (data as AppEdgeData & { tension?: number })?.tension ?? 0;
  const tension = Math.min(1, Math.max(0, rawTension));
  const heartbeatDuration = Math.max(0.5, 1.5 - tension * 1.0); // 1.5s at tension=0, 0.5s at tension=1

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
        style={{ stroke: hc ? HC_COLORS.active : '#06b6d4', strokeWidth: selected ? 4 : 3, strokeDasharray: '12 6', animation: `heartbeat ${heartbeatDuration}s ease-in-out infinite`, ...style }}
      />
    </EdgeWrapper>
  );
};

// Tension Edge - #ef4444, dynamic stroke-width based on tension
const TensionEdgeComponent: React.FC<TypedEdgeProps> = (props) => {
  const { id, source, target, sourceX, sourceY, targetX, targetY, sourcePosition, targetPosition, style = {}, data, selected } = props;
  const interaction = useEdgeInteraction();
  const { isHighContrast: hc } = useTheme();

  // Dynamic stroke-width: scales from 3px to 6px based on tension
  const rawTension = (data as AppEdgeData & { tension?: number })?.tension ?? 0;
  const tension = Math.min(1, Math.max(0, rawTension));
  const dynamicStrokeWidth = 3 + (tension * 3); // 3px at tension=0, 6px at tension=1

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
        style={{ stroke: hc ? HC_COLORS.tension : '#ef4444', strokeWidth: selected ? dynamicStrokeWidth + 1 : dynamicStrokeWidth, strokeDasharray: selected ? '8 4' : '16 8', opacity: selected ? 1 : 0.7, transition: 'stroke-dasharray 0.3s ease, opacity 0.3s ease, stroke-width 0.3s ease', ...style }}
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
    </EdgeWrapper>
  );
};

export const edgeTypes = {
  default: DefaultEdgeComponent,
  active: ActiveEdgeComponent,
  tension: TensionEdgeComponent,
  success: SuccessEdgeComponent,
};

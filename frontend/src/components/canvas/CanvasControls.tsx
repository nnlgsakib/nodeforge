import React, { useState, useEffect, useCallback, useRef } from 'react';
import { MiniMap, Controls, useReactFlow, type Node } from '@xyflow/react';
import { useRtl } from '../../hooks/use-rtl';

interface CanvasControlsProps {
  nodes: Node[];
  edges: unknown[];
}

// Heat intensity calculation based on lastActiveAt timestamp
function getHeatIntensity(lastActiveAt?: number): number {
  if (!lastActiveAt) return 0;
  const now = Date.now();
  const ageMs = now - lastActiveAt;
  const fiveMinutes = 5 * 60 * 1000;
  const thirtyMinutes = 30 * 60 * 1000;

  if (ageMs <= fiveMinutes) return 1; // full glow
  if (ageMs <= thirtyMinutes) return 1 - (ageMs - fiveMinutes) / (thirtyMinutes - fiveMinutes); // dimming
  return 0; // no glow
}

// Calculate node color for MiniMap with heat effect
function getMiniMapNodeColor(node: Node): string {
  const status = (node.data as { status?: string })?.status;
  const lastActiveAt = (node.data as { lastActiveAt?: number })?.lastActiveAt;
  const heat = getHeatIntensity(lastActiveAt);

  let baseColor: string;
  if (status === 'running') baseColor = '#06b6d4';
  else if (status === 'complete') baseColor = '#22c55e';
  else if (status === 'failed') baseColor = '#ef4444';
  else baseColor = '#94a3b8';

  // Apply heat glow by blending with white and adding brightness — only for valid 7-char hex colors
  if (heat > 0 && /^#[0-9a-fA-F]{6}$/.test(baseColor)) {
    const r = parseInt(baseColor.slice(1, 3), 16);
    const g = parseInt(baseColor.slice(3, 5), 16);
    const b = parseInt(baseColor.slice(5, 7), 16);
    const glowR = Math.round(r + (255 - r) * heat * 0.5);
    const glowG = Math.round(g + (255 - g) * heat * 0.5);
    const glowB = Math.round(b + (255 - b) * heat * 0.5);
    return `#${glowR.toString(16).padStart(2, '0')}${glowG.toString(16).padStart(2, '0')}${glowB.toString(16).padStart(2, '0')}`;
  }

  return baseColor;
}

// Calculate stroke color for MiniMap glow effect — hot nodes get a bright ring
function getMiniMapNodeStrokeColor(node: Node): string {
  const lastActiveAt = (node.data as { lastActiveAt?: number })?.lastActiveAt;
  const heat = getHeatIntensity(lastActiveAt);
  if (heat > 0.7) return '#fbbf24'; // bright amber for very hot
  if (heat > 0.3) return '#fcd34d'; // soft amber for warm
  return 'var(--bg-tertiary)'; // default
}

export const CanvasControls: React.FC<CanvasControlsProps> = ({ nodes, edges }) => {
  const [showKeybindings, setShowKeybindings] = useState(false);
  const { setViewport, getViewport } = useReactFlow();
  const canvasRef = useRef<HTMLDivElement>(null);
  const isRtl = useRtl();

  const nodeCount = nodes.length;
  const edgeCount = edges.length;
  const activeNodes = nodes.filter(
    (n) => (n.data as { status?: string })?.status === 'running'
  ).length;

  // Vim/Emacs keyboard navigation
  const handleKeyDown = useCallback((e: KeyboardEvent) => {
    // Skip if user is typing in an input/textarea
    const target = e.target as HTMLElement;
    if (target.tagName === 'INPUT' || target.tagName === 'TEXTAREA' || target.isContentEditable) {
      return;
    }

    const panStep = 50;

    // Vim keys: h=left, j=down, k=up, l=right
    // Only trigger on bare keys — skip if any modifier is held to avoid hijacking browser shortcuts
    if (!e.ctrlKey && !e.metaKey && !e.altKey) {
      const viewport = getViewport();
      switch (e.key) {
        case 'h':
          e.preventDefault();
          setViewport({ ...viewport, x: viewport.x - panStep, zoom: viewport.zoom });
          break;
        case 'j':
          e.preventDefault();
          setViewport({ ...viewport, y: viewport.y - panStep, zoom: viewport.zoom });
          break;
        case 'k':
          e.preventDefault();
          setViewport({ ...viewport, y: viewport.y + panStep, zoom: viewport.zoom });
          break;
        case 'l':
          e.preventDefault();
          setViewport({ ...viewport, x: viewport.x + panStep, zoom: viewport.zoom });
          break;
      }
      return;
    }

    // Emacs keys: Ctrl-f=forward/right, Ctrl-b=back/left, Ctrl-n=next/down, Ctrl-p=previous/up
    if (!e.ctrlKey) return;

    const viewport = getViewport();
    switch (e.key) {
      case 'f':
        e.preventDefault();
        setViewport({ ...viewport, x: viewport.x + panStep, zoom: viewport.zoom });
        break;
      case 'b':
        e.preventDefault();
        setViewport({ ...viewport, x: viewport.x - panStep, zoom: viewport.zoom });
        break;
      case 'n':
        e.preventDefault();
        setViewport({ ...viewport, y: viewport.y - panStep, zoom: viewport.zoom });
        break;
      case 'p':
        e.preventDefault();
        setViewport({ ...viewport, y: viewport.y + panStep, zoom: viewport.zoom });
        break;
    }
  }, [setViewport, getViewport]);

  // Attach keyboard listener to canvas
  useEffect(() => {
    let target: Element | Document | null = null;

    const attach = () => {
      const canvas = canvasRef.current?.closest('.react-flow') || document.querySelector('.react-flow');
      if (!canvas) return false;
      target = canvas as Element;
      canvas.addEventListener('keydown', handleKeyDown as EventListener);
      return true;
    };

    if (!attach()) {
      // Retry once after a short delay in case ReactFlow hasn't rendered yet
      const timer = setTimeout(() => {
        if (!attach()) {
          // Final fallback: attach to document
          target = document;
          document.addEventListener('keydown', handleKeyDown as EventListener);
        }
      }, 100);
      return () => {
        clearTimeout(timer);
        if (target) target.removeEventListener('keydown', handleKeyDown as EventListener);
      };
    }

    return () => {
      if (target) target.removeEventListener('keydown', handleKeyDown as EventListener);
    };
  }, [handleKeyDown]);

  return (
    <div ref={canvasRef} className="canvas-controls" style={{
      display: 'flex',
      flexDirection: 'column',
      gap: '8px',
      position: 'absolute',
      bottom: '16px',
      left: isRtl ? 'auto' : '16px',
      right: isRtl ? '16px' : 'auto',
      zIndex: 10,
    }}>
      {/* MiniMap with heat visualization (Subtask 2.3) */}
      <div style={{
        background: 'var(--bg-secondary)',
        borderRadius: '8px',
        overflow: 'hidden',
        border: '1px solid var(--bg-tertiary)',
      }}>
        <MiniMap
          nodeColor={getMiniMapNodeColor}
          nodeStrokeColor={getMiniMapNodeStrokeColor}
          nodeStrokeWidth={2}
          maskColor="rgba(15, 23, 42, 0.7)"
          pannable
          zoomable
          style={{ width: 200, height: 150 }}
        />
      </div>

      {/* Styled Controls */}
      <div style={{
        display: 'flex',
        flexDirection: 'column',
        gap: '4px',
        background: 'var(--bg-secondary)',
        border: '1px solid var(--bg-tertiary)',
        borderRadius: '8px',
        padding: '4px',
      }}>
        <Controls
          showInteractive={false}
          style={{
            background: 'transparent',
            border: 'none',
            boxShadow: 'none',
          }}
        />
      </div>

      {/* Zoom/Pan indicator */}
      <div style={{
        background: 'var(--bg-secondary)',
        border: '1px solid var(--bg-tertiary)',
        borderRadius: '6px',
        padding: '6px 10px',
        fontSize: '12px',
        color: 'var(--text-secondary)',
        fontFamily: 'JetBrains Mono, monospace',
      }}>
        Nodes: {nodeCount} | Edges: {edgeCount} | Active: {activeNodes}
      </div>

      {/* Keybinding hints toggle */}
      <button
        className="collapse-btn"
        onClick={() => setShowKeybindings(!showKeybindings)}
        title="Toggle keybinding hints"
        style={{
          background: 'var(--bg-secondary)',
          border: '1px solid var(--bg-tertiary)',
          borderRadius: '6px',
          padding: '6px 10px',
          color: 'var(--text-secondary)',
          cursor: 'pointer',
          fontSize: '12px',
          transition: 'color 200ms, background-color 200ms',
        }}
        onMouseEnter={(e) => {
          e.currentTarget.style.color = 'var(--text-primary)';
          e.currentTarget.style.backgroundColor = 'var(--bg-tertiary)';
        }}
        onMouseLeave={(e) => {
          e.currentTarget.style.color = 'var(--text-secondary)';
          e.currentTarget.style.backgroundColor = 'var(--bg-secondary)';
        }}
      >
        Keys
      </button>

      {/* Keybinding hints panel (Subtask 4.6) */}
      {showKeybindings && (
        <div
          style={{
            background: 'var(--bg-secondary)',
            border: '1px solid var(--bg-tertiary)',
            borderRadius: '8px',
            padding: '12px',
            fontSize: '12px',
            color: 'var(--text-primary)',
            fontFamily: 'JetBrains Mono, monospace',
            minWidth: '220px',
          }}
          role="region"
          aria-label="Canvas controls, Vim keys: hjkl"
        >
          <h4 style={{ margin: '0 0 8px 0', fontSize: '13px', fontWeight: 600 }}>Canvas Navigation</h4>
          <div style={{ display: 'flex', flexDirection: 'column', gap: '6px' }}>
            <div>
              <strong style={{ color: 'var(--accent)' }}>Vim</strong>
              <div style={{ marginLeft: '8px', color: 'var(--text-secondary)' }}>
                <kbd style={{ padding: '1px 4px', background: 'var(--bg-tertiary)', borderRadius: '3px', fontSize: '11px' }}>h</kbd> left
                <kbd style={{ padding: '1px 4px', background: 'var(--bg-tertiary)', borderRadius: '3px', fontSize: '11px', marginLeft: '4px' }}>j</kbd> down
                <kbd style={{ padding: '1px 4px', background: 'var(--bg-tertiary)', borderRadius: '3px', fontSize: '11px', marginLeft: '4px' }}>k</kbd> up
                <kbd style={{ padding: '1px 4px', background: 'var(--bg-tertiary)', borderRadius: '3px', fontSize: '11px', marginLeft: '4px' }}>l</kbd> right
              </div>
            </div>
            <div>
              <strong style={{ color: 'var(--accent)' }}>Emacs</strong>
              <div style={{ marginLeft: '8px', color: 'var(--text-secondary)' }}>
                <kbd style={{ padding: '1px 4px', background: 'var(--bg-tertiary)', borderRadius: '3px', fontSize: '11px' }}>Ctrl-f</kbd> forward
                <kbd style={{ padding: '1px 4px', background: 'var(--bg-tertiary)', borderRadius: '3px', fontSize: '11px', marginLeft: '4px' }}>Ctrl-b</kbd> back
                <kbd style={{ padding: '1px 4px', background: 'var(--bg-tertiary)', borderRadius: '3px', fontSize: '11px', marginLeft: '4px' }}>Ctrl-n</kbd> next
                <kbd style={{ padding: '1px 4px', background: 'var(--bg-tertiary)', borderRadius: '3px', fontSize: '11px', marginLeft: '4px' }}>Ctrl-p</kbd> prev
              </div>
            </div>
          </div>
          <h4 style={{ margin: '12px 0 8px 0', fontSize: '13px', fontWeight: 600 }}>Node Controls</h4>
          <div style={{ display: 'flex', flexDirection: 'column', gap: '6px', color: 'var(--text-secondary)' }}>
            <div>
              <kbd style={{ padding: '1px 4px', background: 'var(--bg-tertiary)', borderRadius: '3px', fontSize: '11px' }}>p</kbd> / <kbd style={{ padding: '1px 4px', background: 'var(--bg-tertiary)', borderRadius: '3px', fontSize: '11px' }}>Space</kbd> pause/resume
            </div>
            <div>
              <kbd style={{ padding: '1px 4px', background: 'var(--bg-tertiary)', borderRadius: '3px', fontSize: '11px' }}>r</kbd> retry
              <kbd style={{ padding: '1px 4px', background: 'var(--bg-tertiary)', borderRadius: '3px', fontSize: '11px', marginLeft: '4px' }}>f</kbd> fork
              <kbd style={{ padding: '1px 4px', background: 'var(--bg-tertiary)', borderRadius: '3px', fontSize: '11px', marginLeft: '4px' }}>s</kbd> skip
            </div>
            <div>
              <kbd style={{ padding: '1px 4px', background: 'var(--bg-tertiary)', borderRadius: '3px', fontSize: '11px' }}>m</kbd> toggle monologue
              <kbd style={{ padding: '1px 4px', background: 'var(--bg-tertiary)', borderRadius: '3px', fontSize: '11px', marginLeft: '4px' }}>Esc</kbd> close panel
            </div>
            <div>
              <kbd style={{ padding: '1px 4px', background: 'var(--bg-tertiary)', borderRadius: '3px', fontSize: '11px' }}>Tab</kbd> cycle nodes
              <kbd style={{ padding: '1px 4px', background: 'var(--bg-tertiary)', borderRadius: '3px', fontSize: '11px', marginLeft: '4px' }}>Enter</kbd> open config
            </div>
          </div>
        </div>
      )}
    </div>
  );
};

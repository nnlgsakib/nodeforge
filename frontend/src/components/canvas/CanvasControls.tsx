import React, { useState } from 'react';
import { MiniMap, Controls } from '@xyflow/react';

interface CanvasControlsProps {
  nodes: { data?: { status?: string } }[];
  edges: unknown[];
}

export const CanvasControls: React.FC<CanvasControlsProps> = ({ nodes, edges }) => {
  const [showKeybindings, setShowKeybindings] = useState(false);

  const nodeCount = nodes.length;
  const edgeCount = edges.length;

  // Calculate activity heat (simplified - nodes with status glow)
  const activeNodes = nodes.filter(
    (n) => n.data?.status === 'running'
  ).length;

  return (
    <div className="canvas-controls">
      {/* MiniMap with heat visualization */}
      <MiniMap
        nodeColor={(node) => {
          const status = (node as { data?: { status?: string } }).data?.status;
          if (status === 'running') return '#06b6d4';
          if (status === 'complete') return '#22c55e';
          if (status === 'failed') return '#ef4444';
          return '#94a3b8';
        }}
        maskColor="rgba(26, 27, 30, 0.7)"
        style={{
          background: 'var(--bg-secondary)',
          borderRadius: '8px',
          overflow: 'hidden',
        }}
      />

      {/* Standard Controls */}
      <Controls
        style={{
          display: 'flex',
          flexDirection: 'column',
          gap: '4px',
          background: 'var(--bg-secondary)',
          border: '1px solid var(--bg-tertiary)',
          borderRadius: '8px',
          padding: '4px',
        }}
      />

      {/* Zoom/Pan indicator */}
      <div
        style={{
          background: 'var(--bg-secondary)',
          border: '1px solid var(--bg-tertiary)',
          borderRadius: '6px',
          padding: '6px 10px',
          fontSize: '12px',
          color: 'var(--text-secondary)',
        }}
      >
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
        }}
      >
        Keys
      </button>

      {/* Keybinding hints panel */}
      {showKeybindings && (
        <div
          style={{
            position: 'absolute',
            bottom: '16px',
            right: '16px',
            background: 'var(--bg-secondary)',
            border: '1px solid var(--bg-tertiary)',
            borderRadius: '8px',
            padding: '12px',
            fontSize: '12px',
            color: 'var(--text-primary)',
            zIndex: 10,
            minWidth: '200px',
          }}
        >
          <h4 style={{ margin: '0 0 8px 0', fontSize: '13px' }}>Keybindings</h4>
          <div style={{ display: 'flex', flexDirection: 'column', gap: '4px' }}>
            <div><kbd>p</kbd> / <kbd>space</kbd> - Pause/resume</div>
            <div><kbd>s</kbd> - Skip node</div>
            <div><kbd>f</kbd> - Fork session</div>
            <div><kbd>r</kbd> - Retry failed node</div>
            <div><kbd>m</kbd> - Toggle MonologuePanel</div>
            <hr style={{ border: 'none', borderTop: '1px solid var(--bg-tertiary)', margin: '6px 0' }} />
            <div><kbd>h</kbd> / <kbd>j</kbd> / <kbd>k</kbd> / <kbd>l</kbd> - Vim nav</div>
            <div><kbd>Ctrl-f</kbd> / <kbd>Ctrl-b</kbd> - Emacs nav</div>
          </div>
        </div>
      )}
    </div>
  );
};

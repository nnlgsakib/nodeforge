import { useState, useCallback, useEffect } from 'react';
import {
  ReactFlow,
  Background,
  BackgroundVariant,
  addEdge,
  useNodesState,
  useEdgesState,
  type OnConnect,
} from '@xyflow/react';

import '@xyflow/react/dist/style.css';

import { initialNodes, nodeTypes } from './nodes';
import { initialEdges, edgeTypes } from './edges';
import { SessionExplorer } from './components/panels/SessionExplorer';
import { ChatPanel } from './components/panels/ChatPanel';
import { MonologuePanel } from './components/panels/monologue-panel';
import { CanvasControls } from './components/canvas/CanvasControls';
import { useWebSocket } from './hooks/useWebSocket';
import { useKeyboardShortcuts } from './hooks/useKeyboardShortcuts';
import { useLayoutWorker } from './hooks/useLayoutWorker';

interface ProjectResult {
  sessionId: string;
  projectName: string;
  workspace: string;
}

async function createProject(projectName: string): Promise<ProjectResult> {
  const response = await fetch('/api/v1/sessions', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ projectName }),
  });
  if (!response.ok) {
    const contentType = response.headers.get('content-type');
    if (contentType && contentType.includes('application/json')) {
      const error = await response.json();
      throw new Error(error.error || 'Failed to create project');
    }
    throw new Error(`Failed to create project: ${response.status} ${response.statusText}`);
  }
  return response.json();
}

// Phase band colors for canvas top (Task 3.3)
const phaseBands = [
  { label: 'Discovery', color: '#3B82F6', x: 0, width: 25 },
  { label: 'Execution', color: '#F97316', x: 25, width: 25 },
  { label: 'Recovery', color: '#EF4444', x: 50, width: 25 },
  { label: 'Completion', color: '#22C55E', x: 75, width: 25 },
];

export default function App() {
  const [nodes, setNodes, onNodesChange] = useNodesState(initialNodes);
  const [edges, setEdges, onEdgesChange] = useEdgesState(initialEdges);
  const [chatCollapsed, setChatCollapsed] = useState(false);
  const [monologueCollapsed, setMonologueCollapsed] = useState(false);
  const [chatGenerating, setChatGenerating] = useState(false);
  const [isPaused, setIsPaused] = useState(false);
  const [sessionId, setSessionId] = useState<string | undefined>();
  const [notification, setNotification] = useState<{ type: 'success' | 'error'; message: string } | null>(null);

  // WebSocket connection
  const {
    connected,
    monologueMessages,
    isStreaming,
    sendMessage,
    graphUpdateQueue,
    nodeUpdateQueue,
    edgeUpdateQueue,
    clearGraphUpdates,
    clearNodeUpdates,
    clearEdgeUpdates,
    clearMonologueMessages,
  } = useWebSocket();

  // Web Worker layout offloading (Story 2.7)
  const { runLayout } = useLayoutWorker();

  // Offload layout to Web Worker when graph updates arrive (Story 2.7, AC2-3)
  useEffect(() => {
    if (graphUpdateQueue.length === 0) return;

    // Collect the last update that contains nodes or edges
    let lastNodes: Array<{ id: string; [key: string]: unknown }> | null = null;
    let lastEdges: Array<{ source: string; target: string; [key: string]: unknown }> | null = null;

    for (const item of graphUpdateQueue) {
      const data = item as { nodes?: unknown[]; edges?: unknown[] };
      if (data.nodes && data.nodes.length > 0) {
        lastNodes = data.nodes as Array<{ id: string }>;
      }
      if (data.edges && data.edges.length > 0) {
        lastEdges = data.edges as Array<{ source: string; target: string }>;
      }
    }

    if (!lastNodes && !lastEdges) {
      clearGraphUpdates();
      return;
    }

    // If we have nodes, run layout; otherwise just update edges
    if (lastNodes) {
      runLayout(lastNodes, lastEdges || [])
        .then((positions) => {
          requestAnimationFrame(() => {
            // Replace nodes entirely (preserving positions from layout)
            setNodes(
              lastNodes!.map((node) => ({
                ...node,
                position: positions[node.id] || node.position || { x: 0, y: 0 },
              })) as any[]
            );
            if (lastEdges && lastEdges.length > 0) {
              setEdges(lastEdges as any[]);
            }
            setChatGenerating(false);
          });
        })
        .catch((err) => {
          console.warn('Layout worker failed, falling back to no-layout:', err);
          // Fallback: set nodes and edges directly without layout
          requestAnimationFrame(() => {
            setNodes(lastNodes! as any[]);
            if (lastEdges && lastEdges.length > 0) {
              setEdges(lastEdges as any[]);
            }
            setChatGenerating(false);
          });
        });
    } else if (lastEdges) {
      // Edge-only update
      setEdges(lastEdges as any[]);
      setChatGenerating(false);
    }

    clearGraphUpdates();
  }, [graphUpdateQueue, runLayout, clearGraphUpdates]);

  // Handle node_update messages (Task 5.4, AC2)
  useEffect(() => {
    if (nodeUpdateQueue.length === 0) return;
    for (const item of nodeUpdateQueue) {
      const data = item as { nodeId?: string; status?: string; progress?: number };
      if (data.nodeId) {
        setNodes((nds) =>
          nds.map((node) => {
            if (node.id === data.nodeId) {
              return {
                ...node,
                data: {
                  ...node.data,
                  status: data.status || node.data?.status,
                  progress: data.progress ?? node.data?.progress,
                },
              };
            }
            return node;
          }) as any[]
        );
      }
    }
    clearNodeUpdates();
  }, [nodeUpdateQueue, setNodes, clearNodeUpdates]);

  // Handle edge_update messages (Task 6.2)
  useEffect(() => {
    if (edgeUpdateQueue.length === 0) return;
    for (const item of edgeUpdateQueue) {
      const data = item as { source?: string; target?: string; tension?: number };
      if (data.source && data.target) {
        setEdges((eds) =>
          eds.map((edge) => {
            if (edge.source === data.source && edge.target === data.target) {
              const tension = data.tension || 0;
              let edgeType = 'default';
              if (tension > 0.7) edgeType = 'tension';
              else if (tension > 0.3) edgeType = 'active';
              else if (tension === 0) edgeType = 'success';
              return { ...edge, type: edgeType, data: { ...edge.data, tension } };
            }
            return edge;
          }) as any[]
        );
      }
    }
    clearEdgeUpdates();
  }, [edgeUpdateQueue, setEdges, clearEdgeUpdates]);

  // Keyboard shortcuts via dedicated hook
  useKeyboardShortcuts({
    onToggleMonologue: () => setMonologueCollapsed((prev) => !prev),
    onTogglePause: (paused) => setIsPaused(paused),
    isPaused,
    onSkipNode: () => console.log('Skip node triggered'),
    onForkSession: () => console.log('Fork session triggered'),
    onRetryNode: () => console.log('Retry failed node triggered'),
    sendMessage,
  });

  const onConnect: OnConnect = useCallback(
    (connection) => setEdges((edges) => addEdge(connection, edges)),
    [setEdges]
  );

  const handleCreateProject = async (projectName: string) => {
    try {
      const result = await createProject(projectName);
      setSessionId(result.sessionId);
      console.log('Project created:', result);
      setNotification({ type: 'success', message: `Project "${projectName}" created! Session: ${result.sessionId}` });
    } catch (err) {
      console.error('Failed to create project:', err);
      setNotification({ type: 'error', message: err instanceof Error ? err.message : String(err) });
    }
    setTimeout(() => setNotification(null), 5000);
  };

  const handleSendGoal = useCallback(
    (text: string) => {
      if (!connected) {
        setNotification({ type: 'error', message: 'WebSocket not connected. Please wait or refresh.' });
        return;
      }
      setChatGenerating(true);
      sendMessage({ type: 'goal', text });
    },
    [connected, sendMessage]
  );

  return (
    <div className="app-container">
      {notification && (
        <div className={`notification notification-${notification.type}`}>
          {notification.message}
          <button className="notification-close" onClick={() => setNotification(null)}>×</button>
        </div>
      )}

      {/* Connection status indicator */}
      <div
        style={{
          position: 'absolute',
          top: '30px',
          right: '10px',
          zIndex: 10,
          display: 'flex',
          alignItems: 'center',
          gap: '4px',
          fontSize: '10px',
          color: 'var(--text-secondary)',
        }}
      >
        <div
          style={{
            width: '8px',
            height: '8px',
            borderRadius: '50%',
            background: connected ? '#22c55e' : '#ef4444',
          }}
        />
        {connected ? 'Connected' : 'Disconnected'}
      </div>

      {/* Phase bands at top of canvas (Task 3.3) */}
      <div
        style={{
          position: 'absolute',
          top: 0,
          left: 0,
          right: 0,
          height: '24px',
          display: 'flex',
          zIndex: 5,
          background: 'var(--bg-primary)',
          borderBottom: '1px solid var(--bg-tertiary)',
        }}
      >
        {phaseBands.map((band) => (
          <div
            key={band.label}
            style={{
              flex: `0 0 ${band.width}%`,
              background: band.color,
              opacity: 0.2,
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              fontSize: '10px',
              color: 'var(--text-secondary)',
              fontWeight: 500,
            }}
            title={band.label}
          >
            {band.label}
          </div>
        ))}
      </div>

      <div className="sidebar">
        <SessionExplorer onCreateProject={handleCreateProject} />
      </div>

      <div className="main-content" style={{ marginTop: '24px' }}>
        <ReactFlow
          nodes={nodes}
          nodeTypes={nodeTypes}
          onNodesChange={onNodesChange}
          edges={edges}
          edgeTypes={edgeTypes}
          onEdgesChange={onEdgesChange}
          onConnect={onConnect}
          nodesDraggable={true}
          nodesConnectable={true}
        >
          <Background variant={BackgroundVariant.Dots} gap={12} size={1} color="#334155" />
          <CanvasControls nodes={nodes} edges={edges} />
        </ReactFlow>
      </div>

      <ChatPanel
        collapsed={chatCollapsed}
        onToggleCollapse={() => setChatCollapsed((prev) => !prev)}
        onSendGoal={handleSendGoal}
        generating={chatGenerating}
      />

      <MonologuePanel
        collapsed={monologueCollapsed}
        onToggleCollapse={() => setMonologueCollapsed((prev) => !prev)}
        messages={monologueMessages}
        isStreaming={isStreaming}
        onClear={clearMonologueMessages}
        sessionId={sessionId}
      />
    </div>
  );
}

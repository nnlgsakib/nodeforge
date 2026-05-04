import { useState, useCallback, useEffect, useRef } from 'react';
import {
  ReactFlow,
  Background,
  BackgroundVariant,
  addEdge,
  useNodesState,
  useEdgesState,
  type OnConnect,
  useReactFlow,
} from '@xyflow/react';

import '@xyflow/react/dist/style.css';

import { initialNodes, nodeTypes } from './nodes';
import { initialEdges, edgeTypes } from './edges';
import { SessionExplorer } from './components/panels/SessionExplorer';
import { ChatPanel } from './components/panels/ChatPanel';
import { MonologuePanel } from './components/panels/monologue-panel';
import { SkillMarketplace } from './components/panels/skill-marketplace';
import { AccessibilityToolbar } from './components/ui/AccessibilityToolbar';
import { CanvasControls } from './components/canvas/CanvasControls';
import { PhaseBands } from './components/canvas/PhaseBands';
import { NodeConfig } from './components/panels/node-config';
import { useWebSocket } from './hooks/useWebSocket';
import { useKeyboardShortcuts } from './hooks/useKeyboardShortcuts';
import { useLayoutWorker } from './hooks/useLayoutWorker';

// P8: Hoisted outside component to avoid recreation every render
const FILE_TO_NODE_TYPE: Record<string, string> = {
  'go.mod': 'implement',
  'go.sum': 'implement',
  '.go': 'implement',
  'spec.md': 'spec',
  'plan.md': 'plan',
  'test.go': 'test',
  '_test.go': 'test',
  'review.md': 'review',
  'README.md': 'goal',
};

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

export default function App() {
  const [nodes, setNodes, onNodesChange] = useNodesState(initialNodes);
  const [edges, setEdges, onEdgesChange] = useEdgesState(initialEdges);
  const [chatCollapsed, setChatCollapsed] = useState(false);
  const [monologueCollapsed, setMonologueCollapsed] = useState(false);
  const [chatGenerating, setChatGenerating] = useState(false);
  const [isPaused, setIsPaused] = useState(false);
  const [sessionId, setSessionId] = useState<string | undefined>();
  const [notification, setNotification] = useState<{ type: 'success' | 'error'; message: string } | null>(null);
  const [nodeConfigOpen, setNodeConfigOpen] = useState(false);
  const [nodeConfigNodeId, setNodeConfigNodeId] = useState<string | null>(null);
  const [selectedNodeId, setSelectedNodeId] = useState<string | null>(null);
  const [accessibilityVisible, setAccessibilityVisible] = useState(false);
  const [marketplaceOpen, setMarketplaceOpen] = useState(false);
  const [isRtl, setIsRtl] = useState(false);

  // Track previous node statuses to only announce changes in ARIA live region
  const prevStatusesRef = useRef<Record<string, string>>({});
  const [statusAnnouncements, setStatusAnnouncements] = useState<{ id: string; label: string; status: string }[]>([]);

  // Timer ref for notification cleanup
  const notificationTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null);

  // Cleanup timer on unmount
  useEffect(() => {
    return () => {
      if (notificationTimerRef.current) {
        clearTimeout(notificationTimerRef.current);
      }
    };
  }, []);

  // Detect node status changes and announce only changed nodes (WCAG 2.1 AA, Subtask 2.1)
  useEffect(() => {
    const currentStatuses: Record<string, string> = {};
    const changed: { id: string; label: string; status: string }[] = [];

    for (const node of nodes) {
      const status = (node as any).data?.status;
      if (status) {
        currentStatuses[node.id] = status;
        if (prevStatusesRef.current[node.id] !== status) {
          const label = String((node as any).data?.label ?? node.id);
          changed.push({ id: node.id, label, status });
        }
      }
    }

    if (changed.length > 0) {
      setStatusAnnouncements(changed);
    }
    prevStatusesRef.current = currentStatuses;
  }, [nodes]);

  // Listen for RTL mode changes from AccessibilityToolbar
  useEffect(() => {
    const observer = new MutationObserver(() => {
      setIsRtl(document.documentElement.dir === 'rtl');
    });
    observer.observe(document.documentElement, { attributes: true, attributeFilter: ['dir'] });
    setIsRtl(document.documentElement.dir === 'rtl');
    return () => observer.disconnect();
  }, []);

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
    if (notificationTimerRef.current) clearTimeout(notificationTimerRef.current);
    notificationTimerRef.current = setTimeout(() => setNotification(null), 5000);
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

  // Node configuration handlers
  const handleNodeDoubleClick = useCallback((nodeId: string) => {
    setNodeConfigNodeId(nodeId);
    setNodeConfigOpen(true);
  }, []);

  // Open NodeConfig with Enter key on selected node
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      const target = e.target as HTMLElement;
      if (target.tagName === 'INPUT' || target.tagName === 'TEXTAREA' || target.isContentEditable) {
        return;
      }
      if (e.key === 'Enter' && selectedNodeId) {
        e.preventDefault();
        setNodeConfigNodeId(selectedNodeId);
        setNodeConfigOpen(true);
      }
    };
    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  }, [selectedNodeId]);

  const handleNodeConfigSave = useCallback((nodeId: string, config: { timeout: number; retryCount: number; tokenBudget: number }) => {
    // Send config via WebSocket message
    sendMessage({
      type: 'node_update',
      nodeId,
      config,
    });
    setNodeConfigOpen(false);
    setNotification({ type: 'success', message: `Node configuration saved` });
    if (notificationTimerRef.current) clearTimeout(notificationTimerRef.current);
    notificationTimerRef.current = setTimeout(() => setNotification(null), 3000);
  }, [sendMessage]);

  // Drag-and-drop file to node creation (AC4)
  const reactFlowWrapper = useRef<HTMLDivElement>(null);
  const [isDragging, setIsDragging] = useState(false);
  const isDraggingRef = useRef(false); // P7: ref guard to avoid re-render spam
  const { screenToFlowPosition } = useReactFlow(); // P4: proper canvas coords

  const handleDragOver = useCallback((e: React.DragEvent) => {
    e.preventDefault();
    e.dataTransfer.dropEffect = 'copy';
    if (!isDraggingRef.current) {
      isDraggingRef.current = true;
      setIsDragging(true);
    }
  }, []);

  const handleDragLeave = useCallback(() => {
    isDraggingRef.current = false;
    setIsDragging(false);
  }, []);

  const handleDrop = useCallback(
    (event: React.DragEvent) => {
      isDraggingRef.current = false;
      setIsDragging(false);
      const files = Array.from(event.dataTransfer.files);
      if (files.length === 0) return;

      const file = files[0];
      const fileName = file.name;
      const parts = fileName.split('.');
      const ext = parts.length > 1 ? '.' + parts[parts.length - 1] : undefined;

      // Determine node type from filename
      let nodeType = FILE_TO_NODE_TYPE[fileName] ?? (ext ? FILE_TO_NODE_TYPE[ext] : undefined) ?? 'implement';

      // P4: Use React Flow's screenToFlowPosition for correct canvas coordinates
      const position = screenToFlowPosition({
        x: event.clientX,
        y: event.clientY,
      });

      const newNode = {
        id: `node-${Date.now()}-${crypto.randomUUID?.().slice(0, 8) ?? Math.random().toString(36).slice(2, 10)}`,
        type: nodeType,
        position,
        data: { label: fileName },
      };

      setNodes((nds) => [...nds, newNode]);
      setNotification({
        type: 'success',
        message: `Created "${nodeType}" node from "${fileName}"`,
      });
      if (notificationTimerRef.current) clearTimeout(notificationTimerRef.current);
      notificationTimerRef.current = setTimeout(() => setNotification(null), 3000);
    },
    [setNodes, screenToFlowPosition]
  );

  return (
    <div className="app-container">
      {/* ARIA live region for node status changes (WCAG 2.1 AA, Subtask 2.1) */}
      <div aria-live="polite" aria-atomic="true" className="sr-only">
        {statusAnnouncements.map((announcement) => (
          <div key={`aria-${announcement.id}`} role="status">
            Node {announcement.label} changed to {announcement.status}
          </div>
        ))}
      </div>

      {/* ARIA live region for critical failures */}
      <div aria-live="assertive" aria-atomic="true" className="sr-only">
        {notification && notification.type === 'error' && (
          <div role="alert">{notification.message}</div>
        )}
      </div>

      {notification && (
        <div className={`notification notification-${notification.type}`}>
          {notification.message}
          <button className="notification-close" onClick={() => {
            if (notificationTimerRef.current) {
              clearTimeout(notificationTimerRef.current);
              notificationTimerRef.current = null;
            }
            setNotification(null);
          }}>×</button>
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

      {/* Drag overlay indicator */}
      {isDragging && (
        <div
          style={{
            position: 'absolute',
            top: 0,
            left: 0,
            right: 0,
            bottom: 0,
            zIndex: 1000,
            pointerEvents: 'none',
            border: '2px dashed var(--accent)',
            backgroundColor: 'rgba(6, 182, 212, 0.05)',
          }}
        />
      )}

      <div className="sidebar">
        <SessionExplorer onCreateProject={handleCreateProject} />
      </div>

      <div className="main-content" style={{ marginTop: '24px' }} ref={reactFlowWrapper}>
        <ReactFlow
          nodes={nodes}
          nodeTypes={nodeTypes}
          onNodesChange={onNodesChange}
          edges={edges}
          edgeTypes={edgeTypes}
          onEdgesChange={onEdgesChange}
          onConnect={onConnect}
          onNodeClick={(_, node) => setSelectedNodeId(node.id)}
          onDragOver={handleDragOver}
          onDragLeave={handleDragLeave}
          onDrop={handleDrop}
          onNodeDoubleClick={(_, node) => handleNodeDoubleClick(node.id)}
          nodesDraggable={true}
          nodesConnectable={true}
          proOptions={{ hideAttribution: true }}
          style={isRtl ? { direction: 'rtl' } : undefined}
        >
          <Background variant={BackgroundVariant.Dots} gap={12} size={1} color="#334155" />
          <PhaseBands />
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

      <NodeConfig
        open={nodeConfigOpen}
        onOpenChange={setNodeConfigOpen}
        nodeId={nodeConfigNodeId}
        onSave={handleNodeConfigSave}
      />

      {/* Skill Marketplace Button */}
      <button
        className="marketplace-trigger-btn"
        onClick={() => setMarketplaceOpen(true)}
        aria-label="Open Skill Marketplace"
        title="Skill Marketplace"
      >
        &#9776; Skills
      </button>

      {/* Skill Marketplace Modal */}
      <SkillMarketplace open={marketplaceOpen} onOpenChange={setMarketplaceOpen} />

      {/* Accessibility Toolbar */}
      <AccessibilityToolbar visible={accessibilityVisible} onToggle={() => setAccessibilityVisible((v) => !v)} />
    </div>
  );
}

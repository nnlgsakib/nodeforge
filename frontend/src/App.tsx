import { useState, useCallback } from 'react';
import {
  ReactFlow,
  Background,
  Controls,
  MiniMap,
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
  const [nodes, , onNodesChange] = useNodesState(initialNodes);
  const [edges, setEdges, onEdgesChange] = useEdgesState(initialEdges);
  const onConnect: OnConnect = useCallback(
    (connection) => setEdges((edges) => addEdge(connection, edges)),
    [setEdges]
  );

  const [notification, setNotification] = useState<{ type: 'success' | 'error'; message: string } | null>(null);

  const handleCreateProject = async (projectName: string) => {
    try {
      const result = await createProject(projectName);
      console.log('Project created:', result);
      setNotification({ type: 'success', message: `Project "${projectName}" created! Session: ${result.sessionId}` });
    } catch (err) {
      console.error('Failed to create project:', err);
      setNotification({ type: 'error', message: err instanceof Error ? err.message : String(err) });
    }
    setTimeout(() => setNotification(null), 5000);
  };

  return (
    <div className="app-container">
      {notification && (
        <div className={`notification notification-${notification.type}`}>
          {notification.message}
          <button className="notification-close" onClick={() => setNotification(null)}>×</button>
        </div>
      )}
      <div className="sidebar">
        <SessionExplorer onCreateProject={handleCreateProject} />
      </div>
      <div className="main-content">
        <ChatPanel onCreateProject={handleCreateProject} />
        <ReactFlow
          nodes={nodes}
          nodeTypes={nodeTypes}
          onNodesChange={onNodesChange}
          edges={edges}
          edgeTypes={edgeTypes}
          onEdgesChange={onEdgesChange}
          onConnect={onConnect}
          fitView
        >
          <Background />
          <MiniMap />
          <Controls />
        </ReactFlow>
      </div>
    </div>
  );
}

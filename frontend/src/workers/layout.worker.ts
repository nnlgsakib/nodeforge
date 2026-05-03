import * as dagre from 'dagre';

interface LayoutNode {
  id: string;
  width?: number;
  height?: number;
}

interface LayoutEdge {
  source: string;
  target: string;
}

interface LayoutRequest {
  type: 'layout';
  nodes: LayoutNode[];
  edges: LayoutEdge[];
  config?: {
    rankdir?: 'TB' | 'BT' | 'LR' | 'RL';
    nodeWidth?: number;
    nodeHeight?: number;
    ranksep?: number;
    nodesep?: number;
  };
}

interface LayoutResponse {
  type: 'layout-done';
  positions: Record<string, { x: number; y: number }>;
}

interface LayoutErrorResponse {
  type: 'layout-error';
  error: string;
}

const defaultNodeWidth = 200;
const defaultNodeHeight = 80;

self.onmessage = (event: MessageEvent<LayoutRequest>) => {
  const { type, nodes, edges, config } = event.data;

  if (type !== 'layout') {
    return;
  }

  try {
    const g = new dagre.graphlib.Graph();
    g.setGraph({
      rankdir: config?.rankdir ?? 'TB',
      ranksep: config?.ranksep ?? 50,
      nodesep: config?.nodesep ?? 50,
    });

    g.setDefaultEdgeLabel(() => ({}));

    const nodeWidth = config?.nodeWidth ?? defaultNodeWidth;
    const nodeHeight = config?.nodeHeight ?? defaultNodeHeight;

    // Add nodes, skip duplicates
    const seenNodes = new Set<string>();
    for (const node of nodes) {
      if (seenNodes.has(node.id)) {
        continue;
      }
      seenNodes.add(node.id);
      g.setNode(node.id, { width: node.width ?? nodeWidth, height: node.height ?? nodeHeight });
    }

    // Add edges
    for (const edge of edges) {
      g.setEdge(edge.source, edge.target);
    }

    // Run layout
    dagre.layout(g);

    // Extract positions
    const positions: Record<string, { x: number; y: number }> = {};
    for (const nodeId of g.nodes()) {
      const node = g.node(nodeId);
      positions[nodeId] = { x: node.x, y: node.y };
    }

    const response: LayoutResponse = {
      type: 'layout-done',
      positions,
    };
    self.postMessage(response);
  } catch (err) {
    const errorResponse: LayoutErrorResponse = {
      type: 'layout-error',
      error: err instanceof Error ? err.message : String(err),
    };
    self.postMessage(errorResponse);
  }
};

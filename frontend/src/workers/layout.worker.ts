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
  metrics?: {
    layoutTimeMs: number;
    nodeCount: number;
    edgeCount: number;
  };
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
    const startTime = performance.now();
    const g = new dagre.graphlib.Graph();

    // Optimize for large graphs: reduce ranksep/nodesep for 100+ nodes to tighten layout
    const isLargeGraph = nodes.length > 100;
    const ranksep = config?.ranksep ?? (isLargeGraph ? 30 : 50);
    const nodesep = config?.nodesep ?? (isLargeGraph ? 30 : 50);

    g.setGraph({
      rankdir: config?.rankdir ?? 'TB',
      ranksep,
      nodesep,
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
      // Simplify node dimensions for large graphs to reduce layout computation
      const width = isLargeGraph ? Math.min(node.width ?? nodeWidth, 180) : (node.width ?? nodeWidth);
      const height = isLargeGraph ? Math.min(node.height ?? nodeHeight, 60) : (node.height ?? nodeHeight);
      g.setNode(node.id, { width, height });
    }

    // Add edges
    for (const edge of edges) {
      g.setEdge(edge.source, edge.target);
    }

    // Run layout
    dagre.layout(g);

    const layoutTimeMS = performance.now() - startTime;

    // Extract positions
    const positions: Record<string, { x: number; y: number }> = {};
    for (const nodeId of g.nodes()) {
      const node = g.node(nodeId);
      positions[nodeId] = { x: node.x, y: node.y };
    }

    const response: LayoutResponse = {
      type: 'layout-done',
      positions,
      metrics: {
        layoutTimeMs: Math.round(layoutTimeMS * 100) / 100,
        nodeCount: nodes.length,
        edgeCount: edges.length,
      },
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

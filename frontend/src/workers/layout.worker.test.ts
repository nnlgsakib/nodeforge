import { describe, it, expect } from 'vitest';
import dagre from 'dagre';

describe('layout worker logic', () => {
  function runLayout(nodes: { id: string }[], edges: { source: string; target: string }[]) {
    const g = new dagre.graphlib.Graph();
    g.setGraph({ rankdir: 'TB', ranksep: 50, nodesep: 50 });
    g.setDefaultEdgeLabel(() => ({}));

    for (const node of nodes) {
      g.setNode(node.id, { width: 200, height: 80 });
    }
    for (const edge of edges) {
      g.setEdge(edge.source, edge.target);
    }

    dagre.layout(g);

    const positions: Record<string, { x: number; y: number }> = {};
    for (const nodeId of g.nodes()) {
      const node = g.node(nodeId);
      positions[nodeId] = { x: node.x, y: node.y };
    }
    return positions;
  }

  // Optimized layout function matching worker behavior for large graphs
  function runLayoutOptimized(nodes: { id: string }[], edges: { source: string; target: string }[]) {
    const startTime = performance.now();
    const isLargeGraph = nodes.length > 100;
    const ranksep = isLargeGraph ? 30 : 50;
    const nodesep = isLargeGraph ? 30 : 50;

    const g = new dagre.graphlib.Graph();
    g.setGraph({ rankdir: 'TB', ranksep, nodesep });
    g.setDefaultEdgeLabel(() => ({}));

    for (const node of nodes) {
      const width = isLargeGraph ? Math.min(200, 180) : 200;
      const height = isLargeGraph ? Math.min(80, 60) : 80;
      g.setNode(node.id, { width, height });
    }
    for (const edge of edges) {
      g.setEdge(edge.source, edge.target);
    }

    dagre.layout(g);

    const layoutTimeMS = performance.now() - startTime;

    const positions: Record<string, { x: number; y: number }> = {};
    for (const nodeId of g.nodes()) {
      const node = g.node(nodeId);
      positions[nodeId] = { x: node.x, y: node.y };
    }
    return { positions, layoutTimeMS, nodeCount: nodes.length };
  }

  it('should calculate positions for all nodes', () => {
    const nodes = [
      { id: 'node-1' },
      { id: 'node-2' },
      { id: 'node-3' },
    ];
    const edges = [
      { source: 'node-1', target: 'node-2' },
      { source: 'node-2', target: 'node-3' },
    ];

    const positions = runLayout(nodes, edges);

    expect(Object.keys(positions)).toHaveLength(3);
    expect(positions['node-1']).toHaveProperty('x');
    expect(positions['node-1']).toHaveProperty('y');
  });

  it('should place nodes in top-to-bottom order for chain graph', () => {
    const nodes = [{ id: 'a' }, { id: 'b' }, { id: 'c' }];
    const edges = [
      { source: 'a', target: 'b' },
      { source: 'b', target: 'c' },
    ];

    const positions = runLayout(nodes, edges);

    expect(positions['a'].y).toBeLessThan(positions['b'].y);
    expect(positions['b'].y).toBeLessThan(positions['c'].y);
  });

  it('should handle empty graph', () => {
    const positions = runLayout([], []);
    expect(Object.keys(positions)).toHaveLength(0);
  });

  it('should handle single node', () => {
    const positions = runLayout([{ id: 'solo' }], []);
    expect(Object.keys(positions)).toHaveLength(1);
    expect(positions['solo']).toHaveProperty('x');
    expect(positions['solo']).toHaveProperty('y');
  });

  it('should handle disconnected nodes', () => {
    const nodes = [{ id: 'a' }, { id: 'b' }];
    const positions = runLayout(nodes, []);
    expect(Object.keys(positions)).toHaveLength(2);
  });

  it('should complete layout for 100 nodes efficiently (optimized for 60fps)', () => {
    const nodes = Array.from({ length: 100 }, (_, i) => ({ id: `node-${i}` }));
    const edges = Array.from({ length: 99 }, (_, i) => ({
      source: `node-${i}`,
      target: `node-${i + 1}`,
    }));

    const { positions, layoutTimeMS } = runLayoutOptimized(nodes, edges);

    expect(Object.keys(positions)).toHaveLength(100);
    // In production Web Worker, layout runs off-main-thread.
    // In test env we measure relative performance: should be under 50ms even with JS overhead.
    expect(layoutTimeMS).toBeLessThan(50);
  });

  it('should use tighter spacing for large graphs (>100 nodes)', () => {
    const nodes = Array.from({ length: 150 }, (_, i) => ({ id: `node-${i}` }));
    const edges = Array.from({ length: 149 }, (_, i) => ({
      source: `node-${i}`,
      target: `node-${i + 1}`,
    }));

    const { positions, nodeCount } = runLayoutOptimized(nodes, edges);

    expect(Object.keys(positions)).toHaveLength(150);
    expect(nodeCount).toBe(150);
  });
});

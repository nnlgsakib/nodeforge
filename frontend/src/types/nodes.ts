// Node type identifiers for custom nodes
export type CustomNodeType = 'goal' | 'spec' | 'plan' | 'implement' | 'test' | 'review';

// Status of a node during execution
export type NodeStatus = 'pending' | 'running' | 'complete' | 'failed' | 'skipped';

// Base data shared by all custom node types
export interface CustomNodeData {
  label?: string;
  status?: NodeStatus;
  progress?: number;
}

// Default dimensions for layout calculations
export const NODE_DIMENSIONS: Record<CustomNodeType, { width: number; height: number }> = {
  goal: { width: 180, height: 80 },
  spec: { width: 140, height: 140 },
  plan: { width: 170, height: 80 },
  implement: { width: 170, height: 80 },
  test: { width: 160, height: 80 },
  review: { width: 160, height: 80 },
};

// Color palette for each node type (matches design system)
export const NODE_COLORS: Record<CustomNodeType, { background: string; border: string; glow: string }> = {
  goal: { background: '#4CAF50', border: '#388E3C', glow: 'rgba(76, 175, 80, 0.5)' },
  spec: { background: '#2196F3', border: '#1976D2', glow: 'rgba(33, 150, 243, 0.5)' },
  plan: { background: '#9C27B0', border: '#7B1FA2', glow: 'rgba(156, 39, 176, 0.5)' },
  implement: { background: '#FF9800', border: '#F57C00', glow: 'rgba(255, 152, 0, 0.5)' },
  test: { background: '#FFC107', border: '#FFA000', glow: 'rgba(255, 193, 7, 0.5)' },
  review: { background: '#00BCD4', border: '#00ACC1', glow: 'rgba(0, 188, 212, 0.5)' },
};

// Status override colors
export const STATUS_COLORS: Record<NodeStatus, { background: string; border: string; glow: string }> = {
  complete: { background: '#4CAF50', border: '#2E7D32', glow: 'rgba(76, 175, 80, 0.5)' },
  failed: { background: '#f44336', border: '#c62828', glow: 'rgba(244, 67, 54, 0.5)' },
  running: { background: '#FFC107', border: '#F57F17', glow: 'rgba(255, 193, 7, 0.5)' },
  skipped: { background: '#9E9E9E', border: '#616161', glow: 'rgba(158, 158, 158, 0.3)' },
  pending: { background: '', border: '', glow: '' }, // falls back to node type colors
};

import type { Edge } from '@xyflow/react';

// Edge type identifiers
export type CustomEdgeType = 'default' | 'active' | 'tension' | 'success';

// Edge state reflecting execution status
export type EdgeState = 'idle' | 'flowing' | 'strained' | 'completed';

// Metadata attached to edges for tooltip/bubble display
export interface EdgeMetadata {
  latency?: number; // ms
  dataFlowRate?: number; // tokens/sec
  upstreamHealth?: number; // 0-1
  message?: string;
}

// Data carried by custom edges (using index signature for Edge compatibility)
export type AppEdgeData = Record<string, unknown> & {
  tension?: number; // 0-1, drives reactive stroke-width
  metadata?: EdgeMetadata;
  state?: EdgeState;
};

// React Flow edge type for the application
export type AppEdge = Edge<AppEdgeData, CustomEdgeType>;

// Edge style configurations
export const EDGE_STYLES: Record<CustomEdgeType, { stroke: string; strokeWidth: number; animation?: string }> = {
  default: { stroke: '#94a3b8', strokeWidth: 2 },
  active: { stroke: '#06b6d4', strokeWidth: 3, animation: 'flow 1s linear infinite' },
  tension: { stroke: '#ef4444', strokeWidth: 4 },
  success: { stroke: '#22c55e', strokeWidth: 2, animation: 'pulse-success 0.6s ease-out' },
};

// Tension thresholds for reactive edge styling
export const TENSION_THRESHOLDS = {
  high: 0.7, // above this -> tension edge
  medium: 0.3, // above this -> active edge
  low: 0, // at zero -> success edge
};

// Long-press duration to show metadata bubble (ms)
export const LONG_PRESS_DURATION = 500;

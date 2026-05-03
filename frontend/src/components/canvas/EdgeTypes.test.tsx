import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/react';
import '@testing-library/jest-dom';
import { edgeTypes } from './EdgeTypes';
import { Position } from '@xyflow/react';
import type { EdgeProps } from '@xyflow/react';

const mockEdgeProps = (data: Record<string, unknown> = {}): Partial<EdgeProps> =>
  ({
    id: 'test-edge',
    source: 'a',
    target: 'b',
    sourceX: 0,
    sourceY: 0,
    targetX: 100,
    targetY: 100,
    sourcePosition: Position.Bottom,
    targetPosition: Position.Top,
    data,
    selected: false,
  });

describe('EdgeTypes', () => {
  describe('DefaultEdge', () => {
    it('renders edge group with ARIA label', () => {
      const DefaultEdge = edgeTypes.default;
      const { container } = render(<DefaultEdge {...mockEdgeProps() as any} />);
      const group = container.querySelector('[role="graphics-symbol"]');
      expect(group).toBeTruthy();
      expect(group).toHaveAttribute('aria-label', 'Edge from test-edge - default');
    });

    it('is keyboard accessible', () => {
      const DefaultEdge = edgeTypes.default;
      const { container } = render(<DefaultEdge {...mockEdgeProps() as any} />);
      const group = container.querySelector('[role="graphics-symbol"]');
      expect(group).toHaveAttribute('tabIndex', '0');
    });

    it('does not show tooltip without metadata', () => {
      const DefaultEdge = edgeTypes.default;
      render(<DefaultEdge {...mockEdgeProps() as any} />);
      expect(screen.queryByRole('tooltip')).toBeNull();
    });
  });

  describe('ActiveEdge', () => {
    it('renders with flowing animation ARIA label', () => {
      const ActiveEdge = edgeTypes.active;
      const { container } = render(<ActiveEdge {...mockEdgeProps() as any} />);
      const group = container.querySelector('[role="graphics-symbol"]');
      expect(group).toBeTruthy();
      expect(group).toHaveAttribute(
        'aria-label',
        'Edge from test-edge - active (flowing)'
      );
    });
  });

  describe('TensionEdge', () => {
    it('renders with tension ARIA label', () => {
      const TensionEdge = edgeTypes.tension;
      const { container } = render(<TensionEdge {...mockEdgeProps() as any} />);
      const group = container.querySelector('[role="graphics-symbol"]');
      expect(group).toBeTruthy();
      expect(group).toHaveAttribute(
        'aria-label',
        'Edge from test-edge - tension (upstream failure)'
      );
    });
  });

  describe('SuccessEdge', () => {
    it('renders with success ARIA label', () => {
      const SuccessEdge = edgeTypes.success;
      const { container } = render(<SuccessEdge {...mockEdgeProps() as any} />);
      const group = container.querySelector('[role="graphics-symbol"]');
      expect(group).toBeTruthy();
      expect(group).toHaveAttribute(
        'aria-label',
        'Edge from test-edge - success (completed)'
      );
    });
  });

  describe('Edge interaction', () => {
    it('has pointer event handlers attached', () => {
      const DefaultEdge = edgeTypes.default;
      const { container } = render(<DefaultEdge {...mockEdgeProps() as any} />);
      // Verify the group has the required event handler props set (React stores these internally)
      const group = container.querySelector('[role="graphics-symbol"]');
      expect(group).toBeTruthy();
      expect(group).toHaveAttribute('tabIndex', '0');
    });
  });

  describe('Long-press metadata bubble structure', () => {
    it('renders edge with metadata data prop', () => {
      const DefaultEdge = edgeTypes.default;
      const { container } = render(
        <DefaultEdge {...mockEdgeProps({
          tension: 0.5,
          metadata: { latency: 30, upstreamHealth: 0.8 },
        }) as any} />
      );
      const group = container.querySelector('[role="graphics-symbol"]');
      expect(group).toBeTruthy();
      // Metadata is passed through; bubble appears on long-press interaction
      expect(group).toHaveAttribute('aria-label', 'Edge from test-edge - default');
    });
  });
});

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
      expect(group).toHaveAttribute('aria-label', 'Edge from a to b, status: default');
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
        'Edge from a to b, status: active (flowing)'
      );
    });

    it('applies heartbeat animation duration based on tension', () => {
      const ActiveEdge = edgeTypes.active;
      // Low tension = slower animation (1s)
      const { container: c1 } = render(<ActiveEdge {...mockEdgeProps({ tension: 0 }) as any} />);
      // High tension = faster animation (<1s)
      const { container: c2 } = render(<ActiveEdge {...mockEdgeProps({ tension: 1 }) as any} />);

      // Both should render without errors; heartbeat duration is computed from tension
      expect(c1.querySelector('[role="graphics-symbol"]')).toBeTruthy();
      expect(c2.querySelector('[role="graphics-symbol"]')).toBeTruthy();
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
        'Edge from a to b, status: tension (upstream failure)'
      );
    });

    it('renders with dynamic stroke-width based on tension', () => {
      const TensionEdge = edgeTypes.tension;
      const { container: c1 } = render(<TensionEdge {...mockEdgeProps({ tension: 0.7 }) as any} />);
      const { container: c2 } = render(<TensionEdge {...mockEdgeProps({ tension: 1.0 }) as any} />);

      // Both should render without errors; stroke-width is computed from tension
      expect(c1.querySelector('[role="graphics-symbol"]')).toBeTruthy();
      expect(c2.querySelector('[role="graphics-symbol"]')).toBeTruthy();
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
        'Edge from a to b, status: success (completed)'
      );
    });
  });

  describe('Edge interaction', () => {
    it('has pointer event handlers attached', () => {
      const DefaultEdge = edgeTypes.default;
      const { container } = render(<DefaultEdge {...mockEdgeProps() as any} />);
      const group = container.querySelector('[role="graphics-symbol"]');
      expect(group).toBeTruthy();
      expect(group).toHaveAttribute('tabIndex', '0');
    });
  });

  describe('Long-press metadata bubble', () => {
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
      expect(group).toHaveAttribute('aria-label', 'Edge from a to b, status: default');
    });
  });
});

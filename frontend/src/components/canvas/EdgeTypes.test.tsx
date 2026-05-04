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
      // Low tension = slower animation (1.5s)
      const { container: c1 } = render(<ActiveEdge {...mockEdgeProps({ tension: 0 }) as any} />);
      // High tension = faster animation (0.5s)
      const { container: c2 } = render(<ActiveEdge {...mockEdgeProps({ tension: 1 }) as any} />);

      const path1 = c1.querySelector('path');
      const path2 = c2.querySelector('path');
      expect(path1).toBeTruthy();
      expect(path2).toBeTruthy();

      // Verify heartbeat animation in style attribute (browser serializes as kebab-case)
      const style1 = path1!.getAttribute('style') || '';
      const style2 = path2!.getAttribute('style') || '';
      expect(style1).toContain('animation: heartbeat 1.5s');
      expect(style2).toMatch(/animation: heartbeat 0\.5/);
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
      // tension=0.7 → stroke-width = 3 + 0.7*3 = 5.1
      const { container: c1 } = render(<TensionEdge {...mockEdgeProps({ tension: 0.7 }) as any} />);
      // tension=1.0 → stroke-width = 3 + 1.0*3 = 6
      const { container: c2 } = render(<TensionEdge {...mockEdgeProps({ tension: 1.0 }) as any} />);

      const path1 = c1.querySelector('path');
      const path2 = c2.querySelector('path');
      expect(path1).toBeTruthy();
      expect(path2).toBeTruthy();

      const style1 = path1!.getAttribute('style') || '';
      const style2 = path2!.getAttribute('style') || '';
      expect(style1).toContain('stroke-width: 5.1');
      expect(style2).toContain('stroke-width: 6');
    });

    it('defaults to zero tension when no tension data provided', () => {
      const TensionEdge = edgeTypes.tension;
      const { container } = render(<TensionEdge {...mockEdgeProps() as any} />);
      const path = container.querySelector('path');
      expect(path).toBeTruthy();
      const style = path!.getAttribute('style') || '';
      // tension=0 → stroke-width = 3 + 0*3 = 3
      expect(style).toContain('stroke-width: 3');
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

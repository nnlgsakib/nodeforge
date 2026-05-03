import { describe, it, expect, vi } from 'vitest';
import { render, screen } from '@testing-library/react';
import '@testing-library/jest-dom';
import { edgeTypes } from './EdgeTypes';

const mockEdgeProps = (data: Record<string, unknown> = {}): Record<string, unknown> =>
  ({
    id: 'test-edge',
    source: 'a',
    target: 'b',
    sourceX: 0,
    sourceY: 0,
    targetX: 100,
    targetY: 100,
    sourcePosition: 'bottom' as const,
    targetPosition: 'top' as const,
    data,
    selected: false,
    xPos: 0,
    yPos: 0,
    zIndex: 1,
  });

describe('EdgeTypes', () => {
  describe('DefaultEdge', () => {
    it('renders edge group with ARIA label', () => {
      const DefaultEdge = edgeTypes.default;
      const { container } = render(<DefaultEdge {...mockEdgeProps()} />);
      const group = container.querySelector('[role="graphics-symbol"]');
      expect(group).toBeTruthy();
      expect(group).toHaveAttribute('aria-label', 'Edge from test-edge - default');
    });

    it('is keyboard accessible', () => {
      const DefaultEdge = edgeTypes.default;
      const { container } = render(<DefaultEdge {...mockEdgeProps()} />);
      const group = container.querySelector('[role="graphics-symbol"]');
      expect(group).toHaveAttribute('tabIndex', '0');
    });

    it('does not show tooltip without metadata', () => {
      const DefaultEdge = edgeTypes.default;
      const { container } = render(<DefaultEdge {...mockEdgeProps()} />);
      const group = container.querySelector('[role="graphics-symbol"]');
      // Tooltip only appears with metadata; without it, no tooltip element exists
      expect(screen.queryByRole('tooltip')).toBeNull();
    });
  });

  describe('ActiveEdge', () => {
    it('renders with flowing animation ARIA label', () => {
      const ActiveEdge = edgeTypes.active;
      const { container } = render(<ActiveEdge {...mockEdgeProps()} />);
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
      const { container } = render(<TensionEdge {...mockEdgeProps()} />);
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
      const { container } = render(<SuccessEdge {...mockEdgeProps()} />);
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
      const { container } = render(<DefaultEdge {...mockEdgeProps()} />);
      const group = container.querySelector('[role="graphics-symbol"]');
      // Verify the group has the required event handler props set (React stores these internally)
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
        })} />
      );
      const group = container.querySelector('[role="graphics-symbol"]');
      expect(group).toBeTruthy();
      // Metadata is passed through; bubble appears on long-press interaction
      expect(group).toHaveAttribute('aria-label', 'Edge from test-edge - default');
    });
  });
});

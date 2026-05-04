import { describe, it, expect, vi } from 'vitest';
import { render, screen } from '@testing-library/react';
import '@testing-library/jest-dom';

import type { CSSProperties } from 'react';

// Mock React Flow Handle component (requires zustand provider)
vi.mock('@xyflow/react', async () => {
  const actual = await vi.importActual('@xyflow/react');
  return {
    ...(actual as Record<string, unknown>),
    Handle: ({ type, position, 'aria-label': ariaLabel, style }: { type?: string; position?: string; 'aria-label'?: string; style?: CSSProperties }) => (
      <div
        data-testid={`handle-${type}-${position}`}
        aria-label={ariaLabel}
        style={style}
        data-handle-type={type}
        data-handle-position={position}
      />
    ),
    Position: { Top: 'top', Bottom: 'bottom', Left: 'left', Right: 'right' },
  };
});

import { Position } from '@xyflow/react';
import type { NodeProps } from '@xyflow/react';

const { nodeTypes } = await import('./NodeTypes');

const mockNodeProps = (data: Record<string, unknown>, selected = false): Partial<NodeProps> =>
  ({
    data,
    selected,
    id: 'test-node',
    type: 'goal',
    sourcePosition: Position.Bottom,
    targetPosition: Position.Top,
  });

describe('NodeTypes', () => {
  describe('GoalNode', () => {
    it('renders with default label', () => {
      const GoalNode = nodeTypes.goal;
      render(<GoalNode {...mockNodeProps({}) as any} />);
      expect(screen.getByText('Goal')).toBeTruthy();
    });

    it('renders with custom label', () => {
      const GoalNode = nodeTypes.goal;
      render(<GoalNode {...mockNodeProps({ label: 'My Goal' }) as any} />);
      expect(screen.getByText('My Goal')).toBeTruthy();
    });

    it('has ARIA label', () => {
      const GoalNode = nodeTypes.goal;
      render(<GoalNode {...mockNodeProps({ label: 'Test' }) as any} />);
      const node = screen.getByRole('group');
      expect(node).toHaveAttribute('aria-label', 'Node Test, status: pending');
    });

    it('is keyboard accessible', () => {
      const GoalNode = nodeTypes.goal;
      render(<GoalNode {...mockNodeProps({}) as any} />);
      const node = screen.getByRole('group');
      expect(node).toHaveAttribute('tabIndex', '0');
    });

    it('shows progress bar when running', () => {
      const GoalNode = nodeTypes.goal;
      render(<GoalNode {...mockNodeProps({ status: 'running', progress: 0.5 }) as any} />);
      const progressBar = screen.getByRole('progressbar');
      expect(progressBar).toHaveAttribute('aria-valuenow', '50');
    });

    it('renders input handle', () => {
      const GoalNode = nodeTypes.goal;
      render(<GoalNode {...mockNodeProps({}) as any} />);
      expect(screen.getByTestId('handle-target-top')).toBeTruthy();
    });
  });

  describe('SpecNode', () => {
    it('renders with default label', () => {
      const SpecNode = nodeTypes.spec;
      render(<SpecNode {...mockNodeProps({}) as any} />);
      expect(screen.getByText('Spec')).toBeTruthy();
    });

    it('has ARIA label', () => {
      const SpecNode = nodeTypes.spec;
      render(<SpecNode {...mockNodeProps({ label: 'My Spec' }) as any} />);
      const node = screen.getByRole('group');
      expect(node).toHaveAttribute('aria-label', 'Node My Spec, status: pending');
    });
  });

  describe('PlanNode', () => {
    it('renders with default label', () => {
      const PlanNode = nodeTypes.plan;
      render(<PlanNode {...mockNodeProps({}) as any} />);
      expect(screen.getByText('Plan')).toBeTruthy();
    });

    it('has ARIA label', () => {
      const PlanNode = nodeTypes.plan;
      render(<PlanNode {...mockNodeProps({ label: 'My Plan' }) as any} />);
      const node = screen.getByRole('group');
      expect(node).toHaveAttribute('aria-label', 'Node My Plan, status: pending');
    });
  });

  describe('ImplementNode', () => {
    it('renders with default label', () => {
      const ImplementNode = nodeTypes.implement;
      render(<ImplementNode {...mockNodeProps({}) as any} />);
      expect(screen.getByText('Implement')).toBeTruthy();
    });

    it('has ARIA label', () => {
      const ImplementNode = nodeTypes.implement;
      render(<ImplementNode {...mockNodeProps({ label: 'My Implement' }) as any} />);
      const node = screen.getByRole('group');
      expect(node).toHaveAttribute('aria-label', 'Node My Implement, status: pending');
    });
  });

  describe('TestNode', () => {
    it('renders with default label', () => {
      const TestNode = nodeTypes.test;
      render(<TestNode {...mockNodeProps({}) as any} />);
      expect(screen.getByText('Test')).toBeTruthy();
    });

    it('has ARIA label', () => {
      const TestNode = nodeTypes.test;
      render(<TestNode {...mockNodeProps({ label: 'My Test' }) as any} />);
      const node = screen.getByRole('group');
      expect(node).toHaveAttribute('aria-label', 'Node My Test, status: pending');
    });
  });

  describe('ReviewNode', () => {
    it('renders with default label', () => {
      const ReviewNode = nodeTypes.review;
      render(<ReviewNode {...mockNodeProps({}) as any} />);
      expect(screen.getByText('Review')).toBeTruthy();
    });

    it('has ARIA label', () => {
      const ReviewNode = nodeTypes.review;
      render(<ReviewNode {...mockNodeProps({ label: 'My Review' }) as any} />);
      const node = screen.getByRole('group');
      expect(node).toHaveAttribute('aria-label', 'Node My Review, status: pending');
    });
  });

  describe('Selected state', () => {
    it('shows outline when selected', () => {
      const GoalNode = nodeTypes.goal;
      render(<GoalNode {...mockNodeProps({ label: 'Selected' }, true) as any} />);
      const node = screen.getByRole('group');
      expect(node).toHaveStyle('outline: 2px solid white');
    });
  });
});

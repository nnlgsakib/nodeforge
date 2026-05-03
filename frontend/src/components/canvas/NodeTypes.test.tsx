import { describe, it, expect, vi } from 'vitest';
import { render, screen } from '@testing-library/react';
import '@testing-library/jest-dom';

// Mock React Flow Handle component (requires zustand provider)
vi.mock('@xyflow/react', async () => {
  const actual = await vi.importActual('@xyflow/react');
  return {
    ...(actual as Record<string, unknown>),
    Handle: ({ type, position, 'aria-label': ariaLabel, style }: Record<string, unknown>) => (
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

const { nodeTypes } = await import('./NodeTypes');

const mockNodeProps = (data: Record<string, unknown>, selected = false): Record<string, unknown> =>
  ({
    data,
    selected,
    id: 'test-node',
    type: 'goal',
    sourcePosition: 'bottom',
    targetPosition: 'top',
    xPos: 0,
    yPos: 0,
    zIndex: 1,
    position: { x: 0, y: 0 },
  });

describe('NodeTypes', () => {
  describe('GoalNode', () => {
    it('renders with default label', () => {
      const GoalNode = nodeTypes.goal;
      render(<GoalNode {...mockNodeProps({})} />);
      expect(screen.getByText('Goal')).toBeTruthy();
    });

    it('renders with custom label', () => {
      const GoalNode = nodeTypes.goal;
      render(<GoalNode {...mockNodeProps({ label: 'My Goal' })} />);
      expect(screen.getByText('My Goal')).toBeTruthy();
    });

    it('has ARIA label', () => {
      const GoalNode = nodeTypes.goal;
      render(<GoalNode {...mockNodeProps({ label: 'Test' })} />);
      const node = screen.getByRole('group');
      expect(node).toHaveAttribute('aria-label', 'Goal node: Test');
    });

    it('is keyboard accessible', () => {
      const GoalNode = nodeTypes.goal;
      render(<GoalNode {...mockNodeProps({})} />);
      const node = screen.getByRole('group');
      expect(node).toHaveAttribute('tabIndex', '0');
    });

    it('shows progress bar when running', () => {
      const GoalNode = nodeTypes.goal;
      render(<GoalNode {...mockNodeProps({ status: 'running', progress: 0.5 })} />);
      const progressBar = screen.getByRole('progressbar');
      expect(progressBar).toHaveAttribute('aria-valuenow', '50');
    });

    it('renders input handle', () => {
      const GoalNode = nodeTypes.goal;
      render(<GoalNode {...mockNodeProps({})} />);
      expect(screen.getByTestId('handle-target-top')).toBeTruthy();
    });
  });

  describe('SpecNode', () => {
    it('renders with default label', () => {
      const SpecNode = nodeTypes.spec;
      render(<SpecNode {...mockNodeProps({})} />);
      expect(screen.getByText('Spec')).toBeTruthy();
    });

    it('has ARIA label', () => {
      const SpecNode = nodeTypes.spec;
      render(<SpecNode {...mockNodeProps({ label: 'My Spec' })} />);
      const node = screen.getByRole('group');
      expect(node).toHaveAttribute('aria-label', 'Spec node: My Spec');
    });
  });

  describe('PlanNode', () => {
    it('renders with default label', () => {
      const PlanNode = nodeTypes.plan;
      render(<PlanNode {...mockNodeProps({})} />);
      expect(screen.getByText('Plan')).toBeTruthy();
    });

    it('has ARIA label', () => {
      const PlanNode = nodeTypes.plan;
      render(<PlanNode {...mockNodeProps({ label: 'My Plan' })} />);
      const node = screen.getByRole('group');
      expect(node).toHaveAttribute('aria-label', 'Plan node: My Plan');
    });
  });

  describe('ImplementNode', () => {
    it('renders with default label', () => {
      const ImplementNode = nodeTypes.implement;
      render(<ImplementNode {...mockNodeProps({})} />);
      expect(screen.getByText('Implement')).toBeTruthy();
    });

    it('has ARIA label', () => {
      const ImplementNode = nodeTypes.implement;
      render(<ImplementNode {...mockNodeProps({ label: 'My Implement' })} />);
      const node = screen.getByRole('group');
      expect(node).toHaveAttribute('aria-label', 'Implement node: My Implement');
    });
  });

  describe('TestNode', () => {
    it('renders with default label', () => {
      const TestNode = nodeTypes.test;
      render(<TestNode {...mockNodeProps({})} />);
      expect(screen.getByText('Test')).toBeTruthy();
    });

    it('has ARIA label', () => {
      const TestNode = nodeTypes.test;
      render(<TestNode {...mockNodeProps({ label: 'My Test' })} />);
      const node = screen.getByRole('group');
      expect(node).toHaveAttribute('aria-label', 'Test node: My Test');
    });
  });

  describe('ReviewNode', () => {
    it('renders with default label', () => {
      const ReviewNode = nodeTypes.review;
      render(<ReviewNode {...mockNodeProps({})} />);
      expect(screen.getByText('Review')).toBeTruthy();
    });

    it('has ARIA label', () => {
      const ReviewNode = nodeTypes.review;
      render(<ReviewNode {...mockNodeProps({ label: 'My Review' })} />);
      const node = screen.getByRole('group');
      expect(node).toHaveAttribute('aria-label', 'Review node: My Review');
    });
  });

  describe('Selected state', () => {
    it('shows outline when selected', () => {
      const GoalNode = nodeTypes.goal;
      render(<GoalNode {...mockNodeProps({ label: 'Selected' }, true)} />);
      const node = screen.getByRole('group');
      expect(node).toHaveStyle('outline: 2px solid white');
    });
  });
});

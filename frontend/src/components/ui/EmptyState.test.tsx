import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { EmptyState } from './EmptyState';

describe('EmptyState', () => {
  it('renders icon and title', () => {
    render(<EmptyState icon={<span data-testid="icon">📭</span>} title="No sessions yet" />);
    expect(screen.getByTestId('icon')).toBeInTheDocument();
    expect(screen.getByText('No sessions yet')).toBeInTheDocument();
  });

  it('renders description when provided', () => {
    render(
      <EmptyState
        icon={<span>📭</span>}
        title="No sessions yet"
        description="Start a new project to begin"
      />
    );
    expect(screen.getByText('Start a new project to begin')).toBeInTheDocument();
  });

  it('renders action button when actionLabel and onAction are provided', () => {
    const handleClick = vi.fn();
    render(
      <EmptyState
        icon={<span>📭</span>}
        title="No sessions yet"
        actionLabel="Start Chat"
        onAction={handleClick}
      />
    );
    const button = screen.getByRole('button', { name: 'Start Chat' });
    expect(button).toBeInTheDocument();
    fireEvent.click(button);
    expect(handleClick).toHaveBeenCalledTimes(1);
  });

  it('does not render action button when onAction is missing', () => {
    render(
      <EmptyState
        icon={<span>📭</span>}
        title="No sessions yet"
        actionLabel="Start Chat"
      />
    );
    expect(screen.queryByRole('button')).not.toBeInTheDocument();
  });

  it('renders animated ellipsis when animated is true', () => {
    render(
      <EmptyState
        icon={<span>⏳</span>}
        title="Loading..."
        animated
      />
    );
    const animatedEl = document.querySelector('.empty-state-animated-ellipsis');
    expect(animatedEl).toBeInTheDocument();
  });

  it('does not render animated ellipsis when animated is false', () => {
    render(
      <EmptyState
        icon={<span>📭</span>}
        title="No sessions yet"
      />
    );
    const animatedEl = document.querySelector('.empty-state-animated-ellipsis');
    expect(animatedEl).not.toBeInTheDocument();
  });

  it('has role="status" and aria-live="polite" for accessibility', () => {
    render(<EmptyState icon={<span>📭</span>} title="No sessions yet" />);
    const container = document.querySelector('.empty-state');
    expect(container).toHaveAttribute('role', 'status');
    expect(container).toHaveAttribute('aria-live', 'polite');
  });
});

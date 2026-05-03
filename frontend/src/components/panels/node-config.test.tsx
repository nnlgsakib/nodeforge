import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { NodeConfig } from './node-config';

describe('NodeConfig', () => {
  const mockOnSave = vi.fn();

  beforeEach(() => {
    vi.clearAllMocks();
  });

  const defaultProps = {
    open: true,
    onOpenChange: vi.fn(),
    nodeId: 'test-node-1',
    onSave: mockOnSave,
  };

  it('renders dialog when open', () => {
    render(<NodeConfig {...defaultProps} />);

    expect(screen.getByText(/node configuration/i)).toBeInTheDocument();
    expect(screen.getByLabelText(/timeout/i)).toBeInTheDocument();
    expect(screen.getByLabelText(/retry count/i)).toBeInTheDocument();
    expect(screen.getByLabelText(/token budget/i)).toBeInTheDocument();
  });

  it('does not render dialog when closed', () => {
    render(<NodeConfig {...defaultProps} open={false} />);

    expect(screen.queryByText(/node configuration/i)).not.toBeInTheDocument();
  });

  it('shows default values', () => {
    render(<NodeConfig {...defaultProps} />);

    const timeoutInput = screen.getByRole('spinbutton', { name: /timeout/i }) as HTMLInputElement;
    const retryInput = screen.getAllByRole('spinbutton')[1] as HTMLInputElement;

    expect(timeoutInput.value).toBe('60');
    expect(retryInput.value).toBe('3');
  });

  it('shows validation error for invalid timeout', async () => {
    render(<NodeConfig {...defaultProps} />);

    const timeoutInput = screen.getByRole('spinbutton', { name: /timeout/i });
    fireEvent.change(timeoutInput, { target: { value: '0' } });

    await waitFor(() => {
      expect(screen.getByText(/timeout must be 1-300 seconds/i)).toBeInTheDocument();
    });
  });

  it('shows validation error for invalid retry count', async () => {
    render(<NodeConfig {...defaultProps} />);

    const retryInput = screen.getAllByRole('spinbutton')[1];
    fireEvent.change(retryInput, { target: { value: '-1' } });

    await waitFor(() => {
      expect(screen.getByText(/retry count must be 0-10/i)).toBeInTheDocument();
    });
  });

  it('shows validation error for invalid token budget', async () => {
    render(<NodeConfig {...defaultProps} />);

    const budgetInput = screen.getAllByRole('spinbutton')[2];
    fireEvent.change(budgetInput, { target: { value: '50' } });

    await waitFor(() => {
      expect(screen.getByText(/budget must be 100-100000 tokens/i)).toBeInTheDocument();
    });
  });

  it('disables save button when validation fails', () => {
    render(<NodeConfig {...defaultProps} />);

    const timeoutInput = screen.getByRole('spinbutton', { name: /timeout/i });
    fireEvent.change(timeoutInput, { target: { value: '0' } });

    const saveButton = screen.getByRole('button', { name: /save/i });
    expect(saveButton).toBeDisabled();
  });

  it('calls onSave with valid config when save button clicked', () => {
    render(<NodeConfig {...defaultProps} />);

    const saveButton = screen.getByRole('button', { name: /save/i });
    fireEvent.click(saveButton);

    expect(mockOnSave).toHaveBeenCalledWith('test-node-1', {
      timeout: 60,
      retryCount: 3,
      tokenBudget: 10000,
    });
  });

  it('uses initialConfig values when provided', () => {
    render(
      <NodeConfig
        {...defaultProps}
        initialConfig={{ timeout: 120, retryCount: 5, tokenBudget: 50000 }}
      />
    );

    const timeoutInput = screen.getByRole('spinbutton', { name: /timeout/i }) as HTMLInputElement;
    const retryInput = screen.getAllByRole('spinbutton')[1] as HTMLInputElement;

    expect(timeoutInput.value).toBe('120');
    expect(retryInput.value).toBe('5');
  });

  it('calls onOpenChange when close button clicked', () => {
    const mockOnOpenChange = vi.fn();
    render(<NodeConfig {...defaultProps} onOpenChange={mockOnOpenChange} />);

    const closeButton = screen.getByRole('button', { name: /close/i });
    fireEvent.click(closeButton);

    expect(mockOnOpenChange).toHaveBeenCalledWith(false);
  });
});

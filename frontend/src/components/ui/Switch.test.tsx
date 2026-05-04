import { describe, it, expect, vi } from 'vitest';
import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { Switch } from './Switch';

describe('Switch', () => {
  it('renders with label', () => {
    render(<Switch id="test-switch" label="Test Label" />);
    expect(screen.getByText('Test Label')).toBeInTheDocument();
    expect(screen.getByRole('switch', { name: /test label/i })).toBeInTheDocument();
  });

  it('reflects checked state', () => {
    render(<Switch id="test-switch" checked={true} onCheckedChange={() => {}} />);
    const switchEl = screen.getByRole('switch');
    expect(switchEl).toHaveAttribute('data-state', 'checked');
  });

  it('reflects unchecked state', () => {
    render(<Switch id="test-switch" checked={false} onCheckedChange={() => {}} />);
    const switchEl = screen.getByRole('switch');
    expect(switchEl).toHaveAttribute('data-state', 'unchecked');
  });

  it('calls onCheckedChange when clicked', async () => {
    const handleChange = vi.fn();
    const user = userEvent.setup();
    render(<Switch id="test-switch" checked={false} onCheckedChange={handleChange} />);
    const switchEl = screen.getByRole('switch');
    await user.click(switchEl);
    expect(handleChange).toHaveBeenCalledWith(true);
  });
});

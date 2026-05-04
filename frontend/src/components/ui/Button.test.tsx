import { describe, it, expect, vi } from 'vitest';
import { render, screen } from '@testing-library/react';
import { Button } from './Button';

describe('Button', () => {
  it('renders with default primary variant', () => {
    render(<Button>Click me</Button>);
    const button = screen.getByRole('button', { name: /click me/i });
    expect(button).toBeInTheDocument();
  });

  it('applies primary variant classes', () => {
    render(<Button variant="primary">Primary</Button>);
    const button = screen.getByRole('button', { name: /primary/i });
    expect(button.className).toContain('bg-primary');
    expect(button.className).toContain('text-white');
  });

  it('applies secondary variant classes', () => {
    render(<Button variant="secondary">Secondary</Button>);
    const button = screen.getByRole('button', { name: /secondary/i });
    expect(button.className).toContain('border');
    expect(button.className).toContain('text-primary');
  });

  it('applies danger variant classes', () => {
    render(<Button variant="danger">Danger</Button>);
    const button = screen.getByRole('button', { name: /danger/i });
    expect(button.className).toContain('bg-danger');
    expect(button.className).toContain('text-white');
  });

  it('applies icon variant classes', () => {
    render(<Button variant="icon" aria-label="Icon button">X</Button>);
    const button = screen.getByRole('button', { name: /icon button/i });
    expect(button.className).toContain('w-8');
    expect(button.className).toContain('h-8');
  });

  it('is disabled when disabled prop is set', () => {
    render(<Button disabled>Disabled</Button>);
    const button = screen.getByRole('button', { name: /disabled/i });
    expect(button).toBeDisabled();
  });

  it('calls onClick when clicked', () => {
    const handleClick = vi.fn();
    render(<Button onClick={handleClick}>Clickable</Button>);
    const button = screen.getByRole('button', { name: /clickable/i });
    button.click();
    expect(handleClick).toHaveBeenCalledTimes(1);
  });

  it('does NOT call onClick when disabled', () => {
    const handleClick = vi.fn();
    render(<Button disabled onClick={handleClick}>Disabled</Button>);
    const button = screen.getByRole('button', { name: /disabled/i });
    button.click();
    expect(handleClick).not.toHaveBeenCalled();
  });
});

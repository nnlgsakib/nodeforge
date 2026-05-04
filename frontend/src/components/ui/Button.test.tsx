import { describe, it, expect, vi } from 'vitest';
import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { Button } from './button';

describe('Button', () => {
  it('renders as button by default', () => {
    render(<Button>Click me</Button>);
    const button = screen.getByRole('button', { name: /click me/i });
    expect(button).toBeInTheDocument();
    expect(button.tagName).toBe('BUTTON');
  });

  it('renders as child when asChild is true', () => {
    render(<Button asChild><a href="/test">Link</a></Button>);
    const link = screen.getByRole('link', { name: /link/i });
    expect(link.tagName).toBe('A');
  });

  it('applies default variant classes (cyan bg)', () => {
    render(<Button>Primary</Button>);
    const button = screen.getByRole('button', { name: /primary/i });
    expect(button.className).toContain('bg-[#06b6d4]');
    expect(button.className).toContain('text-white');
  });

  it('applies outline variant classes', () => {
    render(<Button variant="outline">Secondary</Button>);
    const button = screen.getByRole('button', { name: /secondary/i });
    expect(button.className).toContain('border-gray-600');
    expect(button.className).toContain('text-gray-200');
  });

  it('applies destructive variant classes (red bg)', () => {
    render(<Button variant="destructive">Danger</Button>);
    const button = screen.getByRole('button', { name: /danger/i });
    expect(button.className).toContain('bg-[#ef4444]');
    expect(button.className).toContain('text-white');
  });

  it('applies ghost variant classes', () => {
    render(<Button variant="ghost">Ghost</Button>);
    const button = screen.getByRole('button', { name: /ghost/i });
    expect(button.className).not.toContain('bg-[');
    expect(button.className).toContain('text-gray-200');
  });

  it('applies icon variant classes (32x32px)', () => {
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

  it('calls onClick when clicked', async () => {
    const handleClick = vi.fn();
    render(<Button onClick={handleClick}>Clickable</Button>);
    const button = screen.getByRole('button', { name: /clickable/i });
    await userEvent.click(button);
    expect(handleClick).toHaveBeenCalledTimes(1);
  });

  it('does NOT call onClick when disabled', async () => {
    const handleClick = vi.fn();
    render(<Button disabled onClick={handleClick}>Disabled</Button>);
    const button = screen.getByRole('button', { name: /disabled/i });
    await userEvent.click(button);
    expect(handleClick).not.toHaveBeenCalled();
  });

  it('has focus-visible outline for accessibility', () => {
    render(<Button>Focused</Button>);
    const button = screen.getByRole('button', { name: /focused/i });
    expect(button.className).toContain('focus-visible:outline-2');
  });
});

import { describe, it, expect, vi } from 'vitest';
import { render, screen } from '@testing-library/react';
import '@testing-library/jest-dom';

// Mock useViewport from @xyflow/react
const mockViewport = { x: 0, y: 0, zoom: 1 };
vi.mock('@xyflow/react', async () => {
  const actual = await vi.importActual('@xyflow/react');
  return {
    ...(actual as Record<string, unknown>),
    useViewport: () => mockViewport,
  };
});

import { PhaseBands } from './PhaseBands';

describe('PhaseBands', () => {
  it('renders phase band labels', () => {
    render(<PhaseBands />);
    expect(screen.getByText('Discovery')).toBeTruthy();
    expect(screen.getByText('Execution')).toBeTruthy();
    expect(screen.getByText('Recovery')).toBeTruthy();
    expect(screen.getByText('Completion')).toBeTruthy();
  });

  it('renders bands as SVG elements', () => {
    const { container } = render(<PhaseBands />);
    const rects = container.querySelectorAll('rect');
    expect(rects.length).toBeGreaterThan(0);
    expect(rects[0]).toHaveAttribute('fill');
  });

  it('renders bands with correct spec colors', () => {
    const { container } = render(<PhaseBands />);
    const rects = container.querySelectorAll('rect');
    expect(rects.length).toBeGreaterThanOrEqual(4);

    // Extract fill colors from first 4 rects (one per phase)
    const fills = Array.from(rects).slice(0, 4).map((r) => r.getAttribute('fill'));

    // Verify spec colors are present: Discovery=#3B82F6, Execution=#F97316, Recovery=#EF4444, Completion=#22C55E
    expect(fills).toContain('#3B82F6');
    expect(fills).toContain('#F97316');
    expect(fills).toContain('#EF4444');
    expect(fills).toContain('#22C55E');
  });

  it('renders labels as SVG text', () => {
    const { container } = render(<PhaseBands />);
    const texts = container.querySelectorAll('text');
    expect(texts.length).toBeGreaterThan(0);
  });
});

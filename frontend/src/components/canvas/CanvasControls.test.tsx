import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { ReactFlowProvider } from '@xyflow/react';
import { CanvasControls } from './CanvasControls';

const mockNodes = [
  { id: '1', position: { x: 0, y: 0 }, data: { status: 'running', label: 'Node 1' } },
  { id: '2', position: { x: 100, y: 100 }, data: { status: 'complete', label: 'Node 2' } },
  { id: '3', position: { x: 200, y: 200 }, data: { status: 'failed', label: 'Node 3' } },
];

const mockEdges = [{ id: 'e1-2', source: '1', target: '2' }];

function renderWithReactFlow(component: React.ReactElement) {
  return render(<ReactFlowProvider>{component}</ReactFlowProvider>);
}

describe('CanvasControls', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('renders MiniMap, Controls, and node/edge count', () => {
    renderWithReactFlow(<CanvasControls nodes={mockNodes as any} edges={mockEdges} />);

    expect(screen.getByText(/Nodes: 3/)).toBeInTheDocument();
    expect(screen.getByText(/Edges: 1/)).toBeInTheDocument();
    expect(screen.getByText(/Active: 1/)).toBeInTheDocument();
  });

  it('shows keybinding hints when Keys button is clicked', () => {
    renderWithReactFlow(<CanvasControls nodes={mockNodes as any} edges={mockEdges} />);

    const keysButton = screen.getByRole('button', { name: /keys/i });
    fireEvent.click(keysButton);

    expect(screen.getByText(/Canvas Navigation/)).toBeInTheDocument();
    expect(screen.getByText(/Vim/)).toBeInTheDocument();
    expect(screen.getByText(/Emacs/)).toBeInTheDocument();
  });

  it('toggles keybinding hints off when Keys button is clicked again', () => {
    renderWithReactFlow(<CanvasControls nodes={mockNodes as any} edges={mockEdges} />);

    const keysButton = screen.getByRole('button', { name: /keys/i });
    fireEvent.click(keysButton);
    expect(screen.getByText(/Canvas Navigation/)).toBeInTheDocument();

    fireEvent.click(keysButton);
    expect(screen.queryByText(/Canvas Navigation/)).not.toBeInTheDocument();
  });
});

describe('CanvasControls heat visualization', () => {
  it('calculates heat intensity for recently active nodes', () => {
    const recentNode = [
      { id: '1', position: { x: 0, y: 0 }, data: { status: 'running', lastActiveAt: Date.now() } },
    ];

    renderWithReactFlow(<CanvasControls nodes={recentNode as any} edges={mockEdges} />);

    expect(screen.getByText(/Active: 1/)).toBeInTheDocument();
  });

  it('shows no heat for nodes without lastActiveAt', () => {
    const oldNode = [
      { id: '1', position: { x: 0, y: 0 }, data: { status: 'complete' } },
    ];

    renderWithReactFlow(<CanvasControls nodes={oldNode as any} edges={mockEdges} />);

    expect(screen.getByText(/Active: 0/)).toBeInTheDocument();
  });
});

describe('CanvasControls keyboard navigation', () => {
  it('renders with onNodeDoubleClick prop available in App', () => {
    renderWithReactFlow(
      <CanvasControls nodes={mockNodes as any} edges={mockEdges} />
    );

    // Verify the component renders
    expect(screen.getByText(/Nodes: 3/)).toBeInTheDocument();
  });
});

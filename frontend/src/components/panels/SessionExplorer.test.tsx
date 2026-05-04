import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent, act } from '@testing-library/react';
import { SessionExplorer } from './SessionExplorer';

interface MockSessionState {
  sessions: typeof mockSessions;
  loading: boolean;
  error: string | null;
  listSessions: ReturnType<typeof vi.fn>;
  createSession: ReturnType<typeof vi.fn>;
  getSession: ReturnType<typeof vi.fn>;
  autoSaveSession: ReturnType<typeof vi.fn>;
}

// Mutable mock state — tests can change this before rendering
const mockState: MockSessionState = {
  sessions: [] as typeof mockSessions,
  loading: false,
  error: null,
  listSessions: vi.fn(),
  createSession: vi.fn(),
  getSession: vi.fn(),
  autoSaveSession: vi.fn(),
};

// Mock the useSession hook — reads from mutable state on each call
vi.mock('../../hooks/useSession', () => ({
  useSession: (): MockSessionState => {
    const s = (globalThis as { __mockSessionState?: MockSessionState }).__mockSessionState;
    return {
      sessions: s?.sessions ?? [],
      loading: s?.loading ?? false,
      error: s?.error ?? null,
      listSessions: s?.listSessions ?? vi.fn(),
      createSession: s?.createSession ?? vi.fn(),
      getSession: s?.getSession ?? vi.fn(),
      autoSaveSession: s?.autoSaveSession ?? vi.fn(),
    };
  },
}));

const mockSessions = [
  {
    sessionId: 'sess-1',
    projectName: 'nfv2-auth-module',
    status: 'complete' as const,
    goal: 'Build auth module',
    workspace: '/path/nfv2-auth-module',
    createdAt: new Date(Date.now() - 2 * 3600000).toISOString(),
    lastActive: new Date(Date.now() - 1800000).toISOString(),
  },
  {
    sessionId: 'sess-2',
    projectName: 'api-integration',
    status: 'running' as const,
    goal: 'API integration',
    workspace: '/path/api-integration',
    createdAt: new Date(Date.now() - 300000).toISOString(),
    lastActive: new Date(Date.now() - 60000).toISOString(),
  },
  {
    sessionId: 'sess-3',
    projectName: 'dashboard-refactor',
    status: 'failed' as const,
    goal: 'Refactor dashboard',
    workspace: '/path/dashboard',
    createdAt: new Date(Date.now() - 86400000).toISOString(),
    lastActive: new Date(Date.now() - 82800000).toISOString(),
  },
  {
    sessionId: 'sess-4',
    projectName: 'test-suite',
    status: 'paused' as const,
    goal: 'Build tests',
    workspace: '/path/test-suite',
    createdAt: new Date(Date.now() - 7 * 86400000).toISOString(),
    lastActive: new Date(Date.now() - 2 * 86400000).toISOString(),
  },
];

describe('SessionExplorer', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    // Reset mock state to default (4 sessions) before each test
    mockState.sessions = [...mockSessions];
    mockState.loading = false;
    mockState.error = null;
    mockState.listSessions = vi.fn();
    mockState.createSession = vi.fn();
    mockState.getSession = vi.fn();
    mockState.autoSaveSession = vi.fn();
    // Attach to globalThis so the mock factory can read it
    (globalThis as { __mockSessionState?: MockSessionState }).__mockSessionState = mockState;
  });

  it('renders new project form', () => {
    render(<SessionExplorer onCreateProject={vi.fn()} />);

    expect(screen.getByLabelText(/project name/i)).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /new project/i })).toBeInTheDocument();
  });

  it('calls onCreateProject when form submitted', async () => {
    const mockCreate = vi.fn();
    render(<SessionExplorer onCreateProject={mockCreate} />);

    const input = screen.getByLabelText(/project name/i);
    const button = screen.getByRole('button', { name: /new project/i });

    await act(async () => {
      fireEvent.change(input, { target: { value: 'my-test-project' } });
      fireEvent.click(button);
    });

    expect(mockCreate).toHaveBeenCalledWith('my-test-project');
  });

  it('disables create button when project name is empty', () => {
    render(<SessionExplorer onCreateProject={vi.fn()} />);

    const button = screen.getByRole('button', { name: /new project/i });
    expect(button).toBeDisabled();
  });

  it('renders session list', () => {
    render(<SessionExplorer onCreateProject={vi.fn()} />);

    expect(screen.getByText(/nfv2-auth-module/i)).toBeInTheDocument();
    expect(screen.getByText(/api-integration/i)).toBeInTheDocument();
    expect(screen.getByText(/dashboard-refactor/i)).toBeInTheDocument();
    expect(screen.getByText(/test-suite/i)).toBeInTheDocument();
  });

  it('shows session count', () => {
    render(<SessionExplorer onCreateProject={vi.fn()} />);

    expect(screen.getByText(/sessions \(4\)/i)).toBeInTheDocument();
  });

  it('filters sessions by search query', async () => {
    render(<SessionExplorer onCreateProject={vi.fn()} />);

    const searchInput = screen.getByPlaceholderText(/search by project name/i);
    await act(async () => {
      fireEvent.change(searchInput, { target: { value: 'api' } });
    });

    expect(screen.getByText(/api-integration/i)).toBeInTheDocument();
    expect(screen.queryByText(/nfv2-auth-module/i)).not.toBeInTheDocument();
  });

  it('filters sessions by status', async () => {
    render(<SessionExplorer onCreateProject={vi.fn()} />);

    const statusSelect = screen.getByLabelText(/filter by status/i);
    await act(async () => {
      fireEvent.change(statusSelect, { target: { value: 'complete' } });
    });

    expect(screen.getByText(/nfv2-auth-module/i)).toBeInTheDocument();
    expect(screen.queryByText(/api-integration/i)).not.toBeInTheDocument();
  });

  it('calls onResumeSession when resume button clicked', async () => {
    const mockResume = vi.fn();
    render(<SessionExplorer onCreateProject={vi.fn()} onResumeSession={mockResume} />);

    // Find the test-suite card (paused status, so it has a Resume button)
    const testSuiteCard = screen.getByText(/test-suite/i).closest('.session-card');
    const resumeButton = testSuiteCard?.querySelector('button');
    await act(async () => {
      fireEvent.click(resumeButton!);
    });

    expect(mockResume).toHaveBeenCalledWith('sess-4');
  });

  it('calls onForkSession when fork button clicked', async () => {
    const mockFork = vi.fn();
    render(<SessionExplorer onCreateProject={vi.fn()} onForkSession={mockFork} />);

    const forkButtons = screen.getAllByRole('button', { name: /fork/i });
    await act(async () => {
      fireEvent.click(forkButtons[0]);
    });

    expect(mockFork).toHaveBeenCalledWith('sess-1');
  });

  it('calls onExportSession when export button clicked', async () => {
    const mockExport = vi.fn();
    render(<SessionExplorer onCreateProject={vi.fn()} onExportSession={mockExport} />);

    const exportButtons = screen.getAllByRole('button', { name: /export/i });
    await act(async () => {
      fireEvent.click(exportButtons[0]);
    });

    expect(mockExport).toHaveBeenCalledWith('sess-1');
  });

  it('shows date filter options', () => {
    render(<SessionExplorer onCreateProject={vi.fn()} />);

    const dateSelect = screen.getByLabelText(/filter by date/i);
    expect(dateSelect).toBeInTheDocument();
  });

  it('shows empty state with Start Chat button when no sessions exist', () => {
    // Override sessions to empty for this test
    mockState.sessions = [];

    const handleStartChat = vi.fn();
    render(<SessionExplorer onCreateProject={vi.fn()} onStartChat={handleStartChat} />);

    expect(screen.getByText('No sessions yet')).toBeInTheDocument();
    expect(screen.getByText('Start a new project to begin your journey')).toBeInTheDocument();
    const startBtn = screen.getByRole('button', { name: 'Start Chat' });
    expect(startBtn).toBeInTheDocument();

    fireEvent.click(startBtn);
    expect(handleStartChat).toHaveBeenCalled();
  });

  it('shows no match message when filter returns no results', async () => {
    render(<SessionExplorer onCreateProject={vi.fn()} />);

    const searchInput = screen.getByPlaceholderText(/search by project name/i);
    await act(async () => {
      fireEvent.change(searchInput, { target: { value: 'zzznonexistent' } });
    });

    expect(screen.getByText('No sessions match your filters')).toBeInTheDocument();
  });

  it('shows session status badges', () => {
    render(<SessionExplorer onCreateProject={vi.fn()} />);

    expect(screen.getByText('complete')).toBeInTheDocument();
    expect(screen.getByText('running')).toBeInTheDocument();
    expect(screen.getByText('failed')).toBeInTheDocument();
    expect(screen.getByText('paused')).toBeInTheDocument();
  });

  it('session cards are keyboard accessible', async () => {
    const mockSelect = vi.fn();
    render(<SessionExplorer onCreateProject={vi.fn()} onSelectSession={mockSelect} />);

    const card = screen.getByText(/nfv2-auth-module/i).closest('.session-card');
    expect(card).toHaveAttribute('tabindex', '0');

    await act(async () => {
      fireEvent.keyDown(card!, { key: 'Enter' });
    });

    expect(mockSelect).toHaveBeenCalledWith('sess-1');
  });
});

import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent, waitFor, act } from '@testing-library/react';
import { SessionExplorer } from './SessionExplorer';

describe('SessionExplorer', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    vi.stubGlobal('fetch', vi.fn());
  });

  it('renders new project form', async () => {
    vi.mocked(fetch).mockRejectedValue(new Error('API not available'));
    render(<SessionExplorer onCreateProject={vi.fn()} />);

    await waitFor(() => {
      expect(screen.getAllByPlaceholderText(/project name/i)[0]).toBeInTheDocument();
    });
    expect(screen.getByRole('button', { name: /new project/i })).toBeInTheDocument();
  });

  it('calls onCreateProject when form submitted', async () => {
    vi.mocked(fetch).mockRejectedValue(new Error('API not available'));
    const mockCreate = vi.fn();
    render(<SessionExplorer onCreateProject={mockCreate} />);

    await waitFor(() => {
      expect(screen.getAllByPlaceholderText(/project name/i)[0]).toBeInTheDocument();
    });

    const input = screen.getAllByPlaceholderText(/project name/i)[0];
    const button = screen.getByRole('button', { name: /new project/i });

    await act(async () => {
      fireEvent.change(input, { target: { value: 'my-test-project' } });
      fireEvent.click(button);
    });

    expect(mockCreate).toHaveBeenCalledWith('my-test-project');
  });

  it('disables create button when project name is empty', async () => {
    vi.mocked(fetch).mockRejectedValue(new Error('API not available'));
    render(<SessionExplorer onCreateProject={vi.fn()} />);

    await waitFor(() => {
      const button = screen.getByRole('button', { name: /new project/i });
      expect(button).toBeDisabled();
    });
  });

  it('renders session list with mock data when API unavailable', async () => {
    vi.mocked(fetch).mockRejectedValue(new Error('API not available'));

    render(<SessionExplorer onCreateProject={vi.fn()} />);

    await waitFor(() => {
      expect(screen.getByText(/nfv2-auth-module/i)).toBeInTheDocument();
    });
    expect(screen.getByText(/api-integration/i)).toBeInTheDocument();
  });

  it('filters sessions by search query', async () => {
    vi.mocked(fetch).mockRejectedValue(new Error('API not available'));

    render(<SessionExplorer onCreateProject={vi.fn()} />);

    await waitFor(() => {
      expect(screen.getByText(/nfv2-auth-module/i)).toBeInTheDocument();
    });

    const searchInput = screen.getByPlaceholderText(/search by project name/i);
    await act(async () => {
      fireEvent.change(searchInput, { target: { value: 'api' } });
    });

    expect(screen.getByText(/api-integration/i)).toBeInTheDocument();
    expect(screen.queryByText(/nfv2-auth-module/i)).not.toBeInTheDocument();
  });

  it('filters sessions by status', async () => {
    vi.mocked(fetch).mockRejectedValue(new Error('API not available'));

    render(<SessionExplorer onCreateProject={vi.fn()} />);

    await waitFor(() => {
      expect(screen.getByText(/nfv2-auth-module/i)).toBeInTheDocument();
    });

    const statusSelect = screen.getAllByRole('combobox')[0];
    await act(async () => {
      fireEvent.change(statusSelect, { target: { value: 'complete' } });
    });

    expect(screen.getByText(/nfv2-auth-module/i)).toBeInTheDocument();
    expect(screen.queryByText(/api-integration/i)).not.toBeInTheDocument();
  });

  it('calls onResumeSession when resume button clicked', async () => {
    vi.mocked(fetch).mockRejectedValue(new Error('API not available'));
    const mockResume = vi.fn();

    render(<SessionExplorer onCreateProject={vi.fn()} onResumeSession={mockResume} />);

    await waitFor(() => {
      expect(screen.getByText(/test-suite/i)).toBeInTheDocument();
    });

    const resumeButton = screen.getByRole('button', { name: /resume/i });
    await act(async () => {
      fireEvent.click(resumeButton);
    });

    expect(mockResume).toHaveBeenCalledWith('mock-4');
  });

  it('calls onForkSession when fork button clicked', async () => {
    vi.mocked(fetch).mockRejectedValue(new Error('API not available'));
    const mockFork = vi.fn();

    render(<SessionExplorer onCreateProject={vi.fn()} onForkSession={mockFork} />);

    await waitFor(() => {
      expect(screen.getByText(/nfv2-auth-module/i)).toBeInTheDocument();
    });

    const forkButtons = screen.getAllByRole('button', { name: /fork/i });
    await act(async () => {
      fireEvent.click(forkButtons[0]);
    });

    expect(mockFork).toHaveBeenCalledWith('mock-1');
  });

  it('calls onExportSession when export button clicked', async () => {
    vi.mocked(fetch).mockRejectedValue(new Error('API not available'));
    const mockExport = vi.fn();

    render(<SessionExplorer onCreateProject={vi.fn()} onExportSession={mockExport} />);

    await waitFor(() => {
      expect(screen.getByText(/nfv2-auth-module/i)).toBeInTheDocument();
    });

    const exportButtons = screen.getAllByRole('button', { name: /export/i });
    await act(async () => {
      fireEvent.click(exportButtons[0]);
    });

    expect(mockExport).toHaveBeenCalledWith('mock-1');
  });

  it('shows date filter options', async () => {
    vi.mocked(fetch).mockRejectedValue(new Error('API not available'));

    render(<SessionExplorer onCreateProject={vi.fn()} />);

    await waitFor(() => {
      expect(screen.getByText(/nfv2-auth-module/i)).toBeInTheDocument();
    });

    const dateSelect = screen.getAllByRole('combobox')[1];
    expect(dateSelect).toBeInTheDocument();

    await act(async () => {
      fireEvent.change(dateSelect, { target: { value: 'today' } });
    });
  });
});

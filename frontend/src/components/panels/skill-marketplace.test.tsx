import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { SkillMarketplace } from './skill-marketplace';

// Mock fetch globally
const mockFetch = vi.fn();
vi.stubGlobal('fetch', mockFetch);

describe('SkillMarketplace', () => {
  const defaultProps = {
    open: true,
    onOpenChange: vi.fn(),
  };

  const mockSkills = {
    skills: [
      {
        id: 'skill-code-review',
        name: 'Code Review',
        version: '1.0.0',
        description: 'Automated code review',
        author: 'NodeForge',
        category: 'Development',
        rating: 4.5,
        ratingCount: 128,
        downloads: 1200,
        icon: 'code-review',
        tags: ['review', 'quality'],
        installed: false,
      },
      {
        id: 'skill-test-gen',
        name: 'Test Generator',
        version: '1.2.0',
        description: 'Generate tests automatically',
        author: 'NodeForge',
        category: 'Testing',
        rating: 4.2,
        ratingCount: 95,
        downloads: 980,
        icon: 'test-gen',
        tags: ['testing', 'automation'],
        installed: true,
      },
    ],
  };

  beforeEach(() => {
    vi.clearAllMocks();
    mockFetch.mockReset();
  });

  describe('rendering', () => {
    it('should render marketplace header and search', () => {
      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: () => Promise.resolve({ skills: [] }),
      });

      render(<SkillMarketplace {...defaultProps} />);
      expect(screen.getByRole('heading', { name: /skill marketplace/i })).toBeTruthy();
      expect(screen.getByPlaceholderText('Search skills...')).toBeTruthy();
    });

    it('should render skill cards when loaded', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: () => Promise.resolve(mockSkills),
      });

      render(<SkillMarketplace {...defaultProps} />);

      await waitFor(() => {
        expect(screen.getByText('Code Review')).toBeTruthy();
      });
      expect(screen.getByText('Test Generator')).toBeTruthy();
    });

    it('should render rating stars', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: () => Promise.resolve(mockSkills),
      });

      render(<SkillMarketplace {...defaultProps} />);

      await waitFor(() => {
        const stars = screen.getAllByLabelText(/rating: 4\.5 out of 5 stars/i);
        expect(stars.length).toBeGreaterThan(0);
      });
    });

    it('should show install button for uninstalled skills', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: () => Promise.resolve(mockSkills),
      });

      render(<SkillMarketplace {...defaultProps} />);

      await waitFor(() => {
        expect(screen.getByRole('button', { name: /install code review/i })).toBeTruthy();
      });
    });

    it('should show installed badge for installed skills', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: () => Promise.resolve(mockSkills),
      });

      render(<SkillMarketplace {...defaultProps} />);

      await waitFor(() => {
        expect(screen.getByText('Installed')).toBeTruthy();
      });
    });

    it('should not render when open is false', () => {
      render(<SkillMarketplace {...defaultProps} open={false} />);
      expect(screen.queryByRole('dialog')).not.toBeInTheDocument();
    });
  });

  describe('category filter', () => {
    it('should render category filter buttons', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: () => Promise.resolve(mockSkills),
      });

      render(<SkillMarketplace {...defaultProps} />);

      await waitFor(() => {
        expect(screen.getByRole('button', { name: 'All' })).toBeTruthy();
      });
      expect(screen.getByRole('button', { name: 'Development' })).toBeTruthy();
      expect(screen.getByRole('button', { name: 'Testing' })).toBeTruthy();
    });
  });

  describe('install', () => {
    it('should call install API when install button is clicked', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: () => Promise.resolve(mockSkills),
      });

      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: () => Promise.resolve({ installed: ['skill-code-review'] }),
      });

      render(<SkillMarketplace {...defaultProps} />);

      await waitFor(() => {
        const installBtn = screen.getByRole('button', { name: /install code review/i });
        fireEvent.click(installBtn);
      });

      await waitFor(() => {
        expect(mockFetch).toHaveBeenCalledWith(
          '/api/v1/skills/install',
          expect.objectContaining({
            method: 'POST',
            body: JSON.stringify({ skillId: 'skill-code-review' }),
          })
        );
      });
    });
  });

  describe('error handling', () => {
    it('should show error message when fetch fails', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: false,
        status: 500,
      });

      render(<SkillMarketplace {...defaultProps} />);

      await waitFor(() => {
        expect(screen.getByRole('alert')).toBeTruthy();
      });
    });
  });
});

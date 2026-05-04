import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { AccessibilityToolbar } from '../ui/AccessibilityToolbar';

describe('AccessibilityToolbar', () => {
  const defaultProps = {
    visible: true,
    onToggle: vi.fn(),
  };

  beforeEach(() => {
    vi.clearAllMocks();
    sessionStorage.clear();
    document.documentElement.classList.remove('high-contrast');
    document.documentElement.removeAttribute('dir');
    document.documentElement.style.fontSize = '';
  });

  afterEach(() => {
    document.documentElement.classList.remove('high-contrast');
    document.documentElement.removeAttribute('dir');
    document.documentElement.style.fontSize = '';
  });

  describe('rendering', () => {
    it('should render toolbar with controls when visible', () => {
      render(<AccessibilityToolbar {...defaultProps} />);
      expect(screen.getByRole('toolbar', { name: /accessibility settings/i })).toBeTruthy();
      expect(screen.getByText('High Contrast')).toBeTruthy();
      expect(screen.getByText('RTL Mode')).toBeTruthy();
      expect(screen.getByText(/font size: 14px/i)).toBeTruthy();
    });

    it('should not render when visible is false', () => {
      render(<AccessibilityToolbar {...defaultProps} visible={false} />);
      expect(screen.queryByRole('toolbar')).not.toBeInTheDocument();
    });

    it('should render toggle button', () => {
      render(<AccessibilityToolbar {...defaultProps} />);
      expect(screen.getByRole('button', { name: /toggle accessibility toolbar/i })).toBeTruthy();
    });
  });

  describe('high contrast', () => {
    it('should toggle high contrast mode', () => {
      render(<AccessibilityToolbar {...defaultProps} />);
      const toggle = screen.getByRole('switch', { name: /high contrast/i });

      fireEvent.click(toggle);
      expect(document.documentElement.classList.contains('high-contrast')).toBe(true);
    });

    it('should persist high contrast to session storage', () => {
      render(<AccessibilityToolbar {...defaultProps} />);
      const toggle = screen.getByRole('switch', { name: /high contrast/i });

      fireEvent.click(toggle);
      expect(sessionStorage.getItem('accessibility-high-contrast')).toBe('true');
    });
  });

  describe('RTL mode', () => {
    it('should toggle RTL mode', () => {
      render(<AccessibilityToolbar {...defaultProps} />);
      const toggle = screen.getByRole('switch', { name: /rtl mode/i });

      fireEvent.click(toggle);
      expect(document.documentElement.dir).toBe('rtl');
    });

    it('should persist RTL to session storage', () => {
      render(<AccessibilityToolbar {...defaultProps} />);
      const toggle = screen.getByRole('switch', { name: /rtl mode/i });

      fireEvent.click(toggle);
      expect(sessionStorage.getItem('accessibility-rtl')).toBe('true');
    });
  });

  describe('font size', () => {
    it('should render font size slider', () => {
      render(<AccessibilityToolbar {...defaultProps} />);
      const slider = screen.getByRole('slider', { name: /font size/i });
      expect(slider).toBeTruthy();
      expect(slider).toHaveAttribute('min', '12');
      expect(slider).toHaveAttribute('max', '24');
    });

    it('should update font size when slider changes', () => {
      render(<AccessibilityToolbar {...defaultProps} />);
      const slider = screen.getByRole('slider', { name: /font size/i });

      fireEvent.change(slider, { target: { value: '18' } });
      expect(document.documentElement.style.fontSize).toBe('18px');
      expect(screen.getByText(/font size: 18px/i)).toBeTruthy();
    });

    it('should persist font size to session storage', () => {
      render(<AccessibilityToolbar {...defaultProps} />);
      const slider = screen.getByRole('slider', { name: /font size/i });

      fireEvent.change(slider, { target: { value: '18' } });
      expect(sessionStorage.getItem('accessibility-font-size')).toBe('18');
    });
  });

  describe('session storage', () => {
    it('should load preferences from session storage on mount', () => {
      sessionStorage.setItem('accessibility-high-contrast', 'true');
      sessionStorage.setItem('accessibility-rtl', 'true');
      sessionStorage.setItem('accessibility-font-size', '16');

      render(<AccessibilityToolbar {...defaultProps} />);

      expect(document.documentElement.classList.contains('high-contrast')).toBe(true);
      expect(document.documentElement.dir).toBe('rtl');
      expect(document.documentElement.style.fontSize).toBe('16px');
    });
  });
});

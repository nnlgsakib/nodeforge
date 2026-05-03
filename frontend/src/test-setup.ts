// Global test setup for jsdom environment
import '@testing-library/jest-dom/vitest';
import { vi } from 'vitest';

// Mock URL methods not implemented in jsdom
if (typeof window !== 'undefined') {
  (window.URL as any).createObjectURL = () => 'blob:test';
  (window.URL as any).revokeObjectURL = () => {};
  // Mock scrollIntoView (not implemented in jsdom)
  Element.prototype.scrollIntoView = vi.fn();
}

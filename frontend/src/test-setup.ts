// Global test setup for jsdom environment
// Mock URL methods not implemented in jsdom
if (typeof window !== 'undefined') {
  (window.URL as any).createObjectURL = () => 'blob:test';
  (window.URL as any).revokeObjectURL = () => {};
}

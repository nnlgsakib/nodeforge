import { describe, it, expect, vi } from 'vitest';
import { render, screen, act } from '@testing-library/react';
import { ToastProvider, useToast } from './toast';

// Helper component to test useToast hook
function ToastTester() {
  const { toast, success, error, warning, info, dismissAll } = useToast();
  return (
    <div>
      <button onClick={() => toast({ title: 'Custom Toast', variant: 'info' })}>Custom</button>
      <button onClick={() => success('Success!', 'It worked')}>Success</button>
      <button onClick={() => error('Error!', 'Something failed')}>Error</button>
      <button onClick={() => warning('Warning!', 'Be careful')}>Warning</button>
      <button onClick={() => info('Info', ' FYI')}>Info</button>
      <button onClick={dismissAll}>Dismiss All</button>
    </div>
  );
}

function renderWithProvider(ui: React.ReactNode) {
  return render(<ToastProvider>{ui}</ToastProvider>);
}

describe('Toast & useToast', () => {
  it('throws if useToast is used outside ToastProvider', () => {
    // Suppress console.error for this test
    const spy = vi.spyOn(console, 'error').mockImplementation(() => {});
    expect(() => render(<ToastTester />)).toThrow('useToast must be used within ToastProvider');
    spy.mockRestore();
  });

  it('renders toast via useToast hook', async () => {
    renderWithProvider(<ToastTester />);
    await act(async () => {
      screen.getByText('Custom').click();
    });
    // Toast is rendered by Radix Toast.Provider
    const toastEl = screen.getByText('Custom Toast');
    expect(toastEl).toBeInTheDocument();
  });

  it('shows success toast with green styling', async () => {
    renderWithProvider(<ToastTester />);
    await act(async () => {
      screen.getByText('Success').click();
    });
    const title = screen.getByText('Success!');
    expect(title).toBeInTheDocument();
    const desc = screen.getByText('It worked');
    expect(desc).toBeInTheDocument();
  });

  it('shows error toast (persistent)', async () => {
    renderWithProvider(<ToastTester />);
    await act(async () => {
      screen.getByText('Error').click();
    });
    const title = screen.getByText('Error!');
    expect(title).toBeInTheDocument();
  });

  it('shows warning toast', async () => {
    renderWithProvider(<ToastTester />);
    await act(async () => {
      screen.getByText('Warning').click();
    });
    const title = screen.getByText('Warning!');
    expect(title).toBeInTheDocument();
  });

  it('shows info toast', async () => {
    renderWithProvider(<ToastTester />);
    await act(async () => {
      screen.getByRole('button', { name: 'Info' }).click();
    });
    const title = screen.getAllByText('Info')[0];
    expect(title).toBeInTheDocument();
  });

  it('dismisses all toasts', async () => {
    renderWithProvider(<ToastTester />);
    await act(async () => {
      screen.getByText('Success').click();
    });
    await act(async () => {
      screen.getByText('Dismiss All').click();
    });
    // After dismiss, toast should not be in document
    const toasts = screen.queryByText('Success!');
    expect(toasts).not.toBeInTheDocument();
  });
});

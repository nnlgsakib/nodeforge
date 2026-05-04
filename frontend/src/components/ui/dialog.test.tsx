import { describe, it, expect, vi } from 'vitest';
import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { Dialog, AlertDialog } from './dialog';

describe('Dialog', () => {
  it('renders nothing when closed', () => {
    render(
      <Dialog open={false} onOpenChange={() => {}} title="Test Dialog">
        <p>Dialog content</p>
      </Dialog>
    );
    expect(screen.queryByText('Dialog content')).not.toBeInTheDocument();
  });

  it('renders dialog content when open', () => {
    render(
      <Dialog open={true} onOpenChange={() => {}} title="Test Dialog" description="Test description">
        <p>Dialog content</p>
      </Dialog>
    );
    expect(screen.getByText('Test Dialog')).toBeInTheDocument();
    expect(screen.getByText('Dialog content')).toBeInTheDocument();
  });

  it('closes dialog when close button is clicked', async () => {
    const onOpenChange = vi.fn();
    render(
      <Dialog open={true} onOpenChange={onOpenChange} title="Test Dialog">
        <p>Dialog content</p>
      </Dialog>
    );
    const closeBtn = screen.getByRole('button', { name: /close dialog/i });
    await userEvent.click(closeBtn);
    expect(onOpenChange).toHaveBeenCalledWith(false);
  });

  it('applies custom className', () => {
    render(
      <Dialog open={true} onOpenChange={() => {}} title="Test" className="custom-class">
        <p>Content</p>
      </Dialog>
    );
    const content = screen.getByText('Content').closest('[role="dialog"]') ||
                    screen.getByText('Content').parentElement?.parentElement;
    expect(content?.className).toContain('custom-class');
  });
});

describe('AlertDialog', () => {
  it('renders nothing when closed', () => {
    render(
      <AlertDialog
        open={false}
        onOpenChange={() => {}}
        title="Confirm Delete"
        description="Are you sure?"
        onConfirm={() => {}}
      />
    );
    expect(screen.queryByText('Confirm Delete')).not.toBeInTheDocument();
  });

  it('renders alert dialog when open', () => {
    render(
      <AlertDialog
        open={true}
        onOpenChange={() => {}}
        title="Confirm Delete"
        description="This action cannot be undone"
        onConfirm={() => {}}
      />
    );
    expect(screen.getByText('Confirm Delete')).toBeInTheDocument();
    expect(screen.getByText('This action cannot be undone')).toBeInTheDocument();
    expect(screen.getByText('Cancel')).toBeInTheDocument();
    expect(screen.getByText('Confirm')).toBeInTheDocument();
  });

  it('calls onConfirm when confirm button is clicked', async () => {
    const onConfirm = vi.fn();
    render(
      <AlertDialog
        open={true}
        onOpenChange={() => {}}
        title="Confirm"
        description="Proceed?"
        onConfirm={onConfirm}
      />
    );
    await userEvent.click(screen.getByRole('button', { name: 'Confirm' }));
    expect(onConfirm).toHaveBeenCalledTimes(1);
  });

  it('calls onOpenChange(false) when cancel is clicked', async () => {
    const onOpenChange = vi.fn();
    render(
      <AlertDialog
        open={true}
        onOpenChange={onOpenChange}
        title="Confirm"
        description="Proceed?"
        onConfirm={() => {}}
      />
    );
    await userEvent.click(screen.getByText('Cancel'));
    expect(onOpenChange).toHaveBeenCalledWith(false);
  });

  it('uses custom labels', () => {
    render(
      <AlertDialog
        open={true}
        onOpenChange={() => {}}
        title="Confirm"
        description="Proceed?"
        onConfirm={() => {}}
        cancelLabel="Go back"
        confirmLabel="Proceed"
      />
    );
    expect(screen.getByText('Go back')).toBeInTheDocument();
    expect(screen.getByText('Proceed')).toBeInTheDocument();
  });

  it('has role="alertdialog" for accessibility', () => {
    render(
      <AlertDialog
        open={true}
        onOpenChange={() => {}}
        title="Confirm"
        description="Proceed?"
        onConfirm={() => {}}
      />
    );
    const alert = screen.getByRole('alertdialog');
    expect(alert).toBeInTheDocument();
    expect(alert).toHaveAttribute('aria-labelledby');
    expect(alert).toHaveAttribute('aria-describedby');
  });
});

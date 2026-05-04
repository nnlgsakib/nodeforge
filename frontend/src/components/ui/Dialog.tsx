import React, { forwardRef } from 'react';
import * as RadixDialog from '@radix-ui/react-dialog';
import * as RadixAlertDialog from '@radix-ui/react-alert-dialog';
import { cn } from '../../utils/cn';

// -- Dialog --

interface DialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  title?: string;
  description?: string;
  children: React.ReactNode;
  className?: string;
}

export const Dialog = forwardRef<HTMLDivElement, DialogProps>(
  ({ open, onOpenChange, title, description, children, className }, ref) => {
    return (
      <RadixDialog.Root open={open} onOpenChange={onOpenChange}>
        <RadixDialog.Portal>
          <RadixDialog.Overlay className="fixed inset-0 bg-black/60 z-[1999] data-[state=open]:animate-fade-in" />
          <RadixDialog.Content
            ref={ref}
            className={cn(
              'fixed z-[2000] top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2',
              'bg-canvas-secondary border border-canvas-tertiary rounded-xl p-6',
              'w-11/12 max-w-2xl max-h-[85vh] overflow-y-auto',
              'data-[state=open]:animate-slide-in',
              'focus-visible:outline-2 focus-visible:outline-cyan-500 focus-visible:outline-offset-2',
              className
            )}
          >
            {title && (
              <RadixDialog.Title className="text-h2 font-semibold text-text-primary mb-2">
                {title}
              </RadixDialog.Title>
            )}
            {description && (
              <RadixDialog.Description className="text-body text-text-secondary mb-4">
                {description}
              </RadixDialog.Description>
            )}
            {children}
            <RadixDialog.Close asChild>
              <button
                className="absolute top-4 right-4 text-text-secondary hover:text-text-primary cursor-pointer transition-colors duration-200"
                aria-label="Close dialog"
              >
                &times;
              </button>
            </RadixDialog.Close>
          </RadixDialog.Content>
        </RadixDialog.Portal>
      </RadixDialog.Root>
    );
  }
);

Dialog.displayName = 'Dialog';

// -- AlertDialog --

interface AlertDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  title: string;
  description: string;
  onConfirm: () => void;
  cancelLabel?: string;
  confirmLabel?: string;
  destructive?: boolean;
  isLoading?: boolean;
}

export const AlertDialog: React.FC<AlertDialogProps> = ({
  open,
  onOpenChange,
  title,
  description,
  onConfirm,
  cancelLabel = 'Cancel',
  confirmLabel = 'Confirm',
  destructive = true,
  isLoading = false,
}) => {
  return (
    <RadixAlertDialog.Root open={open} onOpenChange={onOpenChange}>
      <RadixAlertDialog.Portal>
        <RadixAlertDialog.Overlay className="fixed inset-0 bg-black/60 z-[2999] data-[state=open]:animate-fade-in" />
        <RadixAlertDialog.Content
          className={cn(
            'fixed z-[3000] top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2',
            'bg-canvas-secondary border border-canvas-tertiary rounded-xl p-6',
            'w-11/12 max-w-md',
            'data-[state=open]:animate-slide-in'
          )}
          role="alertdialog"
          aria-labelledby="alert-dialog-title"
          aria-describedby="alert-dialog-desc"
        >
          <RadixAlertDialog.Title
            id="alert-dialog-title"
            className="text-h2 font-semibold text-text-primary mb-2"
          >
            {title}
          </RadixAlertDialog.Title>
          <RadixAlertDialog.Description
            id="alert-dialog-desc"
            className="text-body text-text-secondary mb-6"
          >
            {description}
          </RadixAlertDialog.Description>
          <div className="flex justify-end gap-3">
            <RadixAlertDialog.Cancel asChild>
              <button
                className={cn(
                  'inline-flex items-center justify-center font-medium rounded-lg',
                  'px-4 py-2 transition-colors duration-200 cursor-pointer',
                  'border border-gray-600 text-gray-200 hover:bg-canvas-tertiary',
                  'focus-visible:outline-2 focus-visible:outline-cyan-500 focus-visible:outline-offset-2'
                )}
              >
                {cancelLabel}
              </button>
            </RadixAlertDialog.Cancel>
            <RadixAlertDialog.Action asChild>
              <button
                onClick={onConfirm}
                disabled={isLoading}
                className={cn(
                  'inline-flex items-center justify-center font-medium rounded-lg',
                  'px-4 py-2 transition-colors duration-200 cursor-pointer',
                  'focus-visible:outline-2 focus-visible:outline-offset-2',
                  destructive
                    ? 'bg-[#ef4444] text-white hover:bg-red-600 focus-visible:outline-red-500'
                    : 'bg-[#06b6d4] text-white hover:bg-cyan-600 focus-visible:outline-cyan-500',
                  isLoading && 'opacity-50 cursor-not-allowed'
                )}
              >
                {isLoading ? 'Confirming...' : confirmLabel}
              </button>
            </RadixAlertDialog.Action>
          </div>
        </RadixAlertDialog.Content>
      </RadixAlertDialog.Portal>
    </RadixAlertDialog.Root>
  );
};

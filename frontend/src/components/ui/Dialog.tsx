import React, { forwardRef } from 'react';
import * as RadixDialog from '@radix-ui/react-dialog';

interface DialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  title?: string;
  children: React.ReactNode;
}

export const Dialog = forwardRef<HTMLDivElement, DialogProps>(
  ({ open, onOpenChange, title, children }, ref) => {
    return (
      <RadixDialog.Root open={open} onOpenChange={onOpenChange}>
        <RadixDialog.Portal>
          <RadixDialog.Overlay className="fixed inset-0 bg-black/60 z-[1999]" />
          <RadixDialog.Content
            ref={ref}
            className="fixed z-[2000] top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 bg-canvas-secondary border border-canvas-tertiary rounded-xl p-6 w-11/12 max-w-2xl max-h-[85vh] overflow-y-auto"
          >
            {title && (
              <RadixDialog.Title className="text-h2 font-semibold text-text-primary mb-4">
                {title}
              </RadixDialog.Title>
            )}
            {children}
            <RadixDialog.Close asChild>
              <button
                className="absolute top-4 right-4 text-text-secondary hover:text-text-primary cursor-pointer transition-colors duration-200"
                aria-label="Close"
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

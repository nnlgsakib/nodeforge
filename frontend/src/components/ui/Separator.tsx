import { forwardRef } from 'react';
import * as RadixSeparator from '@radix-ui/react-separator';

interface SeparatorProps {
  orientation?: 'horizontal' | 'vertical';
  className?: string;
}

export const Separator = forwardRef<HTMLDivElement, SeparatorProps>(
  ({ orientation = 'horizontal', className = '' }, ref) => {
    return (
      <RadixSeparator.Root
        ref={ref}
        orientation={orientation}
        className={`bg-canvas-tertiary ${orientation === 'horizontal' ? 'h-px w-full' : 'w-px h-full'} ${className}`}
      />
    );
  }
);

Separator.displayName = 'Separator';

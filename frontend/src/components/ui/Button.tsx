import React, { forwardRef } from 'react';
import * as RadixSlot from '@radix-ui/react-slot';
import { cn } from '../../utils/cn';

export type ButtonVariant = 'default' | 'outline' | 'destructive' | 'ghost' | 'icon';

export interface ButtonProps extends React.ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: ButtonVariant;
  asChild?: boolean;
}

export const Button = forwardRef<HTMLButtonElement, ButtonProps>(
  ({ variant = 'default', asChild = false, className = '', type = 'button', disabled, ...props }, ref) => {
    const Comp = asChild ? RadixSlot.Slot : 'button';

    const baseClasses =
      'inline-flex items-center justify-center font-medium rounded-lg transition-colors duration-200 cursor-pointer focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-cyan-500 disabled:opacity-50 disabled:cursor-not-allowed';

    const variantClasses: Record<ButtonVariant, string> = {
      default: 'bg-[#06b6d4] text-white hover:bg-cyan-600 active:bg-cyan-700 px-4 py-2',
      outline: 'border border-gray-600 text-gray-200 hover:bg-canvas-tertiary active:bg-canvas-secondary px-4 py-2',
      destructive: 'bg-[#ef4444] text-white hover:bg-red-600 active:bg-red-700 px-4 py-2',
      ghost: 'text-gray-200 hover:bg-canvas-tertiary active:bg-canvas-secondary px-4 py-2',
      icon: 'w-8 h-8 p-0 text-gray-200 hover:bg-canvas-tertiary active:bg-canvas-secondary rounded-lg',
    };

    return (
      <Comp
        ref={ref}
        type={type}
        className={cn(baseClasses, variantClasses[variant], className)}
        disabled={disabled}
        {...props}
      />
    );
  }
);

Button.displayName = 'Button';

import React, { forwardRef } from 'react';

export type ButtonVariant = 'primary' | 'secondary' | 'danger' | 'icon';

interface ButtonProps extends React.ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: ButtonVariant;
  children: React.ReactNode;
}

export const Button = forwardRef<HTMLButtonElement, ButtonProps>(
  ({ variant = 'primary', children, className = '', type = 'button', ...props }, ref) => {
    const baseClasses =
      'inline-flex items-center justify-center font-medium transition-colors duration-200 cursor-pointer disabled:opacity-50 disabled:cursor-not-allowed';

    const variantClasses: Record<ButtonVariant, string> = {
      primary: 'bg-primary text-white hover:bg-cyan-600',
      secondary: 'border border-secondary text-primary hover:bg-canvas-tertiary',
      danger: 'bg-danger text-white hover:bg-red-600',
      icon: 'w-8 h-8 p-0 bg-transparent hover:bg-canvas-tertiary',
    };

    const sizeClasses =
      variant === 'icon'
        ? ''
        : 'px-4 py-2 text-body rounded-md';

    const iconClasses = variant === 'icon' ? 'rounded-md' : '';

    return (
      <button
        ref={ref}
        type={type}
        className={`${baseClasses} ${variantClasses[variant]} ${sizeClasses} ${iconClasses} ${className}`}
        {...props}
      >
        {children}
      </button>
    );
  }
);

Button.displayName = 'Button';

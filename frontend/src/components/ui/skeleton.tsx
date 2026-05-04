import React from 'react';
import { cn } from '../../utils/cn';

interface SkeletonProps extends React.HTMLAttributes<HTMLDivElement> {
  /** Width shorthand, e.g. "w-32" or "100%" */
  width?: string;
  /** Height shorthand, e.g. "h-4" or "20px" */
  height?: string;
  /** Whether to show the pulse animation (60% opacity) */
  animated?: boolean;
  /** Roundness: "none" | "sm" | "md" | "full" */
  rounded?: 'none' | 'sm' | 'md' | 'full';
}

const roundedMap: Record<string, string> = {
  none: 'rounded-none',
  sm: 'rounded-sm',
  md: 'rounded-md',
  full: 'rounded-full',
};

export const Skeleton: React.FC<SkeletonProps> = ({
  className = '',
  width,
  height,
  animated = true,
  rounded = 'md',
  style,
  ...props
}) => {
  const sizeStyle = cn(width, height);
  const pulseClass = animated ? 'animate-pulse opacity-60' : 'opacity-60';

  return (
    <div
      className={cn(
        'bg-canvas-tertiary',
        roundedMap[rounded],
        pulseClass,
        sizeStyle,
        className
      )}
      role="status"
      aria-label="Loading"
      style={style}
      {...props}
    />
  );
};

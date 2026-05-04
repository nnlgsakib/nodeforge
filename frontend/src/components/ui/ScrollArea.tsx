import React, { forwardRef } from 'react';
import * as RadixScrollArea from '@radix-ui/react-scroll-area';

interface ScrollAreaProps {
  children: React.ReactNode;
  className?: string;
  orientation?: 'vertical' | 'horizontal' | 'both';
}

export const ScrollArea = forwardRef<HTMLDivElement, ScrollAreaProps>(
  ({ children, className = '', orientation = 'vertical' }, ref) => {
    const scrollbars =
      orientation === 'both'
        ? (
            <>
              <VerticalScrollbar />
              <HorizontalScrollbar />
            </>
          )
        : orientation === 'horizontal'
          ? <HorizontalScrollbar />
          : <VerticalScrollbar />;

    return (
      <RadixScrollArea.Root ref={ref} className={`overflow-hidden ${className}`}>
        <RadixScrollArea.Viewport className="w-full h-full">
          {children}
        </RadixScrollArea.Viewport>
        {scrollbars}
      </RadixScrollArea.Root>
    );
  }
);

ScrollArea.displayName = 'ScrollArea';

const VerticalScrollbar = () => (
  <RadixScrollArea.Scrollbar
    className="flex select-none touch-none p-0.5 bg-canvas-tertiary data-[orientation=vertical]:w-2.5"
    orientation="vertical"
  >
    <RadixScrollArea.Thumb className="flex-1 rounded-[10px] relative before:content-[''] before:absolute before:top-1/2 before:left-1/2 before:-translate-x-1/2 before:-translate-y-1/2 before:w-full before:min-w-[44px] before:h-full before:min-h-[44px] before:bg-canvas-primary/50 before:rounded-[10px]" />
  </RadixScrollArea.Scrollbar>
);

const HorizontalScrollbar = () => (
  <RadixScrollArea.Scrollbar
    className="flex select-none touch-none p-0.5 bg-canvas-tertiary data-[orientation=horizontal]:h-2.5"
    orientation="horizontal"
  >
    <RadixScrollArea.Thumb className="flex-1 rounded-[10px] relative before:content-[''] before:absolute before:top-1/2 before:left-1/2 before:-translate-x-1/2 before:-translate-y-1/2 before:w-full before:min-w-[44px] before:h-full before:min-h-[44px] before:bg-canvas-primary/50 before:rounded-[10px]" />
  </RadixScrollArea.Scrollbar>
);

import React, { isValidElement } from 'react';
import * as RadixTooltip from '@radix-ui/react-tooltip';

interface TooltipProps {
  content: string;
  children: React.ReactNode;
  side?: 'top' | 'right' | 'bottom' | 'left';
  delayDuration?: number;
}

export const Tooltip: React.FC<TooltipProps> = ({
  content,
  children,
  side = 'top',
  delayDuration = 200,
}) => {
  const trigger = isValidElement(children) && React.Children.count(children) === 1
    ? children
    : <span tabIndex={0}>{children}</span>;

  return (
    <RadixTooltip.Root delayDuration={delayDuration}>
      <RadixTooltip.Trigger asChild>
        {trigger}
      </RadixTooltip.Trigger>
      <RadixTooltip.Portal>
        <RadixTooltip.Content
          className="bg-canvas-tertiary text-text-primary text-tiny px-2 py-1 rounded shadow-lg"
          side={side}
          sideOffset={4}
        >
          {content}
          <RadixTooltip.Arrow className="fill-canvas-tertiary" />
        </RadixTooltip.Content>
      </RadixTooltip.Portal>
    </RadixTooltip.Root>
  );
};

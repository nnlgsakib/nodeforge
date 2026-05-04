import { useId, forwardRef } from 'react';
import * as RadixSwitch from '@radix-ui/react-switch';

interface SwitchProps {
  checked?: boolean;
  onCheckedChange?: (checked: boolean) => void;
  id?: string;
  label?: string;
}

export const Switch = forwardRef<HTMLButtonElement, SwitchProps>(
  ({ checked, onCheckedChange, id, label }, ref) => {
    const autoId = useId();
    const switchId = id ?? (label ? autoId : undefined);

    return (
      <div className="flex items-center gap-2">
        {label && (
          <label htmlFor={switchId} className="text-body font-medium text-text-primary cursor-pointer">
            {label}
          </label>
        )}
        <RadixSwitch.Root
          id={switchId}
          ref={ref}
          checked={checked}
          onCheckedChange={onCheckedChange}
          className="w-[48px] h-[26px] rounded-[13px] bg-canvas-primary border border-canvas-tertiary relative cursor-pointer transition-colors duration-200 data-[state=checked]:bg-primary data-[state=checked]:border-primary"
        >
          <RadixSwitch.Thumb className="block w-[20px] h-[20px] rounded-full bg-white translate-x-[2px] transition-transform duration-200 data-[state=checked]:translate-x-[22px]" />
        </RadixSwitch.Root>
      </div>
    );
  }
);

Switch.displayName = 'Switch';

import React, { createContext, useContext, useState, useCallback } from 'react';
import * as RadixToast from '@radix-ui/react-toast';
import { cn } from '../../utils/cn';

export type ToastVariant = 'success' | 'destructive' | 'warning' | 'info';

export interface ToastOptions {
  title?: string;
  description?: string;
  variant?: ToastVariant;
  duration?: number; // ms, 0 = persistent
}

interface ToastData extends ToastOptions {
  id: string;
}

interface ToastContextValue {
  toasts: ToastData[];
  toast: (options: ToastOptions) => void;
  dismiss: (id: string) => void;
  dismissAll: () => void;
}

const ToastContext = createContext<ToastContextValue | undefined>(undefined);

const variantStyles: Record<ToastVariant, string> = {
  success: 'bg-green-900/90 border-green-700 text-green-100',
  destructive: 'bg-red-900/90 border-red-700 text-red-100',
  warning: 'bg-yellow-900/90 border-yellow-700 text-yellow-100',
  info: 'bg-cyan-900/90 border-cyan-700 text-cyan-100 [animation:edge-dash_2s_linear_infinite]',
};

const variantIndicators: Record<ToastVariant, React.ReactNode> = {
  success: null,
  destructive: null,
  warning: (
    <span className="inline-flex items-center px-1.5 py-0.5 rounded text-[10px] font-medium bg-yellow-700/50 text-yellow-200 uppercase">
      Paused
    </span>
  ),
  info: (
    <span className="inline-block w-2 h-2 rounded-full bg-cyan-400 animate-pulse" />
  ),
};

export const ToastProvider: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const [toasts, setToasts] = useState<ToastData[]>([]);
  const nextIdRef = React.useRef(0);

  const toast = useCallback((options: ToastOptions) => {
    const id = `toast-${nextIdRef.current++}-${Math.random().toString(36).slice(2, 9)}`;
    setToasts((prev) => [...prev, { id, ...options }]);
  }, []);

  const dismiss = useCallback((id: string) => {
    setToasts((prev) => prev.filter((t) => t.id !== id));
  }, []);

  const dismissAll = useCallback(() => {
    setToasts([]);
  }, []);

  return (
    <ToastContext.Provider value={{ toasts, toast, dismiss, dismissAll }}>
      <RadixToast.Provider
        swipeDirection="right"
        label="Notifications"
      >
        {children}
        {toasts.map((t) => {
          const duration = t.duration !== undefined ? t.duration : (t.variant === 'destructive' ? 0 : 5000);
          return (
            <RadixToast.Root
              key={t.id}
              className={cn(
                'border rounded-lg p-4 shadow-lg max-w-md',
                'data-[state=open]:animate-slide-in',
                'data-[state=closed]:animate-fade-out',
                variantStyles[t.variant || 'info']
              )}
              duration={duration}
              onOpenChange={(open) => { if (!open) dismiss(t.id); }}
            >
              <div className="flex items-start gap-3">
                {variantIndicators[t.variant || 'info']}
                <div className="flex-1">
                  {t.title && (
                    <RadixToast.Title className="font-semibold text-sm">
                      {t.title}
                    </RadixToast.Title>
                  )}
                  {t.description && (
                    <RadixToast.Description className="text-xs mt-1 opacity-90">
                      {t.description}
                    </RadixToast.Description>
                  )}
                </div>
                <RadixToast.Close
                  className="text-xs opacity-70 hover:opacity-100 transition-opacity cursor-pointer"
                  aria-label="Dismiss"
                >
                  &times;
                </RadixToast.Close>
              </div>
            </RadixToast.Root>
          );
        })}
        {toasts.length > 1 && (
          <button
            onClick={dismissAll}
            className="text-xs text-gray-400 hover:text-gray-200 transition-colors cursor-pointer self-end mt-1"
            aria-label="Dismiss all notifications"
          >
            Dismiss all
          </button>
        )}
        <RadixToast.Viewport className="fixed bottom-0 right-0 p-4 flex flex-col gap-2 max-w-md" />
      </RadixToast.Provider>
    </ToastContext.Provider>
  );
};

export function useToast() {
  const ctx = useContext(ToastContext);
  if (!ctx) {
    throw new Error('useToast must be used within ToastProvider');
  }
  return {
    toast: ctx.toast,
    dismiss: ctx.dismiss,
    dismissAll: ctx.dismissAll,
    success: (title: string, description?: string) =>
      ctx.toast({ title, description, variant: 'success', duration: 3000 }),
    error: (title: string, description?: string) =>
      ctx.toast({ title, description, variant: 'destructive', duration: 0 }),
    warning: (title: string, description?: string) =>
      ctx.toast({ title, description, variant: 'warning', duration: 5000 }),
    info: (title: string, description?: string) =>
      ctx.toast({ title, description, variant: 'info', duration: 5000 }),
  };
}

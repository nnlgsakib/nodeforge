import { useContext, useCallback, useRef } from 'react';
import { AriaAnnouncerContext } from '../components/ui/aria-announcer';

type AnnouncementType = 'polite' | 'assertive';

/**
 * Hook for queueing screen reader announcements to ARIA live regions (AC:3, Subtask 3.2)
 * Usage: const announce = useAnnounce(); announce('Node changed to running');
 */
export function useAnnounce() {
  const { announce } = useContext(AriaAnnouncerContext);
  const announceRef = useRef(announce);

  // Keep ref in sync
  if (announce !== announceRef.current) {
    announceRef.current = announce;
  }

  return useCallback((message: string, type: AnnouncementType = 'polite') => {
    announceRef.current(message, type);
  }, []);
}

import React, { useState, useCallback, useEffect } from 'react';

type AnnouncementType = 'polite' | 'assertive';

interface Announcement {
  id: number;
  message: string;
  type: AnnouncementType;
  timestamp: number;
}

interface ThemeContextValue {
  isHighContrast: boolean;
}

interface AriaAnnouncerContextValue {
  announce: (message: string, type?: AnnouncementType) => void;
  theme: ThemeContextValue;
}

export const AriaAnnouncerContext = React.createContext<AriaAnnouncerContextValue>({
  announce: () => {},
  theme: { isHighContrast: false },
});

interface AriaAnnouncerProps {
  children: React.ReactNode;
}

let announcementId = 0;

/**
 * ARIA live region component with polite/assertive queues (AC:3, Subtask 3.1)
 * Renders off-screen divs that screen readers will announce.
 * Also tracks high-contrast mode state for reactive node/edge rendering.
 */
export function AriaAnnouncer({ children }: AriaAnnouncerProps) {
  const [announcements, setAnnouncements] = useState<Announcement[]>([]);
  const [isHighContrast, setIsHighContrast] = useState(() =>
    document.documentElement.classList.contains('high-contrast')
  );

  // Listen for high-contrast toggle via MutationObserver
  useEffect(() => {
    const observer = new MutationObserver(() => {
      setIsHighContrast(document.documentElement.classList.contains('high-contrast'));
    });
    observer.observe(document.documentElement, { attributes: true, attributeFilter: ['class'] });
    return () => observer.disconnect();
  }, []);

  const announce = useCallback((message: string, type: AnnouncementType = 'polite') => {
    announcementId += 1;
    setAnnouncements((prev) => [...prev, { id: announcementId, message, type, timestamp: Date.now() }]);
  }, []);

  // Clean up old announcements after they've been announced (time-based expiry)
  useEffect(() => {
    if (announcements.length === 0) return;
    const cutoff = Date.now() - 5000; // 5 second expiry
    const active = announcements.filter((a) => a.timestamp > cutoff);
    if (active.length < announcements.length) {
      setAnnouncements(active);
    }
  }, [announcements]);

  const politeMessages = announcements.filter((a) => a.type === 'polite');
  const assertiveMessages = announcements.filter((a) => a.type === 'assertive');

  return (
    <AriaAnnouncerContext.Provider value={{ announce, theme: { isHighContrast } }}>
      {children}
      <div aria-live="polite" aria-atomic="false" className="sr-only" id="aria-polite-announcer">
        {politeMessages.map((a) => (
          <div key={a.id} role="status">{a.message}</div>
        ))}
      </div>
      <div aria-live="assertive" aria-atomic="false" className="sr-only" id="aria-assertive-announcer">
        {assertiveMessages.map((a) => (
          <div key={a.id} role="alert">{a.message}</div>
        ))}
      </div>
    </AriaAnnouncerContext.Provider>
  );
}

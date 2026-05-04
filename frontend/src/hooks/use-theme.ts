import { useContext } from 'react';
import { AriaAnnouncerContext } from '../components/ui/aria-announcer';

/**
 * Hook for accessing theme state (high-contrast) from AriaAnnouncerContext.
 * Ensures components re-render when theme toggles.
 */
export function useTheme() {
  const { theme } = useContext(AriaAnnouncerContext);
  return theme;
}

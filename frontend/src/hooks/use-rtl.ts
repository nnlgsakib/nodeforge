import { useEffect, useState } from 'react';

/**
 * Hook for detecting RTL mode changes via MutationObserver on document.documentElement.
 * Returns true when dir="rtl" is set.
 */
export function useRtl(): boolean {
  const [isRtl, setIsRtl] = useState(() =>
    document.documentElement.dir === 'rtl'
  );

  useEffect(() => {
    const observer = new MutationObserver(() => {
      setIsRtl(document.documentElement.dir === 'rtl');
    });
    observer.observe(document.documentElement, { attributes: true, attributeFilter: ['dir'] });
    return () => observer.disconnect();
  }, []);

  return isRtl;
}

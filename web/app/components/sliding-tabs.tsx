import { useLayoutEffect, useRef, useState, type CSSProperties } from 'react';

// useLayoutEffect warns when the page is prerendered; this is the same thing
// without the warning.
const useIsoLayoutEffect = typeof window === 'undefined' ? () => {} : useLayoutEffect;

/**
 * Measures the active tab so a highlight can slide to it. Returns a ref
 * callback for each tab and the style for the highlight.
 */
export function useSlidingIndicator(active: string) {
  const tabs = useRef(new Map<string, HTMLElement>());
  const [style, setStyle] = useState<CSSProperties>({ opacity: 0 });

  useIsoLayoutEffect(() => {
    const measure = () => {
      const el = tabs.current.get(active);
      if (!el) return;
      setStyle({ width: el.offsetWidth, transform: `translateX(${el.offsetLeft}px)`, opacity: 1 });
    };
    measure();
    window.addEventListener('resize', measure);
    return () => window.removeEventListener('resize', measure);
  }, [active]);

  const register = (id: string) => (el: HTMLElement | null) => {
    if (el) tabs.current.set(id, el);
    else tabs.current.delete(id);
  };

  return { register, style };
}

/** Classes for the sliding highlight. Combine with a background or border. */
export const indicatorClass =
  'pointer-events-none absolute left-0 transition-[transform,width,opacity] duration-200 ease-[cubic-bezier(0.23,1,0.32,1)] motion-reduce:transition-none';

import { useEffect, useRef, useState } from 'react';

export function useWebComponentScript(src) {
  const [loaded, setLoaded] = useState(false);
  const [error, setError] = useState(null);

  useEffect(() => {
    if (document.querySelector(`script[src="${src}"]`)) {
      setLoaded(true);
      return;
    }
    const script = document.createElement('script');
    script.src = src;
    script.onload = () => setLoaded(true);
    script.onerror = () => setError(`Failed to load ${src}`);
    document.head.appendChild(script);
  }, [src]);

  return { loaded, error };
}

export function useCustomEvent(ref, eventName, handler) {
  useEffect(() => {
    const el = ref.current;
    if (!el) return;
    el.addEventListener(eventName, handler);
    return () => el.removeEventListener(eventName, handler);
  });
}

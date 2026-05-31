import { useEffect } from 'react';

/** Sets the browser tab title, e.g. "Records · Workout Log". */
export function useTitle(title: string | undefined) {
  useEffect(() => {
    document.title = title ? `${title} · Workout Log` : 'Workout Log';
  }, [title]);
}

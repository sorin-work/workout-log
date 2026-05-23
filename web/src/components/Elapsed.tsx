import { useEffect, useState } from 'react';
import { formatDuration } from '../lib/format';

/** Live "38 min" counter for a workout that is still going. */
export function Elapsed({ since }: { since: string }) {
  const [now, setNow] = useState(() => new Date().toISOString());
  useEffect(() => {
    const id = window.setInterval(() => setNow(new Date().toISOString()), 30_000);
    return () => window.clearInterval(id);
  }, []);
  return <span className="elapsed">{formatDuration(since, now)}</span>;
}

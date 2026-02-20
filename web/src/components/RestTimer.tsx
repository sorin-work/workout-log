import { useEffect, useState } from 'react';

interface Props {
  /** Changes every time a new rest period should start. */
  startedAt: number;
  seconds: number;
  onDone?(): void;
  onDismiss(): void;
}

export function formatClock(totalSeconds: number): string {
  const s = Math.max(0, Math.ceil(totalSeconds));
  return `${Math.floor(s / 60)}:${String(s % 60).padStart(2, '0')}`;
}

export function RestTimer({ startedAt, seconds, onDone, onDismiss }: Props) {
  const [now, setNow] = useState(() => Date.now());
  const [extra, setExtra] = useState(0);
  const end = startedAt + (seconds + extra) * 1000;
  const left = (end - now) / 1000;

  useEffect(() => {
    const id = window.setInterval(() => setNow(Date.now()), 250);
    return () => window.clearInterval(id);
  }, []);

  useEffect(() => {
    if (left > 0) return;
    onDone?.();
  }, [left <= 0]); // eslint-disable-line react-hooks/exhaustive-deps

  return (
    <div className={left > 0 ? 'rest-timer' : 'rest-timer done'} role="timer">
      <span>Rest</span>
      <strong>{left > 0 ? formatClock(left) : 'Go!'}</strong>
      <button className="btn small" onClick={() => setExtra((e) => e + 30)}>
        +30 s
      </button>
      <button className="btn small" onClick={onDismiss}>
        Skip
      </button>
    </div>
  );
}

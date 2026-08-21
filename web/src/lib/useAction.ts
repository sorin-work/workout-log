import { useCallback, useState } from 'react';
import { errorMessage } from '../api/client';

/**
 * Runs page actions (buttons that call the API) and keeps the last error.
 * `busy` is true while one is in flight so buttons can be disabled.
 */
export function useAction() {
  const [error, setError] = useState('');
  const [busy, setBusy] = useState(false);
  const run = useCallback(async (fn: () => Promise<unknown>) => {
    setError('');
    setBusy(true);
    try {
      await fn();
    } catch (err) {
      setError(errorMessage(err));
    } finally {
      setBusy(false);
    }
  }, []);
  return { run, error, busy };
}

import { useCallback, useEffect, useState } from 'react';
import { errorMessage, get } from '../api/client';

export interface FetchState<T> {
  data: T | undefined;
  error: string;
  loading: boolean;
  reload(): void;
  setData(data: T): void;
}

interface Result<T> {
  key: string;
  data?: T;
  error: string;
}

/** Loads a GET endpoint and reloads whenever the path changes. */
export function useFetch<T>(path: string | null): FetchState<T> {
  const [version, setVersion] = useState(0);
  const [result, setResult] = useState<Result<T>>({ key: '', error: '' });
  const key = path === null ? '' : `${path}#${version}`;

  useEffect(() => {
    if (path === null) return;
    let cancelled = false;
    const k = `${path}#${version}`;
    get<T>(path)
      .then((data) => {
        if (!cancelled) setResult({ key: k, data, error: '' });
      })
      .catch((err) => {
        if (!cancelled) setResult((r) => ({ key: k, data: r.data, error: errorMessage(err) }));
      });
    return () => {
      cancelled = true;
    };
  }, [path, version]);

  const reload = useCallback(() => setVersion((v) => v + 1), []);
  const setData = useCallback((data: T) => setResult((r) => ({ ...r, data })), []);
  return {
    data: result.data,
    error: result.error,
    loading: path !== null && result.key !== key,
    reload,
    setData,
  };
}

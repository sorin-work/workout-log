import { useCallback, useEffect, useMemo, useState, type ReactNode } from 'react';
import { ApiError, get, post } from '../api/client';
import type { User } from '../api/types';
import { AuthContext, type AuthState } from './context';

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<User | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    get<User>('/me')
      .then(setUser)
      .catch((err) => {
        if (!(err instanceof ApiError && err.status === 401)) console.error(err);
      })
      .finally(() => setLoading(false));
  }, []);

  const login = useCallback(async (email: string, password: string) => {
    setUser(await post<User>('/auth/login', { email, password }));
  }, []);

  const signup = useCallback(async (email: string, password: string, displayName: string) => {
    setUser(await post<User>('/auth/signup', { email, password, displayName }));
  }, []);

  const logout = useCallback(async () => {
    await post('/auth/logout');
    setUser(null);
  }, []);

  const value = useMemo<AuthState>(
    () => ({ user, loading, login, signup, logout, setUser }),
    [user, loading, login, signup, logout],
  );
  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}

import { render } from '@testing-library/react';
import type { ReactElement } from 'react';
import { MemoryRouter, Route, Routes } from 'react-router';
import { vi } from 'vitest';
import { AuthProvider } from '../auth/AuthProvider';

type Handler = (method: string, path: string, body: unknown) => [number, unknown];

/** Replaces fetch with a handler that answers /api requests. */
export function mockApi(handler: Handler) {
  const fn = vi.fn(async (input: RequestInfo | URL, init?: RequestInit) => {
    const url = typeof input === 'string' ? input : input.toString();
    const body = init?.body ? JSON.parse(String(init.body)) : undefined;
    const [status, data] = handler(init?.method ?? 'GET', url.replace(/^\/api/, ''), body);
    return new Response(status === 204 ? null : JSON.stringify(data), {
      status,
      headers: { 'Content-Type': 'application/json' },
    });
  });
  vi.stubGlobal('fetch', fn);
  return fn;
}

/** Renders a page inside the router and auth provider. */
export function renderAt(path: string, routes: Record<string, ReactElement>) {
  return render(
    <MemoryRouter initialEntries={[path]}>
      <AuthProvider>
        <Routes>
          {Object.entries(routes).map(([p, el]) => (
            <Route key={p} path={p} element={el} />
          ))}
        </Routes>
      </AuthProvider>
    </MemoryRouter>,
  );
}

export const demoUser = {
  id: 1,
  email: 'demo@example.com',
  displayName: 'Demo',
  unit: 'kg',
  timezone: 'UTC',
  createdAt: '2026-01-01T00:00:00Z',
};

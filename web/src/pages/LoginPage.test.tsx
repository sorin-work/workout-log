import { screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { afterEach, describe, expect, it, vi } from 'vitest';
import { demoUser, mockApi, renderAt } from '../test/render';
import LoginPage from './LoginPage';

afterEach(() => {
  vi.unstubAllGlobals();
});

describe('LoginPage', () => {
  it('logs in and goes to the home page', async () => {
    const fetch = mockApi((method, path) => {
      if (path === '/me') return [401, { error: 'unauthorized', message: 'Please log in.' }];
      if (method === 'POST' && path === '/auth/login') return [200, demoUser];
      return [404, {}];
    });
    const user = userEvent.setup();
    renderAt('/login', { '/login': <LoginPage />, '/': <p>home page</p> });

    await user.type(await screen.findByLabelText('Email'), 'demo@example.com');
    await user.type(screen.getByLabelText('Password'), 'demo-password');
    await user.click(screen.getByRole('button', { name: 'Log in' }));

    expect(await screen.findByText('home page')).toBeInTheDocument();
    const [, init] = fetch.mock.calls.find(([url]) => url === '/api/auth/login')!;
    expect(JSON.parse(String(init?.body))).toEqual({
      email: 'demo@example.com',
      password: 'demo-password',
    });
  });

  it('shows the error from the server', async () => {
    mockApi((_, path) =>
      path === '/me'
        ? [401, { error: 'unauthorized', message: 'Please log in.' }]
        : [401, { error: 'invalid_credentials', message: 'Email or password is incorrect.' }],
    );
    const user = userEvent.setup();
    renderAt('/login', { '/login': <LoginPage /> });
    await user.type(await screen.findByLabelText('Email'), 'demo@example.com');
    await user.type(screen.getByLabelText('Password'), 'nope');
    await user.click(screen.getByRole('button', { name: 'Log in' }));
    expect(await screen.findByText('Email or password is incorrect.')).toBeInTheDocument();
  });

  it('redirects when already logged in', async () => {
    mockApi(() => [200, demoUser]);
    renderAt('/login', { '/login': <LoginPage />, '/': <p>home page</p> });
    expect(await screen.findByText('home page')).toBeInTheDocument();
  });
});

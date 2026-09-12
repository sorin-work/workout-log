import { screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { afterEach, describe, expect, it, vi } from 'vitest';
import { demoUser, mockApi, renderAt } from '../test/render';
import WorkoutsPage from './WorkoutsPage';

afterEach(() => {
  vi.unstubAllGlobals();
});

function workout(id: number) {
  return {
    id,
    name: `Workout ${id}`,
    startedAt: `2026-08-${String(id).padStart(2, '0')}T18:00:00Z`,
    finishedAt: `2026-08-${String(id).padStart(2, '0')}T19:00:00Z`,
    exerciseCount: 1,
    setCount: 3,
    volumeKg: 1500,
    exercises: [],
  };
}

describe('WorkoutsPage', () => {
  it('pages through the history up to the last partial page', async () => {
    mockApi((_, path) => {
      if (path === '/me') return [200, demoUser];
      if (path.startsWith('/exercises')) return [200, []];
      if (path.startsWith('/stats/calendar')) return [200, []];
      const page = Number(new URLSearchParams(path.split('?')[1]).get('page'));
      const items = page === 3 ? [workout(1)] : [workout(10 - page * 2), workout(9 - page * 2)];
      return [200, { items, page, pageSize: 2, total: 5, totalPages: 3 }];
    });
    const user = userEvent.setup();
    renderAt('/workouts', { '/workouts': <WorkoutsPage /> });

    expect(await screen.findByText('Workout 8')).toBeInTheDocument();
    expect(screen.getByText('Page 1 of 3')).toBeInTheDocument();
    await user.click(screen.getByRole('button', { name: 'Older' }));
    await user.click(await screen.findByRole('button', { name: 'Older' }));
    expect(await screen.findByText('Workout 1')).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'Older' })).toBeDisabled();
  });
});

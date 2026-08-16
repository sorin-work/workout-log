import { screen, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { afterEach, describe, expect, it, vi } from 'vitest';
import { demoUser, mockApi, renderAt } from '../test/render';
import RecordsPage from './RecordsPage';

afterEach(() => {
  vi.unstubAllGlobals();
});

function rec(id: number, name: string, max: number, lastDone: string) {
  const set = { value: max, weightKg: max, reps: 1, workoutId: id, performedAt: lastDone };
  return {
    exerciseId: id,
    exerciseName: name,
    estimatedMax: set,
    heaviest: set,
    mostReps: set,
    bestVolume: set,
    repMaxes: {},
    setCount: 10,
    lastDone,
  };
}

describe('RecordsPage', () => {
  it('lists, filters and sorts records', async () => {
    mockApi((_, path) => {
      if (path === '/me') return [200, demoUser];
      return [
        200,
        [
          rec(1, 'Bench Press', 100, '2026-08-01T18:00:00Z'),
          rec(2, 'Deadlift', 180, '2026-07-20T18:00:00Z'),
          rec(3, 'Back Squat', 140, '2026-08-10T18:00:00Z'),
        ],
      ];
    });
    const user = userEvent.setup();
    renderAt('/records', { '/records': <RecordsPage /> });

    const names = () =>
      within(screen.getByRole('table'))
        .getAllByRole('link', { name: /Press|Deadlift|Squat/ })
        .map((a) => a.textContent);
    await screen.findByText('Deadlift');
    expect(names()).toEqual(['Back Squat', 'Bench Press', 'Deadlift']);

    await user.selectOptions(screen.getByLabelText('Sort records'), 'max');
    expect(names()).toEqual(['Deadlift', 'Back Squat', 'Bench Press']);

    await user.selectOptions(screen.getByLabelText('Sort records'), 'recent');
    expect(names()[0]).toBe('Back Squat');

    await user.type(screen.getByPlaceholderText('Filter exercises'), 'dead');
    expect(names()).toEqual(['Deadlift']);
  });

  it('shows an empty state', async () => {
    mockApi((_, path) => (path === '/me' ? [200, demoUser] : [200, []]));
    renderAt('/records', { '/records': <RecordsPage /> });
    expect(await screen.findByText(/your records will show up here/)).toBeInTheDocument();
  });
});

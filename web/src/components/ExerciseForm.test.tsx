import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { afterEach, describe, expect, it, vi } from 'vitest';
import { mockApi } from '../test/render';
import { ExerciseForm } from './ExerciseForm';

afterEach(() => {
  vi.unstubAllGlobals();
});

describe('ExerciseForm', () => {
  it('creates an exercise without listing the primary muscle twice', async () => {
    const fetch = mockApi((_, __, body) => [201, { id: 9, custom: true, ...(body as object) }]);
    const onSaved = vi.fn();
    const user = userEvent.setup();
    render(<ExerciseForm onSaved={onSaved} onCancel={() => {}} />);

    await user.type(screen.getByLabelText('Name'), 'Belt Squat');
    await user.selectOptions(screen.getByLabelText('Primary muscle'), 'quads');
    await user.selectOptions(screen.getByLabelText('Equipment'), 'machine');
    await user.click(screen.getByLabelText('Glutes'));
    await user.click(screen.getByRole('button', { name: 'Save' }));

    expect(onSaved).toHaveBeenCalled();
    const [url, init] = fetch.mock.calls[0];
    expect(url).toBe('/api/exercises');
    expect(JSON.parse(String(init?.body))).toEqual({
      name: 'Belt Squat',
      primaryMuscle: 'quads',
      secondaryMuscles: ['glutes'],
      equipment: 'machine',
    });
  });

  it('shows field errors from the server', async () => {
    mockApi(() => [
      422,
      {
        error: 'validation_failed',
        message: 'Some fields are invalid.',
        fields: { name: 'An exercise with this name already exists.' },
      },
    ]);
    const user = userEvent.setup();
    render(<ExerciseForm onSaved={() => {}} onCancel={() => {}} />);
    await user.type(screen.getByLabelText('Name'), 'Bench Press');
    await user.click(screen.getByRole('button', { name: 'Save' }));
    expect(
      await screen.findByText('An exercise with this name already exists.'),
    ).toBeInTheDocument();
  });
});

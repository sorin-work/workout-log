import { screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { afterEach, describe, expect, it, vi } from 'vitest';
import { demoUser, mockApi, renderAt } from '../test/render';
import TemplateEditPage from './TemplateEditPage';

afterEach(() => {
  vi.unstubAllGlobals();
});

const squat = {
  id: 1,
  name: 'Back Squat',
  primaryMuscle: 'quads',
  secondaryMuscles: [],
  equipment: 'barbell',
  custom: false,
  archivedAt: null,
};

describe('TemplateEditPage', () => {
  it('builds a template with progression and converts pounds', async () => {
    const fetch = mockApi((method, path) => {
      if (path === '/me') return [200, { ...demoUser, unit: 'lb' }];
      if (path.startsWith('/exercises')) return [200, [squat]];
      if (method === 'POST' && path === '/templates') return [201, { id: 5 }];
      return [200, []];
    });
    const user = userEvent.setup();
    renderAt('/templates/new', {
      '/templates/new': <TemplateEditPage />,
      '/templates': <p>list</p>,
    });

    await user.type(await screen.findByLabelText('Name'), 'Squat day');
    await user.click(screen.getByRole('button', { name: 'Add exercise' }));
    await user.click(await screen.findByRole('button', { name: /Back Squat/ }));
    await user.clear(screen.getByLabelText('Sets'));
    await user.type(screen.getByLabelText('Sets'), '5');
    await user.clear(screen.getByLabelText('Reps'));
    await user.type(screen.getByLabelText('Reps'), '5');
    await user.clear(screen.getByLabelText('Weight'));
    await user.type(screen.getByLabelText('Weight'), '225');
    await user.clear(screen.getByLabelText('Increment'));
    await user.type(screen.getByLabelText('Increment'), '5');
    await user.click(screen.getByRole('button', { name: 'Save template' }));

    expect(await screen.findByText('list')).toBeInTheDocument();
    const call = fetch.mock.calls.find(
      ([url, init]) => url === '/api/templates' && init?.method === 'POST',
    )!;
    const body = JSON.parse(String(call[1]?.body));
    expect(body.name).toBe('Squat day');
    expect(body.exercises[0]).toMatchObject({ exerciseId: 1, sets: 5, reps: 5 });
    expect(body.exercises[0].weightKg).toBeCloseTo(102.058, 3);
    expect(body.exercises[0].incrementKg).toBeCloseTo(2.268, 3);
  });
});

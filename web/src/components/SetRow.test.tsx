import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, expect, it, vi } from 'vitest';
import type { WorkoutSet } from '../api/types';
import { SetRow } from './SetRow';

const base: WorkoutSet = {
  id: 1,
  exerciseId: 6,
  exerciseName: 'Bench Press',
  position: 1,
  reps: 5,
  weightKg: 80,
  rpe: null,
  completed: false,
  warmup: false,
};

function renderRow(set: WorkoutSet, unit: 'kg' | 'lb' = 'kg') {
  const onChange = vi.fn();
  render(
    <table>
      <tbody>
        <SetRow set={set} index={1} unit={unit} onChange={onChange} onDelete={() => {}} />
      </tbody>
    </table>,
  );
  return onChange;
}

describe('SetRow', () => {
  it('saves edited reps and weight on blur', async () => {
    const user = userEvent.setup();
    const onChange = renderRow(base);
    const weight = screen.getByLabelText('Weight (kg)');
    await user.clear(weight);
    await user.type(weight, '82,5');
    await user.tab();
    expect(onChange).toHaveBeenLastCalledWith(expect.objectContaining({ weightKg: 82.5, reps: 5 }));
  });

  it('does not save when nothing changed', async () => {
    const user = userEvent.setup();
    const onChange = renderRow(base);
    await user.click(screen.getByLabelText('Reps'));
    await user.tab();
    expect(onChange).not.toHaveBeenCalled();
  });

  it('ticks a set off', async () => {
    const user = userEvent.setup();
    const onChange = renderRow(base);
    await user.click(screen.getByLabelText('Completed'));
    expect(onChange).toHaveBeenCalledWith(expect.objectContaining({ completed: true }));
  });

  it('shows weights in pounds', () => {
    renderRow({ ...base, weightKg: 102.0583 }, 'lb');
    expect(screen.getByLabelText('Weight (lb)')).toHaveValue('225');
  });

  it('clears the RPE', async () => {
    const user = userEvent.setup();
    const onChange = renderRow({ ...base, rpe: 8 });
    await user.clear(screen.getByLabelText('RPE'));
    await user.tab();
    expect(onChange).toHaveBeenCalledWith(expect.objectContaining({ rpe: null }));
  });
});

describe('SetRow after saving', () => {
  it('shows the value the server stored', () => {
    const props = { index: 1, unit: 'kg' as const, onChange: () => {}, onDelete: () => {} };
    const { rerender } = render(
      <table>
        <tbody>
          <SetRow set={base} {...props} />
        </tbody>
      </table>,
    );
    rerender(
      <table>
        <tbody>
          <SetRow set={{ ...base, reps: 3, weightKg: 82.5 }} {...props} />
        </tbody>
      </table>,
    );
    expect(screen.getByLabelText('Weight (kg)')).toHaveValue('82.5');
    expect(screen.getByLabelText('Reps')).toHaveValue('3');
  });
});

describe('SetRow keyboard', () => {
  it('ticks the set off with Enter', async () => {
    const user = userEvent.setup();
    const onChange = renderRow(base);
    await user.click(screen.getByLabelText('Reps'));
    await user.keyboard('{Backspace}6{Enter}');
    expect(onChange).toHaveBeenCalledWith(expect.objectContaining({ reps: 6, completed: true }));
  });
});

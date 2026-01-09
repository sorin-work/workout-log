import { useState } from 'react';
import type { WorkoutSet } from '../api/types';
import { fromKg, toKg, trimNumber, type Unit } from '../lib/units';

export interface SetChanges {
  reps: number;
  weightKg: number;
  rpe: number | null;
  completed: boolean;
  warmup: boolean;
}

interface Props {
  set: WorkoutSet;
  index: number;
  unit: Unit;
  onChange(changes: SetChanges): void;
  onDelete(): void;
}

function parseNumber(s: string): number {
  const n = Number(s.replace(',', '.'));
  return Number.isFinite(n) ? n : 0;
}

export function SetRow({ set, index, unit, onChange, onDelete }: Props) {
  const [reps, setReps] = useState(String(set.reps));
  const [weight, setWeight] = useState(trimNumber(fromKg(set.weightKg, unit), 2));
  const [rpe, setRpe] = useState(set.rpe === null ? '' : String(set.rpe));

  function current(): SetChanges {
    return {
      reps: Math.round(parseNumber(reps)),
      weightKg: toKg(parseNumber(weight), unit),
      rpe: rpe.trim() === '' ? null : parseNumber(rpe),
      completed: set.completed,
      warmup: set.warmup,
    };
  }

  function commit(patch: Partial<SetChanges> = {}) {
    const next = { ...current(), ...patch };
    if (
      next.reps !== set.reps ||
      next.weightKg !== set.weightKg ||
      next.rpe !== set.rpe ||
      next.completed !== set.completed ||
      next.warmup !== set.warmup
    ) {
      onChange(next);
    }
  }

  return (
    <tr className={set.completed ? 'set done' : 'set'}>
      <td>
        <button
          className={set.warmup ? 'set-tag warmup' : 'set-tag'}
          title={set.warmup ? 'Warm-up set (click to make it a working set)' : 'Mark as warm-up'}
          onClick={() => commit({ warmup: !set.warmup })}
        >
          {set.warmup ? 'W' : index}
        </button>
      </td>
      <td>
        <input
          aria-label={`Weight (${unit})`}
          inputMode="decimal"
          value={weight}
          onChange={(e) => setWeight(e.target.value)}
          onBlur={() => commit()}
        />
      </td>
      <td>
        <input
          aria-label="Reps"
          inputMode="numeric"
          value={reps}
          onChange={(e) => setReps(e.target.value)}
          onBlur={() => commit()}
        />
      </td>
      <td>
        <input
          aria-label="RPE"
          inputMode="decimal"
          placeholder="–"
          value={rpe}
          onChange={(e) => setRpe(e.target.value)}
          onBlur={() => commit()}
        />
      </td>
      <td>
        <input
          type="checkbox"
          aria-label="Completed"
          checked={set.completed}
          onChange={(e) => commit({ completed: e.target.checked })}
        />
      </td>
      <td>
        <button className="icon-btn" aria-label="Delete set" onClick={onDelete}>
          ×
        </button>
      </td>
    </tr>
  );
}

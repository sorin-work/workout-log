import { useEffect, useState } from 'react';
import { ApiError, get } from '../api/client';
import type { LastSession } from '../api/types';
import { formatDate } from '../lib/format';
import { fromKg, trimNumber, type Unit } from '../lib/units';

interface Props {
  exerciseId: number;
  workoutId: number;
  unit: Unit;
}

/** "Last time (Mon, 16 Feb 2026): 100×5, 100×5, 100×4". */
export function LastTime({ exerciseId, workoutId, unit }: Props) {
  const [last, setLast] = useState<LastSession | null>(null);

  useEffect(() => {
    let cancelled = false;
    get<LastSession>(`/exercises/${exerciseId}/last?excludeWorkout=${workoutId}`)
      .then((l) => !cancelled && setLast(l))
      .catch((err) => {
        if (!(err instanceof ApiError && err.status === 404)) console.error(err);
      });
    return () => {
      cancelled = true;
    };
  }, [exerciseId, workoutId]);

  if (!last) return null;
  const sets = last.sets
    .filter((s) => !s.warmup)
    .map((s) => `${trimNumber(fromKg(s.weightKg, unit))}×${s.reps}`)
    .join(', ');
  return (
    <p className="muted small last-time">
      Last time ({formatDate(last.performedAt)}): {sets || 'warm-ups only'}
    </p>
  );
}

import type { Workout } from '../api/types';
import { groupSets } from './groupSets';
import { formatDate } from './format';
import { fromKg, trimNumber, type Unit } from './units';

/**
 * Plain-text summary of a workout for chats and notes:
 *
 *   Day A, Mon, 5 Oct 2026
 *   Back Squat: 100×5, 100×5, 100×5
 */
export function workoutText(w: Workout, unit: Unit): string {
  const lines = [`${w.name}, ${formatDate(w.startedAt)}`];
  for (const g of groupSets(w.sets)) {
    const sets = g.sets
      .filter((s) => s.completed && !s.warmup)
      .map((s) =>
        s.weightKg > 0 ? `${trimNumber(fromKg(s.weightKg, unit))}×${s.reps}` : `${s.reps} reps`,
      );
    if (sets.length > 0) lines.push(`${g.exerciseName}: ${sets.join(', ')}`);
  }
  if (unit === 'lb') lines.push('(weights in lb)');
  return lines.join('\n');
}

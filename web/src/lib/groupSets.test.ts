import { describe, expect, it } from 'vitest';
import type { WorkoutSet } from '../api/types';
import { groupSets } from './groupSets';

function set(id: number, exerciseId: number, position: number): WorkoutSet {
  return {
    id,
    exerciseId,
    exerciseName: `Exercise ${exerciseId}`,
    position,
    reps: 5,
    weightKg: 100,
    rpe: null,
    completed: true,
    warmup: false,
  };
}

describe('groupSets', () => {
  it('groups by exercise in the order they were first logged', () => {
    const groups = groupSets([set(1, 7, 1), set(2, 3, 3), set(3, 7, 2), set(4, 3, 4)]);
    expect(groups.map((g) => g.exerciseId)).toEqual([7, 3]);
    expect(groups[0].sets.map((s) => s.id)).toEqual([1, 3]);
  });

  it('returns nothing for an empty workout', () => {
    expect(groupSets([])).toEqual([]);
  });
});

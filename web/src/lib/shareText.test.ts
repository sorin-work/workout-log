import { describe, expect, it } from 'vitest';
import type { Workout } from '../api/types';
import { workoutText } from './shareText';

const w: Workout = {
  id: 1,
  name: 'Day A',
  notes: '',
  startedAt: '2026-10-05T17:00:00Z',
  finishedAt: null,
  templateId: null,
  sets: [
    {
      id: 1,
      exerciseId: 1,
      exerciseName: 'Back Squat',
      position: 1,
      reps: 5,
      weightKg: 20,
      rpe: null,
      completed: true,
      warmup: true,
    },
    {
      id: 2,
      exerciseId: 1,
      exerciseName: 'Back Squat',
      position: 2,
      reps: 5,
      weightKg: 100,
      rpe: 8,
      completed: true,
      warmup: false,
    },
    {
      id: 3,
      exerciseId: 2,
      exerciseName: 'Pull-Up',
      position: 3,
      reps: 8,
      weightKg: 0,
      rpe: null,
      completed: true,
      warmup: false,
    },
    {
      id: 4,
      exerciseId: 3,
      exerciseName: 'Dip',
      position: 4,
      reps: 8,
      weightKg: 0,
      rpe: null,
      completed: false,
      warmup: false,
    },
  ],
};

describe('workoutText', () => {
  it('lists completed working sets per exercise', () => {
    expect(workoutText(w, 'kg')).toBe('Day A, Mon, 5 Oct 2026\nBack Squat: 100×5\nPull-Up: 8 reps');
  });

  it('notes the unit for pounds', () => {
    expect(workoutText(w, 'lb')).toContain('Back Squat: 220.5×5');
    expect(workoutText(w, 'lb').endsWith('(weights in lb)')).toBe(true);
  });
});

import type { WorkoutSet } from '../api/types';

export interface ExerciseGroup {
  exerciseId: number;
  exerciseName: string;
  sets: WorkoutSet[];
}

/** Groups sets by exercise, keeping exercises in the order first logged. */
export function groupSets(sets: WorkoutSet[]): ExerciseGroup[] {
  const groups = new Map<number, ExerciseGroup>();
  for (const s of [...sets].sort((a, b) => a.position - b.position)) {
    let g = groups.get(s.exerciseId);
    if (!g) {
      g = { exerciseId: s.exerciseId, exerciseName: s.exerciseName, sets: [] };
      groups.set(s.exerciseId, g);
    }
    g.sets.push(s);
  }
  return [...groups.values()];
}

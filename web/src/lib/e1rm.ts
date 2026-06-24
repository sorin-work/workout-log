export type Formula = 'epley' | 'brzycki';

/** Same rules as the API: singles are the 1RM, nothing above 12 reps. */
export function estimateOneRepMax(weight: number, reps: number, formula: Formula): number {
  if (weight <= 0 || reps <= 0 || reps > 12) return 0;
  if (reps === 1) return weight;
  const est = formula === 'brzycki' ? (weight * 36) / (37 - reps) : weight * (1 + reps / 30);
  return Math.round(est * 10) / 10;
}

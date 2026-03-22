export interface User {
  id: number;
  email: string;
  displayName: string;
  unit: 'kg' | 'lb';
  timezone: string;
  createdAt: string;
}

export interface Exercise {
  id: number;
  name: string;
  primaryMuscle: string;
  secondaryMuscles: string[];
  equipment: string;
  custom: boolean;
  archivedAt: string | null;
}

export interface WorkoutSet {
  id: number;
  exerciseId: number;
  exerciseName: string;
  position: number;
  reps: number;
  weightKg: number;
  rpe: number | null;
  completed: boolean;
  warmup: boolean;
}

export interface Workout {
  id: number;
  name: string;
  notes: string;
  startedAt: string;
  finishedAt: string | null;
  templateId: number | null;
  sets: WorkoutSet[];
}

export interface WorkoutSummary {
  id: number;
  name: string;
  startedAt: string;
  finishedAt: string | null;
  exerciseCount: number;
  setCount: number;
  volumeKg: number;
}

export interface Page<T> {
  items: T[];
  page: number;
  pageSize: number;
  total: number;
  totalPages: number;
}

export const MUSCLE_GROUPS = [
  'chest',
  'back',
  'shoulders',
  'biceps',
  'triceps',
  'forearms',
  'quads',
  'hamstrings',
  'glutes',
  'calves',
  'core',
] as const;

export const EQUIPMENT = [
  'barbell',
  'dumbbell',
  'machine',
  'cable',
  'bodyweight',
  'kettlebell',
  'band',
  'other',
] as const;

export interface TemplateExercise {
  id: number;
  exerciseId: number;
  exerciseName: string;
  position: number;
  sets: number;
  reps: number;
  weightKg: number;
  restSeconds: number;
  incrementKg: number;
  deloadAfter: number;
  deloadPercent: number;
  failures: number;
}

export interface Template {
  id: number;
  name: string;
  notes: string;
  exercises: TemplateExercise[];
}

export interface ProgressionChange {
  templateExerciseId: number;
  exerciseId: number;
  exerciseName: string;
  outcome: 'increased' | 'kept' | 'deloaded';
  previousWeightKg: number;
  weightKg: number;
  failures: number;
}

export interface FinishResult {
  workout: Workout;
  progression: ProgressionChange[];
}

export interface ProgramDay {
  position: number;
  templateId: number;
  templateName: string;
}

export interface Program {
  id: number;
  name: string;
  active: boolean;
  nextPosition: number;
  days: ProgramDay[];
}

export interface RecordSet {
  value: number;
  weightKg: number;
  reps: number;
  workoutId: number;
  performedAt: string;
}

export interface ExerciseRecords {
  exerciseId: number;
  exerciseName: string;
  estimatedMax: RecordSet | null;
  heaviest: RecordSet | null;
  mostReps: RecordSet | null;
  bestVolume: RecordSet | null;
  repMaxes: Record<string, RecordSet>;
  setCount: number;
  lastDone: string;
}

export interface ProgressPoint {
  workoutId: number;
  performedAt: string;
  estimatedMax: number;
  topWeightKg: number;
  volumeKg: number;
  sets: number;
  reps: number;
}

export interface MuscleVolume {
  sets: number;
  volumeKg: number;
}

export interface WeekVolume {
  weekStart: string;
  muscles: Record<string, MuscleVolume>;
  totalSets: number;
}

export interface LastSession {
  workoutId: number;
  performedAt: string;
  sets: WorkoutSet[];
}

export interface WeekTotals {
  workouts: number;
  sets: number;
  volumeKg: number;
}

export interface Summary {
  thisWeek: WeekTotals;
  lastWeek: WeekTotals;
  weekStreak: number;
  totalWorkouts: number;
}

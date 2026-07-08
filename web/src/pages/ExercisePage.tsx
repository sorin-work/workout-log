import { useState } from 'react';
import { Link, useParams } from 'react-router';
import type { Exercise, ExerciseRecords, LastSession, ProgressPoint } from '../api/types';
import { LineChart } from '../components/charts/LineChart';
import { ExerciseNotes } from '../components/ExerciseNotes';
import { formatDate } from '../lib/format';
import { capitalize } from '../lib/text';
import { formatWeight, fromKg, trimNumber } from '../lib/units';
import { useFetch } from '../lib/useFetch';
import { useTitle } from '../lib/useTitle';
import { useUnit } from '../lib/useUnit';

type Metric = 'estimatedMax' | 'topWeightKg' | 'volumeKg';

const METRICS: { key: Metric; label: string }[] = [
  { key: 'estimatedMax', label: 'Estimated 1RM' },
  { key: 'topWeightKg', label: 'Top set' },
  { key: 'volumeKg', label: 'Volume' },
];

export default function ExercisePage() {
  const { id } = useParams();
  const unit = useUnit();
  const [metric, setMetric] = useState<Metric>('estimatedMax');
  const exercise = useFetch<Exercise>(`/exercises/${id}`);
  useTitle(exercise.data?.name);
  const records = useFetch<ExerciseRecords>(`/exercises/${id}/records`);
  const progress = useFetch<ProgressPoint[]>(`/exercises/${id}/progress`);
  const history = useFetch<LastSession[]>(`/exercises/${id}/history?limit=8`);

  if (exercise.error) return <p className="alert">{exercise.error}</p>;
  if (!exercise.data) return <p className="muted">Loading…</p>;
  const e = exercise.data;
  const r = records.data;
  const points = (progress.data ?? [])
    .filter((p) => p[metric] > 0)
    .map((p) => ({
      x: Date.parse(p.performedAt),
      y: fromKg(p[metric], unit),
      label: `${formatDate(p.performedAt)}: ${trimNumber(fromKg(p[metric], unit))} ${unit}`,
    }));
  const repMaxes = r
    ? Object.entries(r.repMaxes)
        .map(([reps, set]) => ({ reps: Number(reps), set }))
        .sort((a, b) => a.reps - b.reps)
    : [];

  return (
    <>
      <div className="page-header">
        <div>
          <h1>{e.name}</h1>
          <div className="muted small">
            {capitalize(e.primaryMuscle)}
            {e.secondaryMuscles.length > 0 && ` · ${e.secondaryMuscles.join(', ')}`} ·{' '}
            {capitalize(e.equipment)}
          </div>
        </div>
      </div>

      <ExerciseNotes exerciseId={e.id} />

      {r && r.setCount === 0 && <p className="empty">You have not logged this exercise yet.</p>}

      {r && r.setCount > 0 && (
        <>
          <div className="stats">
            <div className="card stat">
              <div className="muted small">Estimated 1RM</div>
              <strong>{r.estimatedMax ? formatWeight(r.estimatedMax.value, unit) : '–'}</strong>
            </div>
            <div className="card stat">
              <div className="muted small">Heaviest</div>
              <strong>
                {r.heaviest
                  ? `${formatWeight(r.heaviest.weightKg, unit)} × ${r.heaviest.reps}`
                  : '–'}
              </strong>
            </div>
            <div className="card stat">
              <div className="muted small">Sets logged</div>
              <strong>{r.setCount}</strong>
            </div>
          </div>

          <section className="card">
            <div className="page-header">
              <h2>Progress</h2>
              <div className="segmented">
                {METRICS.map((m) => (
                  <button
                    key={m.key}
                    className={metric === m.key ? 'active' : ''}
                    onClick={() => setMetric(m.key)}
                  >
                    {m.label}
                  </button>
                ))}
              </div>
            </div>
            <LineChart
              points={points}
              ariaLabel={`${e.name} ${metric} over time`}
              formatY={(v) => trimNumber(v, 0)}
            />
          </section>

          {history.data && history.data.length > 0 && (
            <section className="card">
              <h2>Recent sessions</h2>
              <ul className="list">
                {history.data.map((h) => (
                  <li key={h.workoutId} className="list-row">
                    <Link to={`/workouts/${h.workoutId}`}>{formatDate(h.performedAt)}</Link>
                    <span className="muted small">
                      {h.sets
                        .filter((s) => !s.warmup)
                        .map((s) => `${trimNumber(fromKg(s.weightKg, unit))}×${s.reps}`)
                        .join(', ')}
                    </span>
                  </li>
                ))}
              </ul>
            </section>
          )}

          {repMaxes.length > 0 && (
            <section className="card">
              <h2>Rep maxes</h2>
              <table className="table">
                <thead>
                  <tr>
                    <th>Reps</th>
                    <th>Best weight</th>
                    <th>When</th>
                  </tr>
                </thead>
                <tbody>
                  {repMaxes.map(({ reps, set }) => (
                    <tr key={reps}>
                      <td>{reps}</td>
                      <td>{formatWeight(set.weightKg, unit)}</td>
                      <td>
                        <Link to={`/workouts/${set.workoutId}`}>{formatDate(set.performedAt)}</Link>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </section>
          )}
        </>
      )}
      <div className="page-footer">
        <Link to="/exercises">Back to exercises</Link>
        {r && r.setCount > 0 && (
          <Link to={`/workouts?exercise=${e.id}`}>Workouts with {e.name}</Link>
        )}
      </div>
    </>
  );
}

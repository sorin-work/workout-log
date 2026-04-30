import { Link, useSearchParams } from 'react-router';
import type { Exercise, Page, WorkoutSummary } from '../api/types';
import { formatDate, formatDuration } from '../lib/format';
import { formatVolume, formatWeight } from '../lib/units';
import { useFetch } from '../lib/useFetch';
import { useUnit } from '../lib/useUnit';

export default function WorkoutsPage() {
  const [params, setParams] = useSearchParams();
  const unit = useUnit();
  const page = Math.max(1, Number(params.get('page')) || 1);
  const exercise = params.get('exercise') ?? '';
  const query = new URLSearchParams({ page: String(page) });
  if (exercise) query.set('exerciseId', exercise);
  const { data, error, loading } = useFetch<Page<WorkoutSummary>>(`/workouts?${query}`);
  const exercises = useFetch<Exercise[]>('/exercises?archived=true');

  function goTo(p: number, ex = exercise) {
    const next: Record<string, string> = {};
    if (p > 1) next.page = String(p);
    if (ex) next.exercise = ex;
    setParams(next);
  }

  return (
    <>
      <div className="page-header">
        <h1>Workouts</h1>
        <select
          aria-label="Filter by exercise"
          value={exercise}
          onChange={(e) => goTo(1, e.target.value)}
        >
          <option value="">All exercises</option>
          {exercises.data?.map((e) => (
            <option key={e.id} value={e.id}>
              {e.name}
            </option>
          ))}
        </select>
      </div>
      {error && <p className="alert">{error}</p>}
      {loading && !data && <p className="muted">Loading…</p>}
      {data && data.total === 0 && exercise && (
        <p className="empty">No workouts with this exercise yet.</p>
      )}
      {data && data.total === 0 && !exercise && (
        <div className="empty">
          <p>No workouts logged yet.</p>
          <Link to="/">Start your first one</Link>
        </div>
      )}
      {data && data.items.length > 0 && (
        <ul className="list card">
          {data.items.map((w) => (
            <li key={w.id} className="list-row">
              <div>
                <Link to={`/workouts/${w.id}`}>
                  <strong>{w.name}</strong>
                </Link>
                <div className="muted small">
                  {formatDate(w.startedAt)}
                  {w.finishedAt
                    ? ` · ${formatDuration(w.startedAt, w.finishedAt)}`
                    : ' · in progress'}
                </div>
                {w.exercises.length > 0 && (
                  <ul className="exercise-summary small">
                    {w.exercises.map((e) => (
                      <li key={e.exerciseId}>
                        {e.sets}× {e.name}
                        {e.topKg > 0 && (
                          <span className="muted">
                            {' '}
                            · best {formatWeight(e.topKg, unit)} × {e.topReps}
                          </span>
                        )}
                      </li>
                    ))}
                  </ul>
                )}
              </div>
              <div className="muted small right">
                {w.exerciseCount} exercises · {w.setCount} sets
                <br />
                {formatVolume(w.volumeKg, unit)}
              </div>
            </li>
          ))}
        </ul>
      )}
      {data && data.total > 0 && (
        <div className="pager">
          <button className="btn small" disabled={page <= 1} onClick={() => goTo(page - 1)}>
            Newer
          </button>
          <span className="muted small">
            Page {page} of {Math.max(1, data.totalPages)}
          </span>
          <button
            className="btn small"
            disabled={page >= data.totalPages}
            onClick={() => goTo(page + 1)}
          >
            Older
          </button>
        </div>
      )}
    </>
  );
}

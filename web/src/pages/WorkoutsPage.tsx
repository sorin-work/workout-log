import { Link, useSearchParams } from 'react-router';
import type { Page, WorkoutSummary } from '../api/types';
import { formatDate, formatDuration } from '../lib/format';
import { formatVolume } from '../lib/units';
import { useFetch } from '../lib/useFetch';

export default function WorkoutsPage() {
  const [params, setParams] = useSearchParams();
  const page = Math.max(1, Number(params.get('page')) || 1);
  const { data, error, loading } = useFetch<Page<WorkoutSummary>>(`/workouts?page=${page}`);

  function goTo(p: number) {
    setParams(p === 1 ? {} : { page: String(p) });
  }

  return (
    <>
      <div className="page-header">
        <h1>Workouts</h1>
      </div>
      {error && <p className="alert">{error}</p>}
      {loading && !data && <p className="muted">Loading…</p>}
      {data && data.total === 0 && (
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
              </div>
              <div className="muted small right">
                {w.exerciseCount} exercises · {w.setCount} sets
                <br />
                {formatVolume(w.volumeKg, 'kg')}
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

import { useState } from 'react';
import { Link, useNavigate } from 'react-router';
import { errorMessage, post } from '../api/client';
import type { Page, Template, Workout, WorkoutSummary } from '../api/types';
import { useAuth } from '../auth/context';
import { formatDate } from '../lib/format';
import { formatVolume } from '../lib/units';
import { useFetch } from '../lib/useFetch';
import { useUnit } from '../lib/useUnit';

export default function HomePage() {
  const { user } = useAuth();
  const navigate = useNavigate();
  const unit = useUnit();
  const [error, setError] = useState('');
  const [busy, setBusy] = useState(false);
  const recent = useFetch<Page<WorkoutSummary>>('/workouts?pageSize=5');
  const inProgress = recent.data?.items.find((w) => !w.finishedAt);

  const templates = useFetch<Template[]>('/templates');

  async function start(template?: Template) {
    setBusy(true);
    try {
      const w = template
        ? await post<Workout>(`/templates/${template.id}/start`)
        : await post<Workout>('/workouts', {});
      navigate(`/workouts/${w.id}`);
    } catch (err) {
      setError(errorMessage(err));
      setBusy(false);
    }
  }

  return (
    <>
      <h1>Welcome back{user?.displayName ? `, ${user.displayName}` : ''}</h1>
      {error && <p className="alert">{error}</p>}
      {inProgress && (
        <div className="card highlight">
          <strong>{inProgress.name}</strong> is still in progress.{' '}
          <Link to={`/workouts/${inProgress.id}`}>Continue</Link>
        </div>
      )}
      <div className="start-buttons">
        {templates.data?.map((t) => (
          <button key={t.id} className="btn primary" onClick={() => start(t)} disabled={busy}>
            Start {t.name}
          </button>
        ))}
        <button className="btn" onClick={() => start()} disabled={busy}>
          Start an empty workout
        </button>
      </div>
      {templates.data?.length === 0 && (
        <p className="muted small">
          Tip: <Link to="/templates/new">create a template</Link> for the workouts you repeat.
        </p>
      )}

      <h2 className="section-title">Recent workouts</h2>
      {recent.data && recent.data.items.length === 0 && (
        <p className="muted">Nothing logged yet. Your workouts will show up here.</p>
      )}
      {recent.data && recent.data.items.length > 0 && (
        <ul className="list card">
          {recent.data.items.map((w) => (
            <li key={w.id} className="list-row">
              <Link to={`/workouts/${w.id}`}>{w.name}</Link>
              <span className="muted small">
                {formatDate(w.startedAt)} · {formatVolume(w.volumeKg, unit)}
              </span>
            </li>
          ))}
        </ul>
      )}
    </>
  );
}

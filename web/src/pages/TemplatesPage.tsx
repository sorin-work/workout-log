import { useState } from 'react';
import { Link, useNavigate } from 'react-router';
import { del, errorMessage, post } from '../api/client';
import type { Template, Workout } from '../api/types';
import { formatWeight } from '../lib/units';
import { useFetch } from '../lib/useFetch';
import { useUnit } from '../lib/useUnit';

export default function TemplatesPage() {
  const unit = useUnit();
  const navigate = useNavigate();
  const { data, error, loading, reload } = useFetch<Template[]>('/templates');
  const [actionError, setActionError] = useState('');

  async function start(t: Template) {
    setActionError('');
    try {
      const w = await post<Workout>(`/templates/${t.id}/start`);
      navigate(`/workouts/${w.id}`);
    } catch (err) {
      setActionError(errorMessage(err));
    }
  }

  async function remove(t: Template) {
    if (!window.confirm(`Delete the template "${t.name}"? Past workouts are kept.`)) return;
    setActionError('');
    try {
      await del(`/templates/${t.id}`);
      reload();
    } catch (err) {
      setActionError(errorMessage(err));
    }
  }

  return (
    <>
      <div className="page-header">
        <h1>Templates</h1>
        <Link className="btn primary" to="/templates/new">
          New template
        </Link>
      </div>
      {(error || actionError) && <p className="alert">{error || actionError}</p>}
      {loading && !data && <p className="muted">Loading…</p>}
      {data && data.length === 0 && (
        <div className="empty">
          <p>Templates let you start the same workout again with one click.</p>
          <Link to="/templates/new">Create your first template</Link>
        </div>
      )}
      <div className="grid">
        {data?.map((t) => (
          <section key={t.id} className="card template-card">
            <h2>{t.name}</h2>
            <ul className="plain">
              {t.exercises.map((e) => (
                <li key={e.id}>
                  {e.exerciseName}{' '}
                  <span className="muted">
                    {e.sets}×{e.reps}
                    {e.weightKg > 0 && ` @ ${formatWeight(e.weightKg, unit)}`}
                  </span>
                  {e.incrementKg > 0 && (
                    <span className="badge" title="Linear progression">
                      +{formatWeight(e.incrementKg, unit)}
                    </span>
                  )}
                </li>
              ))}
              {t.exercises.length === 0 && <li className="muted">No exercises yet</li>}
            </ul>
            <div className="actions">
              <button className="btn primary small" onClick={() => start(t)}>
                Start
              </button>
              <Link className="btn small" to={`/templates/${t.id}/edit`}>
                Edit
              </Link>
              <button className="btn danger small" onClick={() => remove(t)}>
                Delete
              </button>
            </div>
          </section>
        ))}
      </div>
    </>
  );
}

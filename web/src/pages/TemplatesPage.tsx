import { Link, useNavigate } from 'react-router';
import { del, post } from '../api/client';
import type { Template, Workout } from '../api/types';
import { formatWeight } from '../lib/units';
import { Skeleton } from '../components/Skeleton';
import { useAction } from '../lib/useAction';
import { useFetch } from '../lib/useFetch';
import { useUnit } from '../lib/useUnit';
import { useTitle } from '../lib/useTitle';

export default function TemplatesPage() {
  useTitle('Templates');
  const unit = useUnit();
  const navigate = useNavigate();
  const { data, error, loading, reload } = useFetch<Template[]>('/templates');
  const action = useAction();

  const start = (t: Template) =>
    action.run(async () => {
      const w = await post<Workout>(`/templates/${t.id}/start`);
      navigate(`/workouts/${w.id}`);
    });

  const copy = (t: Template) =>
    action.run(async () => {
      const created = await post<Template>(`/templates/${t.id}/copy`);
      navigate(`/templates/${created.id}/edit`);
    });

  const remove = (t: Template) =>
    action.run(async () => {
      if (!window.confirm(`Delete the template "${t.name}"? Past workouts are kept.`)) return;
      await del(`/templates/${t.id}`);
      reload();
    });

  return (
    <>
      <div className="page-header">
        <h1>Templates</h1>
        <Link className="btn primary" to="/templates/new">
          New template
        </Link>
      </div>
      {(error || action.error) && <p className="alert">{error || action.error}</p>}
      {loading && !data && <Skeleton />}
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
                  {e.failures > 0 && (
                    <span
                      className="badge muted-badge"
                      title={`Missed ${e.failures} session(s) in a row at this weight`}
                    >
                      {e.failures}/{e.deloadAfter || '∞'} missed
                    </span>
                  )}
                </li>
              ))}
              {t.exercises.length === 0 && <li className="muted">No exercises yet</li>}
            </ul>
            <div className="actions">
              <button className="btn primary small" disabled={action.busy} onClick={() => start(t)}>
                Start
              </button>
              <Link className="btn small" to={`/templates/${t.id}/edit`}>
                Edit
              </Link>
              <button className="btn small" onClick={() => copy(t)}>
                Duplicate
              </button>
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

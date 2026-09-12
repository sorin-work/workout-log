import { useState } from 'react';
import { Link } from 'react-router';
import { del, errorMessage, post } from '../api/client';
import { MUSCLE_GROUPS, EQUIPMENT, type Exercise } from '../api/types';
import { ExerciseForm } from '../components/ExerciseForm';
import { Skeleton } from '../components/Skeleton';
import { useFetch } from '../lib/useFetch';
import { capitalize } from '../lib/text';
import { useTitle } from '../lib/useTitle';

export default function ExercisesPage() {
  useTitle('Exercises');
  const [q, setQ] = useState('');
  const [muscle, setMuscle] = useState('');
  const [equipment, setEquipment] = useState('');
  const [showArchived, setShowArchived] = useState(false);

  const params = new URLSearchParams();
  if (q.trim()) params.set('q', q.trim());
  if (muscle) params.set('muscle', muscle);
  if (equipment) params.set('equipment', equipment);
  if (showArchived) params.set('archived', 'true');
  const { data, error, loading, reload } = useFetch<Exercise[]>(`/exercises?${params}`);
  const [creating, setCreating] = useState(false);
  const [editing, setEditing] = useState<Exercise | null>(null);
  const [actionError, setActionError] = useState('');

  async function setArchived(e: Exercise, archived: boolean) {
    setActionError('');
    try {
      await post(`/exercises/${e.id}/${archived ? 'archive' : 'restore'}`);
      reload();
    } catch (err) {
      setActionError(errorMessage(err));
    }
  }

  async function remove(e: Exercise) {
    if (!window.confirm(`Delete ${e.name}?`)) return;
    setActionError('');
    try {
      await del(`/exercises/${e.id}`);
      reload();
    } catch (err) {
      setActionError(errorMessage(err));
    }
  }

  return (
    <>
      <div className="page-header">
        <h1>Exercises</h1>
        {!creating && (
          <button className="btn primary" onClick={() => setCreating(true)}>
            New exercise
          </button>
        )}
      </div>
      {editing && (
        <ExerciseForm
          key={editing.id}
          exercise={editing}
          onCancel={() => setEditing(null)}
          onSaved={() => {
            setEditing(null);
            reload();
          }}
        />
      )}
      {creating && (
        <ExerciseForm
          onCancel={() => setCreating(false)}
          onSaved={() => {
            setCreating(false);
            reload();
          }}
        />
      )}
      <div className="filters">
        <input
          type="search"
          placeholder="Search exercises"
          value={q}
          onChange={(e) => setQ(e.target.value)}
        />
        <select value={muscle} onChange={(e) => setMuscle(e.target.value)}>
          <option value="">All muscles</option>
          {MUSCLE_GROUPS.map((m) => (
            <option key={m} value={m}>
              {capitalize(m)}
            </option>
          ))}
        </select>
        <select value={equipment} onChange={(e) => setEquipment(e.target.value)}>
          <option value="">All equipment</option>
          {EQUIPMENT.map((m) => (
            <option key={m} value={m}>
              {capitalize(m)}
            </option>
          ))}
        </select>
        <label className="chip">
          <input
            type="checkbox"
            checked={showArchived}
            onChange={(e) => setShowArchived(e.target.checked)}
          />
          Show archived
        </label>
      </div>
      {actionError && <p className="alert">{actionError}</p>}
      {error && <p className="alert">{error}</p>}
      {loading && !data && <Skeleton />}
      {data && data.length === 0 && <p className="empty">No exercises match these filters.</p>}
      {data && data.length > 0 && (
        <ul className="list card">
          {data.map((e) => (
            <li key={e.id} className="list-row">
              <div>
                <Link to={`/exercises/${e.id}`}>
                  <strong>{e.name}</strong>
                </Link>
                {e.custom && <span className="badge">Custom</span>}
                {e.archivedAt && <span className="badge muted-badge">Archived</span>}
                <div className="muted small">
                  {capitalize(e.primaryMuscle)}
                  {e.secondaryMuscles.length > 0 && ` · ${e.secondaryMuscles.join(', ')}`}
                </div>
              </div>
              <div className="row-actions">
                <span className="muted small">{capitalize(e.equipment)}</span>
                {e.custom && (
                  <>
                    <button className="btn small" onClick={() => setEditing(e)}>
                      Edit
                    </button>
                    <button className="btn small" onClick={() => setArchived(e, !e.archivedAt)}>
                      {e.archivedAt ? 'Restore' : 'Archive'}
                    </button>
                    <button
                      className="icon-btn"
                      aria-label={`Delete ${e.name}`}
                      title="Delete (only exercises you never logged)"
                      onClick={() => remove(e)}
                    >
                      ×
                    </button>
                  </>
                )}
              </div>
            </li>
          ))}
        </ul>
      )}
    </>
  );
}

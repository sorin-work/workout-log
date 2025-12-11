import { useState } from 'react';
import { MUSCLE_GROUPS, EQUIPMENT, type Exercise } from '../api/types';
import { ExerciseForm } from '../components/ExerciseForm';
import { useFetch } from '../lib/useFetch';
import { capitalize } from '../lib/text';

export default function ExercisesPage() {
  const [q, setQ] = useState('');
  const [muscle, setMuscle] = useState('');
  const [equipment, setEquipment] = useState('');

  const params = new URLSearchParams();
  if (q.trim()) params.set('q', q.trim());
  if (muscle) params.set('muscle', muscle);
  if (equipment) params.set('equipment', equipment);
  const { data, error, loading, reload } = useFetch<Exercise[]>(`/exercises?${params}`);
  const [creating, setCreating] = useState(false);

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
      </div>
      {error && <p className="alert">{error}</p>}
      {loading && !data && <p className="muted">Loading…</p>}
      {data && data.length === 0 && <p className="empty">No exercises match these filters.</p>}
      {data && data.length > 0 && (
        <ul className="list card">
          {data.map((e) => (
            <li key={e.id} className="list-row">
              <div>
                <strong>{e.name}</strong>
                {e.custom && <span className="badge">Custom</span>}
                <div className="muted small">
                  {capitalize(e.primaryMuscle)}
                  {e.secondaryMuscles.length > 0 && ` · ${e.secondaryMuscles.join(', ')}`}
                </div>
              </div>
              <span className="muted small">{capitalize(e.equipment)}</span>
            </li>
          ))}
        </ul>
      )}
    </>
  );
}

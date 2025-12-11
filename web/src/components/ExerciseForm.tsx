import { useState, type FormEvent } from 'react';
import { ApiError, errorMessage, post, put } from '../api/client';
import { EQUIPMENT, MUSCLE_GROUPS, type Exercise } from '../api/types';
import { capitalize } from '../lib/text';
import { FieldError } from './FieldError';

interface Props {
  exercise?: Exercise;
  onSaved(e: Exercise): void;
  onCancel(): void;
}

export function ExerciseForm({ exercise, onSaved, onCancel }: Props) {
  const [name, setName] = useState(exercise?.name ?? '');
  const [primaryMuscle, setPrimary] = useState(exercise?.primaryMuscle ?? 'chest');
  const [secondary, setSecondary] = useState<string[]>(exercise?.secondaryMuscles ?? []);
  const [equipment, setEquipment] = useState(exercise?.equipment ?? 'barbell');
  const [fields, setFields] = useState<Record<string, string>>({});
  const [error, setError] = useState('');
  const [busy, setBusy] = useState(false);

  function toggleSecondary(m: string) {
    setSecondary((s) => (s.includes(m) ? s.filter((x) => x !== m) : [...s, m]));
  }

  async function onSubmit(e: FormEvent) {
    e.preventDefault();
    setBusy(true);
    setFields({});
    setError('');
    const body = {
      name,
      primaryMuscle,
      secondaryMuscles: secondary.filter((m) => m !== primaryMuscle),
      equipment,
    };
    try {
      const saved = exercise
        ? await put<Exercise>(`/exercises/${exercise.id}`, body)
        : await post<Exercise>('/exercises', body);
      onSaved(saved);
    } catch (err) {
      if (err instanceof ApiError && err.code === 'validation_failed') setFields(err.fields);
      else setError(errorMessage(err));
    } finally {
      setBusy(false);
    }
  }

  return (
    <form className="card form" onSubmit={onSubmit}>
      <h2>{exercise ? 'Edit exercise' : 'New exercise'}</h2>
      {error && <p className="alert">{error}</p>}
      <label>
        Name
        <input value={name} onChange={(e) => setName(e.target.value)} required maxLength={40} />
        <FieldError message={fields.name} />
      </label>
      <div className="form-row">
        <label>
          Primary muscle
          <select value={primaryMuscle} onChange={(e) => setPrimary(e.target.value)}>
            {MUSCLE_GROUPS.map((m) => (
              <option key={m} value={m}>
                {capitalize(m)}
              </option>
            ))}
          </select>
        </label>
        <label>
          Equipment
          <select value={equipment} onChange={(e) => setEquipment(e.target.value)}>
            {EQUIPMENT.map((m) => (
              <option key={m} value={m}>
                {capitalize(m)}
              </option>
            ))}
          </select>
          <FieldError message={fields.equipment} />
        </label>
      </div>
      <fieldset className="chips">
        <legend>Also works</legend>
        {MUSCLE_GROUPS.filter((m) => m !== primaryMuscle).map((m) => (
          <label key={m} className="chip">
            <input
              type="checkbox"
              checked={secondary.includes(m)}
              onChange={() => toggleSecondary(m)}
            />
            {capitalize(m)}
          </label>
        ))}
      </fieldset>
      <div className="form-actions">
        <button type="button" className="btn" onClick={onCancel}>
          Cancel
        </button>
        <button className="btn primary" disabled={busy}>
          {busy ? 'Saving…' : 'Save'}
        </button>
      </div>
    </form>
  );
}

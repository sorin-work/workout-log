import { useState, type FormEvent } from 'react';
import { ApiError, errorMessage, put } from '../api/client';
import type { Workout } from '../api/types';
import { toLocalInput } from '../lib/format';
import { FieldError } from './FieldError';

interface Props {
  workout: Workout;
  onSaved(w: Workout): void;
  onCancel(): void;
}

export function WorkoutDetailsForm({ workout, onSaved, onCancel }: Props) {
  const [name, setName] = useState(workout.name);
  const [notes, setNotes] = useState(workout.notes);
  const [startedAt, setStartedAt] = useState(toLocalInput(workout.startedAt));
  const [finishedAt, setFinishedAt] = useState(
    workout.finishedAt ? toLocalInput(workout.finishedAt) : '',
  );
  const [fields, setFields] = useState<Record<string, string>>({});
  const [error, setError] = useState('');

  async function onSubmit(e: FormEvent) {
    e.preventDefault();
    setFields({});
    setError('');
    try {
      onSaved(
        await put<Workout>(`/workouts/${workout.id}`, {
          name,
          notes,
          startedAt: new Date(startedAt).toISOString(),
          finishedAt: finishedAt ? new Date(finishedAt).toISOString() : null,
        }),
      );
    } catch (err) {
      if (err instanceof ApiError && err.code === 'validation_failed') setFields(err.fields);
      else setError(errorMessage(err));
    }
  }

  return (
    <form className="card form" onSubmit={onSubmit}>
      {error && <p className="alert">{error}</p>}
      <label>
        Name
        <input value={name} onChange={(e) => setName(e.target.value)} maxLength={80} />
        <FieldError message={fields.name} />
      </label>
      <div className="form-row">
        <label>
          Started
          <input
            type="datetime-local"
            value={startedAt}
            onChange={(e) => setStartedAt(e.target.value)}
            required
          />
          <FieldError message={fields.startedAt} />
        </label>
        <label>
          Finished
          <input
            type="datetime-local"
            value={finishedAt}
            onChange={(e) => setFinishedAt(e.target.value)}
          />
          <FieldError message={fields.finishedAt} />
        </label>
      </div>
      <label>
        Notes
        <textarea rows={4} value={notes} onChange={(e) => setNotes(e.target.value)} />
        <FieldError message={fields.notes} />
      </label>
      <div className="form-actions">
        <button type="button" className="btn" onClick={onCancel}>
          Cancel
        </button>
        <button className="btn primary">Save</button>
      </div>
    </form>
  );
}

import { useState } from 'react';
import { errorMessage, put } from '../api/client';
import { useFetch } from '../lib/useFetch';

/** Personal cues for an exercise, editable in place. */
export function ExerciseNotes({ exerciseId }: { exerciseId: number }) {
  const { data, setData } = useFetch<{ notes: string }>(`/exercises/${exerciseId}/notes`);
  const [draft, setDraft] = useState<string | null>(null);
  const [error, setError] = useState('');

  async function save() {
    if (draft === null) return;
    setError('');
    try {
      setData(await put<{ notes: string }>(`/exercises/${exerciseId}/notes`, { notes: draft }));
      setDraft(null);
    } catch (err) {
      setError(errorMessage(err));
    }
  }

  if (!data) return null;
  return (
    <section className="card">
      <div className="page-header">
        <h2>My notes</h2>
        {draft === null && (
          <button className="btn small" onClick={() => setDraft(data.notes)}>
            {data.notes ? 'Edit' : 'Add notes'}
          </button>
        )}
      </div>
      {error && <p className="alert">{error}</p>}
      {draft === null ? (
        data.notes ? (
          <p className="notes-text">{data.notes}</p>
        ) : (
          <p className="muted small">Cues, seat settings, grip width… anything to remember.</p>
        )
      ) : (
        <div className="form">
          <textarea
            rows={3}
            maxLength={1000}
            value={draft}
            onChange={(e) => setDraft(e.target.value)}
            aria-label="Exercise notes"
          />
          <div className="form-actions">
            <button className="btn" onClick={() => setDraft(null)}>
              Cancel
            </button>
            <button className="btn primary" onClick={save}>
              Save
            </button>
          </div>
        </div>
      )}
    </section>
  );
}

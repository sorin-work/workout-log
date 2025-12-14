import { useState } from 'react';
import type { Exercise } from '../api/types';
import { useFetch } from '../lib/useFetch';

interface Props {
  onPick(e: Exercise): void;
  onCancel(): void;
}

export function ExercisePicker({ onPick, onCancel }: Props) {
  const [q, setQ] = useState('');
  const { data, loading } = useFetch<Exercise[]>(`/exercises?q=${encodeURIComponent(q.trim())}`);
  return (
    <div className="card picker">
      <div className="picker-head">
        <input
          type="search"
          autoFocus
          placeholder="Find an exercise"
          value={q}
          onChange={(e) => setQ(e.target.value)}
        />
        <button className="btn" onClick={onCancel}>
          Cancel
        </button>
      </div>
      {loading && !data && <p className="muted">Loading…</p>}
      {data && data.length === 0 && <p className="muted">Nothing found.</p>}
      <ul className="list picker-list">
        {data?.map((e) => (
          <li key={e.id}>
            <button className="picker-item" onClick={() => onPick(e)}>
              {e.name}
              <span className="muted small">{e.primaryMuscle}</span>
            </button>
          </li>
        ))}
      </ul>
    </div>
  );
}

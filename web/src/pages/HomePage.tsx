import { useState } from 'react';
import { useNavigate } from 'react-router';
import { errorMessage, post } from '../api/client';
import type { Workout } from '../api/types';
import { useAuth } from '../auth/context';

export default function HomePage() {
  const { user } = useAuth();
  const navigate = useNavigate();
  const [error, setError] = useState('');
  const [busy, setBusy] = useState(false);

  async function start() {
    setBusy(true);
    try {
      const w = await post<Workout>('/workouts', {});
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
      <button className="btn primary" onClick={start} disabled={busy}>
        Start an empty workout
      </button>
    </>
  );
}

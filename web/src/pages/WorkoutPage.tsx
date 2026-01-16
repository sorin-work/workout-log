import { useState } from 'react';
import { Link, useNavigate, useParams } from 'react-router';
import { del, errorMessage, post, put } from '../api/client';
import type { Exercise, FinishResult, Workout, WorkoutSet } from '../api/types';
import { ExercisePicker } from '../components/ExercisePicker';
import { SetRow, type SetChanges } from '../components/SetRow';
import { WorkoutDetailsForm } from '../components/WorkoutDetailsForm';
import { formatDate, formatDuration, formatTime } from '../lib/format';
import { groupSets } from '../lib/groupSets';
import { useFetch } from '../lib/useFetch';
import { useUnit } from '../lib/useUnit';

export default function WorkoutPage() {
  const { id } = useParams();
  const { data: workout, error, loading, setData } = useFetch<Workout>(`/workouts/${id}`);
  const navigate = useNavigate();
  const unit = useUnit();
  const [picking, setPicking] = useState(false);
  const [editing, setEditing] = useState(false);
  const [actionError, setActionError] = useState('');

  if (loading && !workout) return <p className="muted">Loading…</p>;
  if (error && !workout) return <p className="alert">{error}</p>;
  if (!workout) return null;
  const w = workout;

  async function run(fn: () => Promise<void>) {
    setActionError('');
    try {
      await fn();
    } catch (err) {
      setActionError(errorMessage(err));
    }
  }

  function replaceSet(updated: WorkoutSet) {
    setData({ ...w, sets: w.sets.map((s) => (s.id === updated.id ? updated : s)) });
  }

  const addSet = (exerciseId: number, template?: WorkoutSet) =>
    run(async () => {
      const created = await post<WorkoutSet>(`/workouts/${w.id}/sets`, {
        exerciseId,
        reps: template?.reps ?? 5,
        weightKg: template?.weightKg ?? 0,
        completed: false,
      });
      setData({ ...w, sets: [...w.sets, created] });
    });

  const updateSet = (set: WorkoutSet, changes: SetChanges) =>
    run(async () => {
      replaceSet(
        await put<WorkoutSet>(`/workouts/${w.id}/sets/${set.id}`, {
          exerciseId: set.exerciseId,
          ...changes,
        }),
      );
    });

  const deleteSet = (set: WorkoutSet) =>
    run(async () => {
      await del(`/workouts/${w.id}/sets/${set.id}`);
      setData({ ...w, sets: w.sets.filter((s) => s.id !== set.id) });
    });

  const finish = () =>
    run(async () => {
      const res = await post<FinishResult>(`/workouts/${w.id}/finish`);
      setData(res.workout);
    });

  const remove = () =>
    run(async () => {
      if (!window.confirm('Delete this workout and all of its sets?')) return;
      await del(`/workouts/${w.id}`);
      navigate('/workouts', { replace: true });
    });

  function pick(e: Exercise) {
    setPicking(false);
    addSet(e.id);
  }

  const groups = groupSets(w.sets);
  return (
    <>
      <div className="page-header">
        <div>
          <h1>{w.name}</h1>
          <div className="muted small">
            {formatDate(w.startedAt)} · {formatTime(w.startedAt)}
            {w.finishedAt && ` · ${formatDuration(w.startedAt, w.finishedAt)}`}
          </div>
        </div>
        <div className="actions">
          <button className="btn" onClick={() => setEditing(true)}>
            Edit
          </button>
          {!w.finishedAt && (
            <button className="btn primary" onClick={finish}>
              Finish workout
            </button>
          )}
        </div>
      </div>
      {actionError && <p className="alert">{actionError}</p>}
      {editing ? (
        <WorkoutDetailsForm
          workout={w}
          onCancel={() => setEditing(false)}
          onSaved={(saved) => {
            setData(saved);
            setEditing(false);
          }}
        />
      ) : (
        w.notes && (
          <div
            className="card notes"
            dangerouslySetInnerHTML={{ __html: w.notes.replace(/\n/g, '<br>') }}
          />
        )
      )}

      {groups.length === 0 && <p className="empty">No exercises yet. Add one to get started.</p>}
      {groups.map((g) => {
        let working = 0;
        return (
          <section key={g.exerciseId} className="card exercise-block">
            <h2>{g.exerciseName}</h2>
            <table className="sets">
              <thead>
                <tr>
                  <th>Set</th>
                  <th>{unit}</th>
                  <th>Reps</th>
                  <th>RPE</th>
                  <th>Done</th>
                  <th />
                </tr>
              </thead>
              <tbody>
                {g.sets.map((s) => (
                  <SetRow
                    key={s.id}
                    set={s}
                    index={s.warmup ? 0 : ++working}
                    unit={unit}
                    onChange={(c) => updateSet(s, c)}
                    onDelete={() => deleteSet(s)}
                  />
                ))}
              </tbody>
            </table>
            <button className="btn small" onClick={() => addSet(g.exerciseId, g.sets.at(-1))}>
              Add set
            </button>
          </section>
        );
      })}

      {picking ? (
        <ExercisePicker onPick={pick} onCancel={() => setPicking(false)} />
      ) : (
        <button className="btn" onClick={() => setPicking(true)}>
          Add exercise
        </button>
      )}
      <div className="page-footer">
        <Link to="/workouts">Back to workouts</Link>
        <button className="btn danger small" onClick={remove}>
          Delete workout
        </button>
      </div>
    </>
  );
}

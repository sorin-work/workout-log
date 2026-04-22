import { useState } from 'react';
import { Link, useNavigate, useParams } from 'react-router';
import { del, errorMessage, post, put } from '../api/client';
import type {
  Exercise,
  FinishResult,
  ProgressionChange,
  Template,
  Workout,
  WorkoutSet,
} from '../api/types';
import { ExercisePicker } from '../components/ExercisePicker';
import { LastTime } from '../components/LastTime';
import { PlateCalculator } from '../components/PlateCalculator';
import { RestTimer } from '../components/RestTimer';
import { ProgressionSummary } from '../components/ProgressionSummary';
import { SetRow, type SetChanges } from '../components/SetRow';
import { WorkoutDetailsForm } from '../components/WorkoutDetailsForm';
import { WorkoutNotes } from '../components/WorkoutNotes';
import { formatDate, formatDuration, formatTime } from '../lib/format';
import { groupSets } from '../lib/groupSets';
import { fromKg } from '../lib/units';
import { useFetch } from '../lib/useFetch';
import { useUnit } from '../lib/useUnit';

const DEFAULT_REST = 120;

export default function WorkoutPage() {
  const { id } = useParams();
  const { data: workout, error, loading, setData } = useFetch<Workout>(`/workouts/${id}`);
  const navigate = useNavigate();
  const unit = useUnit();
  const [picking, setPicking] = useState(false);
  const [editing, setEditing] = useState(false);
  const [restStart, setRestStart] = useState<number | null>(null);
  const [restSeconds, setRestSeconds] = useState(DEFAULT_REST);
  const template = useFetch<Template>(
    workout?.templateId ? `/templates/${workout.templateId}` : null,
  );
  const [plates, setPlates] = useState<number | null>(null);
  const [progression, setProgression] = useState<ProgressionChange[] | null>(null);
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
      if (changes.completed && !set.completed && !changes.warmup) {
        const planned = template.data?.exercises.find((e) => e.exerciseId === set.exerciseId);
        setRestSeconds(planned?.restSeconds ?? DEFAULT_REST);
        setRestStart(Date.now());
      }
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
      const open = w.sets.filter((s) => !s.completed).length;
      if (open > 0 && !window.confirm(`${open} sets are not ticked off. Finish anyway?`)) return;
      const res = await post<FinishResult>(`/workouts/${w.id}/finish`);
      setData(res.workout);
      setProgression(res.progression);
    });

  const repeat = () =>
    run(async () => {
      const copy = await post<Workout>(`/workouts/${w.id}/repeat`);
      navigate(`/workouts/${copy.id}`);
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
          {w.finishedAt ? (
            <button className="btn" onClick={repeat}>
              Do again
            </button>
          ) : (
            <button className="btn primary" onClick={finish}>
              Finish workout
            </button>
          )}
        </div>
      </div>
      {actionError && <p className="alert">{actionError}</p>}
      {progression && (
        <ProgressionSummary
          changes={progression}
          unit={unit}
          onClose={() => setProgression(null)}
        />
      )}
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
        <WorkoutNotes notes={w.notes} />
      )}

      {groups.length === 0 && <p className="empty">No exercises yet. Add one to get started.</p>}
      {groups.map((g) => {
        let working = 0;
        return (
          <section key={g.exerciseId} className="card exercise-block">
            <h2>
              <Link to={`/exercises/${g.exerciseId}`}>{g.exerciseName}</Link>
            </h2>
            <LastTime exerciseId={g.exerciseId} workoutId={w.id} unit={unit} />
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
            <div className="actions">
              <button className="btn small" onClick={() => addSet(g.exerciseId, g.sets.at(-1))}>
                Add set
              </button>
              <button
                className="btn small"
                onClick={() => setPlates(fromKg(g.sets.at(-1)?.weightKg ?? 0, unit))}
              >
                Plates
              </button>
            </div>
          </section>
        );
      })}

      {restStart !== null && !w.finishedAt && (
        <RestTimer
          key={restStart}
          startedAt={restStart}
          seconds={restSeconds}
          onDismiss={() => setRestStart(null)}
        />
      )}
      {plates !== null && (
        <PlateCalculator initial={plates} unit={unit} onClose={() => setPlates(null)} />
      )}
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

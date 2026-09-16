import { useState } from 'react';
import { Link, useNavigate, useParams } from 'react-router';
import { del, errorMessage, get, post, put } from '../api/client';
import type { Exercise, FinishResult, Template, Workout, WorkoutSet } from '../api/types';
import { Elapsed } from '../components/Elapsed';
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
import { restOverAlert } from '../lib/alert';
import { estimateOneRepMax } from '../lib/e1rm';
import { formatWeight, fromKg } from '../lib/units';
import { useAuth } from '../auth/context';
import { useFetch } from '../lib/useFetch';
import { useTitle } from '../lib/useTitle';
import { useUnit } from '../lib/useUnit';

const DEFAULT_REST = 120;

export default function WorkoutPage() {
  const { id } = useParams();
  const { data: workout, error, loading, setData } = useFetch<Workout>(`/workouts/${id}`);
  const navigate = useNavigate();
  const unit = useUnit();
  const formula = useAuth().user?.e1rmFormula ?? 'epley';
  useTitle(workout?.name);
  const [picking, setPicking] = useState(false);
  const [editing, setEditing] = useState(false);
  const [restStart, setRestStart] = useState<number | null>(null);
  const [restSeconds, setRestSeconds] = useState(DEFAULT_REST);
  const template = useFetch<Template>(
    workout?.templateId ? `/templates/${workout.templateId}` : null,
  );
  const [plates, setPlates] = useState<number | null>(null);
  const [finished, setFinished] = useState<FinishResult | null>(null);
  const [finishing, setFinishing] = useState(false);
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

  const addWarmups = (exerciseId: number) =>
    run(async () => {
      await post<WorkoutSet[]>(`/workouts/${w.id}/warmups`, { exerciseId });
      // Later sets moved down, so reload the whole workout.
      setData(await get<Workout>(`/workouts/${w.id}`));
    });

  const moveExercise = (index: number, delta: number) =>
    run(async () => {
      const order = groupSets(w.sets);
      const [g] = order.splice(index, 1);
      order.splice(index + delta, 0, g);
      const setIds = order.flatMap((x) => x.sets.map((s) => s.id));
      setData(await put<Workout>(`/workouts/${w.id}/order`, { setIds }));
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
      setFinishing(true);
      try {
        const res = await post<FinishResult>(`/workouts/${w.id}/finish`);
        setData(res.workout);
        setFinished(res);
      } finally {
        setFinishing(false);
      }
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
            {w.finishedAt ? (
              ` · ${formatDuration(w.startedAt, w.finishedAt)}`
            ) : (
              <>
                {' · '}
                <Elapsed since={w.startedAt} /> so far
              </>
            )}
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
            <button className="btn primary" onClick={finish} disabled={finishing}>
              {finishing ? 'Finishing…' : 'Finish workout'}
            </button>
          )}
        </div>
      </div>
      {actionError && <p className="alert">{actionError}</p>}
      {finished && (
        <ProgressionSummary
          changes={finished.progression}
          records={finished.records}
          unit={unit}
          onClose={() => setFinished(null)}
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
      {groups.map((g, gi) => {
        let working = 0;
        return (
          <section key={g.exerciseId} className="card exercise-block">
            <div className="page-header">
              <h2>
                <Link to={`/exercises/${g.exerciseId}`}>{g.exerciseName}</Link>
              </h2>
              {groups.length > 1 && (
                <div className="row-actions">
                  <button
                    className="icon-btn"
                    aria-label={`Move ${g.exerciseName} up`}
                    disabled={gi === 0}
                    onClick={() => moveExercise(gi, -1)}
                  >
                    ↑
                  </button>
                  <button
                    className="icon-btn"
                    aria-label={`Move ${g.exerciseName} down`}
                    disabled={gi === groups.length - 1}
                    onClick={() => moveExercise(gi, 1)}
                  >
                    ↓
                  </button>
                </div>
              )}
            </div>
            <LastTime exerciseId={g.exerciseId} workoutId={w.id} unit={unit} />
            {(() => {
              const best = Math.max(
                0,
                ...g.sets
                  .filter((s) => s.completed && !s.warmup)
                  .map((s) => estimateOneRepMax(s.weightKg, s.reps, formula)),
              );
              return best > 0 ? (
                <p className="muted small last-time">
                  Best today: est. 1RM {formatWeight(best, unit)}
                </p>
              ) : null;
            })()}
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
              {!g.sets.some((s) => s.warmup) && g.sets.some((s) => s.weightKg > 20) && (
                <button className="btn small" onClick={() => addWarmups(g.exerciseId)}>
                  Add warm-ups
                </button>
              )}
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
          onDone={restOverAlert}
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

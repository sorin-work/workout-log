import { useEffect, useState, type FormEvent } from 'react';
import { useNavigate, useParams } from 'react-router';
import { ApiError, errorMessage, post, put } from '../api/client';
import type { Exercise, Template } from '../api/types';
import { ExercisePicker } from '../components/ExercisePicker';
import { FieldError } from '../components/FieldError';
import { fromKg, toKg, trimNumber, type Unit } from '../lib/units';
import { useFetch } from '../lib/useFetch';
import { useUnit } from '../lib/useUnit';

interface Row {
  key: string;
  id?: number;
  exerciseId: number;
  exerciseName: string;
  sets: string;
  reps: string;
  weight: string;
  rest: string;
  increment: string;
  deloadAfter: string;
  deloadPercent: string;
}

let nextKey = 0;
const newKey = () => `row-${++nextKey}`;

function toRows(t: Template, unit: Unit): Row[] {
  return t.exercises.map((e) => ({
    key: newKey(),
    id: e.id,
    exerciseId: e.exerciseId,
    exerciseName: e.exerciseName,
    sets: String(e.sets),
    reps: String(e.reps),
    weight: trimNumber(fromKg(e.weightKg, unit), 2),
    rest: String(e.restSeconds),
    increment: trimNumber(fromKg(e.incrementKg, unit), 2),
    deloadAfter: String(e.deloadAfter),
    deloadPercent: String(e.deloadPercent),
  }));
}

export default function TemplateEditPage() {
  const { id } = useParams();
  const { data, error, loading } = useFetch<Template>(id ? `/templates/${id}` : null);
  if (id && loading) return <p className="muted">Loading…</p>;
  if (id && error) return <p className="alert">{error}</p>;
  if (id && !data) return null;
  return <TemplateForm key={data?.id ?? 'new'} template={data} />;
}

function TemplateForm({ template }: { template?: Template }) {
  const unit = useUnit();
  const navigate = useNavigate();
  const [name, setName] = useState(template?.name ?? '');
  const [notes, setNotes] = useState(template?.notes ?? '');
  const [rows, setRows] = useState<Row[]>(template ? toRows(template, unit) : []);
  const [picking, setPicking] = useState(false);
  const [fields, setFields] = useState<Record<string, string>>({});
  const [error, setError] = useState('');
  const [busy, setBusy] = useState(false);
  const [dirty, setDirty] = useState(false);

  // Warn before closing the tab with unsaved changes.
  useEffect(() => {
    if (!dirty) return;
    const onBeforeUnload = (e: BeforeUnloadEvent) => e.preventDefault();
    window.addEventListener('beforeunload', onBeforeUnload);
    return () => window.removeEventListener('beforeunload', onBeforeUnload);
  }, [dirty]);

  function update(key: string, patch: Partial<Row>) {
    setDirty(true);
    setRows((rs) => rs.map((r) => (r.key === key ? { ...r, ...patch } : r)));
  }

  function move(index: number, delta: number) {
    setRows((rs) => {
      const next = [...rs];
      const [r] = next.splice(index, 1);
      next.splice(index + delta, 0, r);
      return next;
    });
  }

  function add(e: Exercise) {
    setPicking(false);
    setDirty(true);
    setRows((rs) => [
      ...rs,
      {
        key: newKey(),
        exerciseId: e.id,
        exerciseName: e.name,
        sets: '3',
        reps: '8',
        weight: '0',
        rest: '120',
        increment: '0',
        deloadAfter: '0',
        deloadPercent: '10',
      },
    ]);
  }

  async function onSubmit(e: FormEvent) {
    e.preventDefault();
    setBusy(true);
    setFields({});
    setError('');
    const body = {
      name,
      notes,
      exercises: rows.map((r) => ({
        id: r.id,
        exerciseId: r.exerciseId,
        sets: Number(r.sets),
        reps: Number(r.reps),
        weightKg: toKg(Number(r.weight.replace(',', '.')) || 0, unit),
        restSeconds: Number(r.rest),
        incrementKg: toKg(Number(r.increment.replace(',', '.')) || 0, unit),
        deloadAfter: Number(r.deloadAfter) || 0,
        deloadPercent: Number(r.deloadPercent) || 0,
      })),
    };
    try {
      if (template) await put<Template>(`/templates/${template.id}`, body);
      else await post<Template>('/templates', body);
      setDirty(false);
      navigate('/templates');
    } catch (err) {
      if (err instanceof ApiError && err.code === 'validation_failed') setFields(err.fields);
      else setError(errorMessage(err));
      setBusy(false);
    }
  }

  return (
    <form className="form" onSubmit={onSubmit}>
      <h1>{template ? `Edit ${template.name}` : 'New template'}</h1>
      {error && <p className="alert">{error}</p>}
      <div className="card form">
        <label>
          Name
          <input value={name} onChange={(e) => setName(e.target.value)} required maxLength={60} />
          <FieldError message={fields.name} />
        </label>
        <label>
          Notes
          <textarea rows={2} value={notes} onChange={(e) => setNotes(e.target.value)} />
        </label>
      </div>

      <div className="card">
        {rows.length === 0 && <p className="muted">Add the exercises this workout is made of.</p>}
        {rows.length > 0 && (
          <table className="sets template-rows">
            <thead>
              <tr>
                <th>Exercise</th>
                <th>Sets</th>
                <th>Reps</th>
                <th>{unit}</th>
                <th>Rest (s)</th>
                <th title="Added after every session where all sets were completed">+{unit}</th>
                <th title="Deload after this many failed sessions in a row (0 = never)">
                  Deload after
                </th>
                <th>Deload %</th>
                <th />
              </tr>
            </thead>
            <tbody>
              {rows.map((r, i) => (
                <tr key={r.key}>
                  <td>
                    {r.exerciseName}
                    <FieldError message={fields[`exercises.${i}.exerciseId`]} />
                  </td>
                  <td>
                    <input
                      aria-label="Sets"
                      inputMode="numeric"
                      value={r.sets}
                      onChange={(e) => update(r.key, { sets: e.target.value })}
                    />
                    <FieldError message={fields[`exercises.${i}.sets`]} />
                  </td>
                  <td>
                    <input
                      aria-label="Reps"
                      inputMode="numeric"
                      value={r.reps}
                      onChange={(e) => update(r.key, { reps: e.target.value })}
                    />
                    <FieldError message={fields[`exercises.${i}.reps`]} />
                  </td>
                  <td>
                    <input
                      aria-label="Weight"
                      inputMode="decimal"
                      value={r.weight}
                      onChange={(e) => update(r.key, { weight: e.target.value })}
                    />
                    <FieldError message={fields[`exercises.${i}.weightKg`]} />
                  </td>
                  <td>
                    <input
                      aria-label="Rest"
                      inputMode="numeric"
                      value={r.rest}
                      onChange={(e) => update(r.key, { rest: e.target.value })}
                    />
                  </td>
                  <td>
                    <input
                      aria-label="Increment"
                      inputMode="decimal"
                      value={r.increment}
                      onChange={(e) => update(r.key, { increment: e.target.value })}
                    />
                    <FieldError message={fields[`exercises.${i}.incrementKg`]} />
                  </td>
                  <td>
                    <input
                      aria-label="Deload after failures"
                      inputMode="numeric"
                      value={r.deloadAfter}
                      onChange={(e) => update(r.key, { deloadAfter: e.target.value })}
                    />
                    <FieldError message={fields[`exercises.${i}.deloadAfter`]} />
                  </td>
                  <td>
                    <input
                      aria-label="Deload percent"
                      inputMode="decimal"
                      value={r.deloadPercent}
                      disabled={r.deloadAfter === '0' || r.deloadAfter === ''}
                      onChange={(e) => update(r.key, { deloadPercent: e.target.value })}
                    />
                    <FieldError message={fields[`exercises.${i}.deloadPercent`]} />
                  </td>
                  <td className="row-actions">
                    <button
                      type="button"
                      className="icon-btn"
                      aria-label="Move up"
                      disabled={i === 0}
                      onClick={() => move(i, -1)}
                    >
                      ↑
                    </button>
                    <button
                      type="button"
                      className="icon-btn"
                      aria-label="Move down"
                      disabled={i === rows.length - 1}
                      onClick={() => move(i, 1)}
                    >
                      ↓
                    </button>
                    <button
                      type="button"
                      className="icon-btn"
                      aria-label="Remove"
                      onClick={() => setRows((rs) => rs.filter((x) => x.key !== r.key))}
                    >
                      ×
                    </button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
        {picking ? (
          <ExercisePicker onPick={add} onCancel={() => setPicking(false)} />
        ) : (
          <button type="button" className="btn" onClick={() => setPicking(true)}>
            Add exercise
          </button>
        )}
      </div>

      <div className="form-actions">
        <button
          type="button"
          className="btn"
          onClick={() => {
            if (!dirty || window.confirm('Discard your changes to this template?'))
              navigate('/templates');
          }}
        >
          Cancel
        </button>
        <button className="btn primary" disabled={busy}>
          {busy ? 'Saving…' : 'Save template'}
        </button>
      </div>
    </form>
  );
}

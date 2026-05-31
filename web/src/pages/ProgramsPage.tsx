import { useState, type FormEvent } from 'react';
import { ApiError, del, errorMessage, post, put } from '../api/client';
import type { Program, Template } from '../api/types';
import { FieldError } from '../components/FieldError';
import { useFetch } from '../lib/useFetch';
import { useTitle } from '../lib/useTitle';

interface FormProps {
  program?: Program;
  templates: Template[];
  onDone(): void;
}

function ProgramForm({ program, templates, onDone }: FormProps) {
  const [name, setName] = useState(program?.name ?? '');
  const [days, setDays] = useState<number[]>(program?.days.map((d) => d.templateId) ?? []);
  const [fields, setFields] = useState<Record<string, string>>({});
  const [error, setError] = useState('');
  const nameOf = (id: number) => templates.find((t) => t.id === id)?.name ?? '?';

  async function onSubmit(e: FormEvent) {
    e.preventDefault();
    setFields({});
    setError('');
    const body = { name, templateIds: days };
    try {
      if (program) await put(`/programs/${program.id}`, body);
      else await post('/programs', body);
      onDone();
    } catch (err) {
      if (err instanceof ApiError && err.code === 'validation_failed') setFields(err.fields);
      else setError(errorMessage(err));
    }
  }

  return (
    <form className="card form" onSubmit={onSubmit}>
      <h2>{program ? `Edit ${program.name}` : 'New program'}</h2>
      {error && <p className="alert">{error}</p>}
      <label>
        Name
        <input value={name} onChange={(e) => setName(e.target.value)} maxLength={60} />
        <FieldError message={fields.name} />
      </label>
      <div>
        <div className="label">Rotation</div>
        <ol className="rotation">
          {days.map((id, i) => (
            <li key={`${id}-${i}`}>
              {nameOf(id)}
              <button
                type="button"
                className="icon-btn"
                aria-label="Remove"
                onClick={() => setDays(days.filter((_, j) => j !== i))}
              >
                ×
              </button>
            </li>
          ))}
        </ol>
        <div className="chips">
          {templates.map((t) => (
            <button
              key={t.id}
              type="button"
              className="btn small"
              onClick={() => setDays([...days, t.id])}
            >
              + {t.name}
            </button>
          ))}
        </div>
        <FieldError message={fields.templateIds} />
      </div>
      <div className="form-actions">
        <button type="button" className="btn" onClick={onDone}>
          Cancel
        </button>
        <button className="btn primary">Save</button>
      </div>
    </form>
  );
}

export default function ProgramsPage() {
  useTitle('Programs');
  const programs = useFetch<Program[]>('/programs');
  const templates = useFetch<Template[]>('/templates');
  const [editing, setEditing] = useState<Program | 'new' | null>(null);
  const [error, setError] = useState('');

  async function act(fn: () => Promise<unknown>) {
    setError('');
    try {
      await fn();
      programs.reload();
    } catch (err) {
      setError(errorMessage(err));
    }
  }

  const done = () => {
    setEditing(null);
    programs.reload();
  };

  return (
    <>
      <div className="page-header">
        <h1>Programs</h1>
        {!editing && (
          <button className="btn primary" onClick={() => setEditing('new')}>
            New program
          </button>
        )}
      </div>
      <p className="muted">
        A program rotates through your templates, so the home page always knows which workout is
        next.
      </p>
      {(error || programs.error) && <p className="alert">{error || programs.error}</p>}
      {editing && templates.data && (
        <ProgramForm
          program={editing === 'new' ? undefined : editing}
          templates={templates.data}
          onDone={done}
        />
      )}
      {programs.data?.length === 0 && !editing && (
        <p className="empty">No programs yet. Create templates first, then put them in order.</p>
      )}
      <div className="grid">
        {programs.data?.map((p) => (
          <section
            key={p.id}
            className={p.active ? 'card template-card active' : 'card template-card'}
          >
            <h2>
              {p.name} {p.active && <span className="badge">Active</span>}
            </h2>
            <ol className="rotation">
              {p.days.map((d) => (
                <li key={d.position} className={d.position === p.nextPosition ? 'next' : ''}>
                  {d.templateName}
                  {d.position === p.nextPosition && <span className="muted small"> · next</span>}
                </li>
              ))}
            </ol>
            <div className="actions">
              {!p.active && (
                <button
                  className="btn small primary"
                  onClick={() => act(() => post(`/programs/${p.id}/activate`))}
                >
                  Follow
                </button>
              )}
              <button className="btn small" onClick={() => setEditing(p)}>
                Edit
              </button>
              <button
                className="btn small danger"
                onClick={() =>
                  window.confirm(`Delete ${p.name}?`) && act(() => del(`/programs/${p.id}`))
                }
              >
                Delete
              </button>
            </div>
          </section>
        ))}
      </div>
    </>
  );
}

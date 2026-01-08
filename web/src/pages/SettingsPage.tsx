import { useMemo, useState, type FormEvent } from 'react';
import { ApiError, errorMessage, put } from '../api/client';
import type { User } from '../api/types';
import { useAuth } from '../auth/context';
import { FieldError } from '../components/FieldError';

function timeZones(): string[] {
  try {
    return Intl.supportedValuesOf('timeZone');
  } catch {
    return ['UTC'];
  }
}

export default function SettingsPage() {
  const { user, setUser } = useAuth();
  const [displayName, setDisplayName] = useState(user?.displayName ?? '');
  const [unit, setUnit] = useState(user?.unit ?? 'kg');
  const [timezone, setTimezone] = useState(user?.timezone ?? 'UTC');
  const [fields, setFields] = useState<Record<string, string>>({});
  const [error, setError] = useState('');
  const [saved, setSaved] = useState(false);
  const zones = useMemo(() => {
    const list = timeZones();
    return list.includes(timezone) ? list : [timezone, ...list];
  }, [timezone]);

  async function onSubmit(e: FormEvent) {
    e.preventDefault();
    setFields({});
    setError('');
    setSaved(false);
    try {
      setUser(await put<User>('/me', { displayName, unit, timezone }));
      setSaved(true);
    } catch (err) {
      if (err instanceof ApiError && err.code === 'validation_failed') setFields(err.fields);
      else setError(errorMessage(err));
    }
  }

  return (
    <>
      <h1>Settings</h1>
      <form className="card form narrow" onSubmit={onSubmit}>
        <h2>Profile</h2>
        {error && <p className="alert">{error}</p>}
        {saved && <p className="success">Settings saved.</p>}
        <label>
          Name
          <input value={displayName} onChange={(e) => setDisplayName(e.target.value)} />
          <FieldError message={fields.displayName} />
        </label>
        <fieldset className="chips">
          <legend>Weight unit</legend>
          {(['kg', 'lb'] as const).map((u) => (
            <label key={u} className="chip">
              <input type="radio" name="unit" checked={unit === u} onChange={() => setUnit(u)} />
              {u === 'kg' ? 'Kilograms' : 'Pounds'}
            </label>
          ))}
        </fieldset>
        <label>
          Time zone
          <select value={timezone} onChange={(e) => setTimezone(e.target.value)}>
            {zones.map((z) => (
              <option key={z} value={z}>
                {z}
              </option>
            ))}
          </select>
          <span className="muted small">Used to decide which week a workout belongs to.</span>
          <FieldError message={fields.timezone} />
        </label>
        <div className="form-actions">
          <button className="btn primary">Save</button>
        </div>
      </form>
    </>
  );
}

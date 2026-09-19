import { useMemo, useState, type FormEvent } from 'react';
import { ApiError, errorMessage, post, put } from '../api/client';
import type { User } from '../api/types';
import { useAuth } from '../auth/context';
import { FieldError } from '../components/FieldError';
import { ImportCsv } from '../components/ImportCsv';
import { useFetch } from '../lib/useFetch';
import { DeleteAccount } from '../components/DeleteAccount';
import { PasswordForm } from '../components/PasswordForm';
import { useTitle } from '../lib/useTitle';

function timeZones(): string[] {
  try {
    return Intl.supportedValuesOf('timeZone');
  } catch {
    return ['UTC'];
  }
}

export default function SettingsPage() {
  useTitle('Settings');
  const { user, setUser } = useAuth();
  const [displayName, setDisplayName] = useState(user?.displayName ?? '');
  const [unit, setUnit] = useState(user?.unit ?? 'kg');
  const [timezone, setTimezone] = useState(user?.timezone ?? 'UTC');
  const [e1rmFormula, setFormula] = useState(user?.e1rmFormula ?? 'epley');
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
      setUser(await put<User>('/me', { displayName, unit, timezone, e1rmFormula }));
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
        <label>
          Estimated 1RM formula
          <select
            value={e1rmFormula}
            onChange={(e) => setFormula(e.target.value as 'epley' | 'brzycki')}
          >
            <option value="epley">Epley: weight × (1 + reps / 30)</option>
            <option value="brzycki">Brzycki: weight × 36 / (37 − reps)</option>
          </select>
          <span className="muted small">Brzycki is a little more conservative above 5 reps.</span>
          <FieldError message={fields.e1rmFormula} />
        </label>
        <div className="form-actions">
          <button className="btn primary">Save</button>
        </div>
      </form>
      <div className="card narrow export">
        <h2>Your data</h2>
        <p className="muted">
          Download every set you have logged as a CSV file (weights in kg), or import one in the
          same format. Workouts already in your log are skipped.
        </p>
        <a className="btn" href="/api/export/sets.csv" download>
          Export CSV
        </a>
        <ImportCsv />
      </div>
      <PasswordForm />
      <OtherDevices />
      <DeleteAccount />
      <AppVersion />
    </>
  );
}

function AppVersion() {
  const { data } = useFetch<{ version: string }>('/version');
  if (!data) return null;
  return <p className="muted small center version">Workout Log {data.version}</p>;
}

function OtherDevices() {
  const [message, setMessage] = useState('');
  const [error, setError] = useState('');

  async function logoutOthers() {
    setError('');
    try {
      const res = await post<{ loggedOut: number }>('/me/logout-others');
      setMessage(
        res.loggedOut === 0
          ? 'No other devices were logged in.'
          : `Logged out ${res.loggedOut} other session${res.loggedOut === 1 ? '' : 's'}.`,
      );
    } catch (err) {
      setError(errorMessage(err));
    }
  }

  return (
    <div className="card narrow export">
      <h2>Devices</h2>
      <p className="muted">Lost a phone or used a gym computer? End every session but this one.</p>
      {message && <p className="success">{message}</p>}
      {error && <p className="alert">{error}</p>}
      <button className="btn" onClick={logoutOthers}>
        Log out other devices
      </button>
    </div>
  );
}

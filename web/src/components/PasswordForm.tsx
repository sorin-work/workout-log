import { useState, type FormEvent } from 'react';
import { ApiError, errorMessage, put } from '../api/client';
import { FieldError } from './FieldError';

export function PasswordForm() {
  const [currentPassword, setCurrent] = useState('');
  const [newPassword, setNew] = useState('');
  const [confirm, setConfirm] = useState('');
  const [fields, setFields] = useState<Record<string, string>>({});
  const [error, setError] = useState('');
  const [saved, setSaved] = useState(false);

  async function onSubmit(e: FormEvent) {
    e.preventDefault();
    setFields({});
    setError('');
    setSaved(false);
    if (newPassword !== confirm) {
      setFields({ confirm: 'Passwords do not match.' });
      return;
    }
    try {
      await put('/me/password', { currentPassword, newPassword });
      setSaved(true);
      setCurrent('');
      setNew('');
      setConfirm('');
    } catch (err) {
      if (err instanceof ApiError && err.code === 'validation_failed') setFields(err.fields);
      else setError(errorMessage(err));
    }
  }

  return (
    <form className="card form narrow" onSubmit={onSubmit}>
      <h2>Password</h2>
      {error && <p className="alert">{error}</p>}
      {saved && <p className="success">Password changed. Other devices were logged out.</p>}
      <label>
        Current password
        <input
          type="password"
          autoComplete="current-password"
          value={currentPassword}
          onChange={(e) => setCurrent(e.target.value)}
          required
        />
        <FieldError message={fields.currentPassword} />
      </label>
      <label>
        New password
        <input
          type="password"
          autoComplete="new-password"
          minLength={8}
          value={newPassword}
          onChange={(e) => setNew(e.target.value)}
          required
        />
        <FieldError message={fields.newPassword} />
      </label>
      <label>
        Repeat new password
        <input
          type="password"
          autoComplete="new-password"
          value={confirm}
          onChange={(e) => setConfirm(e.target.value)}
          required
        />
        <FieldError message={fields.confirm} />
      </label>
      <div className="form-actions">
        <button className="btn primary">Change password</button>
      </div>
    </form>
  );
}

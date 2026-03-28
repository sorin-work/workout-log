import { useState, type FormEvent } from 'react';
import { useNavigate } from 'react-router';
import { ApiError, del, errorMessage } from '../api/client';
import { useAuth } from '../auth/context';
import { FieldError } from './FieldError';

export function DeleteAccount() {
  const { logout } = useAuth();
  const navigate = useNavigate();
  const [open, setOpen] = useState(false);
  const [password, setPassword] = useState('');
  const [error, setError] = useState('');
  const [fieldError, setFieldError] = useState('');

  async function onSubmit(e: FormEvent) {
    e.preventDefault();
    setError('');
    setFieldError('');
    try {
      await del('/me', { password });
      await logout().catch(() => undefined);
      navigate('/login', { replace: true });
    } catch (err) {
      if (err instanceof ApiError && err.fields.password) setFieldError(err.fields.password);
      else setError(errorMessage(err));
    }
  }

  if (!open) {
    return (
      <div className="card narrow danger-zone">
        <h2>Delete account</h2>
        <p className="muted">Removes your account, workouts, templates and records for good.</p>
        <button className="btn danger" onClick={() => setOpen(true)}>
          Delete my account
        </button>
      </div>
    );
  }
  return (
    <form className="card form narrow danger-zone" onSubmit={onSubmit}>
      <h2>Delete account</h2>
      <p>This cannot be undone. Enter your password to confirm.</p>
      {error && <p className="alert">{error}</p>}
      <label>
        Password
        <input
          type="password"
          autoComplete="current-password"
          value={password}
          onChange={(e) => setPassword(e.target.value)}
          required
        />
        <FieldError message={fieldError} />
      </label>
      <div className="form-actions">
        <button type="button" className="btn" onClick={() => setOpen(false)}>
          Cancel
        </button>
        <button className="btn danger">Delete everything</button>
      </div>
    </form>
  );
}

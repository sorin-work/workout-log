import { useState, type ChangeEvent } from 'react';
import { ApiError, errorMessage } from '../api/client';

interface Result {
  workouts: number;
  sets: number;
  skipped: number;
}

/** Uploads a CSV in the export format. */
export function ImportCsv() {
  const [result, setResult] = useState<Result | null>(null);
  const [errors, setErrors] = useState<string[]>([]);
  const [busy, setBusy] = useState(false);

  async function onFile(e: ChangeEvent<HTMLInputElement>) {
    const file = e.target.files?.[0];
    e.target.value = '';
    if (!file) return;
    setBusy(true);
    setResult(null);
    setErrors([]);
    try {
      const res = await fetch('/api/import/sets.csv', {
        method: 'POST',
        headers: { 'Content-Type': 'text/csv' },
        body: await file.text(),
      });
      const data = await res.json();
      if (!res.ok) throw new ApiError(res.status, data.error, data.message, data.fields);
      setResult(data);
    } catch (err) {
      if (err instanceof ApiError && Object.keys(err.fields).length > 0) {
        setErrors(Object.entries(err.fields).map(([line, msg]) => `${line}: ${msg}`));
      } else {
        setErrors([errorMessage(err)]);
      }
    } finally {
      setBusy(false);
    }
  }

  return (
    <div>
      <label className="btn file-btn">
        {busy ? 'Importing…' : 'Import CSV'}
        <input type="file" accept=".csv,text/csv" onChange={onFile} disabled={busy} hidden />
      </label>
      {result && (
        <p className="success">
          Imported {result.workouts} workouts with {result.sets} sets
          {result.skipped > 0 && `, skipped ${result.skipped} already in your log`}.
        </p>
      )}
      {errors.length > 0 && (
        <div className="alert">
          Nothing was imported:
          <ul>
            {errors.map((e) => (
              <li key={e}>{e}</li>
            ))}
          </ul>
        </div>
      )}
    </div>
  );
}

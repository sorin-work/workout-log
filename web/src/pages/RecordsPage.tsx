import { useState } from 'react';
import { Link } from 'react-router';
import type { ExerciseRecords, RecordSet } from '../api/types';
import { formatDate } from '../lib/format';
import { formatWeight, type Unit } from '../lib/units';
import { Skeleton } from '../components/Skeleton';
import { useFetch } from '../lib/useFetch';
import { useUnit } from '../lib/useUnit';
import { useTitle } from '../lib/useTitle';

function SetCell({ r, unit }: { r: RecordSet | null; unit: Unit }) {
  if (!r) return <td className="muted">–</td>;
  return (
    <td>
      {formatWeight(r.weightKg, unit)} × {r.reps}
      <div className="muted small">
        <Link to={`/workouts/${r.workoutId}`}>{formatDate(r.performedAt)}</Link>
      </div>
    </td>
  );
}

export default function RecordsPage() {
  useTitle('Records');
  const unit = useUnit();
  const { data, error, loading } = useFetch<ExerciseRecords[]>('/records');
  const [q, setQ] = useState('');
  const shown = data?.filter((r) => r.exerciseName.toLowerCase().includes(q.trim().toLowerCase()));

  return (
    <>
      <div className="page-header">
        <h1>Personal records</h1>
      </div>
      {error && <p className="alert">{error}</p>}
      {loading && !data && <Skeleton />}
      {data && data.length === 0 && (
        <p className="empty">Finish a few sets and your records will show up here.</p>
      )}
      {data && data.length > 0 && (
        <>
          <div className="filters">
            <input
              type="search"
              placeholder="Filter exercises"
              value={q}
              onChange={(e) => setQ(e.target.value)}
            />
          </div>
          <div className="card">
            <table className="table">
              <thead>
                <tr>
                  <th>Exercise</th>
                  <th>Est. 1RM</th>
                  <th>Heaviest</th>
                  <th>Best set</th>
                  <th>Most reps</th>
                </tr>
              </thead>
              <tbody>
                {shown?.map((r) => (
                  <tr key={r.exerciseId}>
                    <td>
                      <Link to={`/exercises/${r.exerciseId}`}>{r.exerciseName}</Link>
                      <div className="muted small">{r.setCount} sets</div>
                    </td>
                    <td>
                      {r.estimatedMax ? (
                        <strong>{formatWeight(r.estimatedMax.value, unit)}</strong>
                      ) : (
                        '–'
                      )}
                    </td>
                    <SetCell r={r.heaviest} unit={unit} />
                    <SetCell r={r.bestVolume} unit={unit} />
                    <td>{r.mostReps ? r.mostReps.reps : '–'}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </>
      )}
    </>
  );
}

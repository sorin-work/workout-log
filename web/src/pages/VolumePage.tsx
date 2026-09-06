import { useState } from 'react';
import { MUSCLE_GROUPS, type WeekVolume } from '../api/types';
import { BarChart } from '../components/charts/BarChart';
import { capitalize } from '../lib/text';
import { fromKg, trimNumber } from '../lib/units';
import { useUnit } from '../lib/useUnit';
import { useFetch } from '../lib/useFetch';
import { useTitle } from '../lib/useTitle';

/** "2026-02-09" -> "9 Feb". */
function weekLabel(weekStart: string): string {
  const [y, m, d] = weekStart.split('-').map(Number);
  return new Date(y, m - 1, d).toLocaleDateString('en-GB', { day: 'numeric', month: 'short' });
}

/** Background for a heat-map cell; 10-20 hard sets a week is the usual target. */
function heat(sets: number): string {
  if (sets <= 0) return 'transparent';
  const alpha = Math.min(1, sets / 20);
  return `rgba(31, 111, 235, ${0.12 + alpha * 0.6})`;
}

export default function VolumePage() {
  useTitle('Weekly volume');
  const [weeks, setWeeks] = useState(12);
  const [muscle, setMuscle] = useState('');
  const [measure, setMeasure] = useState<'sets' | 'volume'>('sets');
  const unit = useUnit();
  const { data, error, loading } = useFetch<WeekVolume[]>(`/stats/volume?weeks=${weeks}`);

  const bars =
    data?.map((w) => ({
      label: weekLabel(w.weekStart),
      value:
        measure === 'sets'
          ? muscle
            ? (w.muscles[muscle]?.sets ?? 0)
            : w.totalSets
          : fromKg(
              muscle
                ? (w.muscles[muscle]?.volumeKg ?? 0)
                : Object.values(w.muscles).reduce((sum, m) => sum + m.volumeKg, 0),
              unit,
            ),
    })) ?? [];
  const trained = MUSCLE_GROUPS.filter((m) => data?.some((w) => (w.muscles[m]?.sets ?? 0) > 0));

  return (
    <>
      <div className="page-header">
        <h1>Weekly volume</h1>
        <select value={weeks} onChange={(e) => setWeeks(Number(e.target.value))}>
          <option value={4}>4 weeks</option>
          <option value={12}>12 weeks</option>
          <option value={26}>26 weeks</option>
          <option value={52}>52 weeks</option>
        </select>
      </div>
      <p className="muted">
        Hard sets per week. A set counts fully for the main muscle and half for the others it works.
      </p>
      {error && <p className="alert">{error}</p>}
      {loading && !data && <p className="muted">Loading…</p>}
      {data && (
        <>
          <section className="card">
            <div className="page-header">
              <h2>{muscle ? capitalize(muscle) : 'All muscles'}</h2>
              <select value={muscle} onChange={(e) => setMuscle(e.target.value)}>
                <option value="">All muscles</option>
                {MUSCLE_GROUPS.map((m) => (
                  <option key={m} value={m}>
                    {capitalize(m)}
                  </option>
                ))}
              </select>
            </div>
            <div className="segmented range">
              <button
                className={measure === 'sets' ? 'active' : ''}
                onClick={() => setMeasure('sets')}
              >
                Sets
              </button>
              <button
                className={measure === 'volume' ? 'active' : ''}
                onClick={() => setMeasure('volume')}
              >
                Tonnage ({unit})
              </button>
            </div>
            <BarChart
              bars={bars}
              ariaLabel={measure === 'sets' ? 'Sets per week' : 'Tonnage per week'}
              formatY={(v) =>
                measure === 'sets' ? trimNumber(v) : Math.round(v).toLocaleString('en-US')
              }
            />
          </section>
          {trained.length === 0 ? (
            <p className="empty">No training logged in this period.</p>
          ) : (
            <section className="card">
              <h2>By muscle group</h2>
              <table className="table heat">
                <thead>
                  <tr>
                    <th>Muscle</th>
                    {data.map((w) => (
                      <th key={w.weekStart}>{weekLabel(w.weekStart)}</th>
                    ))}
                  </tr>
                </thead>
                <tbody>
                  {trained.map((m) => (
                    <tr key={m}>
                      <td>{capitalize(m)}</td>
                      {data.map((w) => {
                        const sets = w.muscles[m]?.sets ?? 0;
                        return (
                          <td key={w.weekStart} style={{ background: heat(sets) }}>
                            {sets > 0 ? trimNumber(sets) : ''}
                          </td>
                        );
                      })}
                    </tr>
                  ))}
                </tbody>
              </table>
            </section>
          )}
        </>
      )}
    </>
  );
}

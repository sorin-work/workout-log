import type { Summary } from '../api/types';
import { formatVolume, type Unit } from '../lib/units';

function delta(now: number, before: number): string {
  if (before === 0) return '';
  const pct = Math.round(((now - before) / before) * 100);
  return pct === 0 ? '' : ` (${pct > 0 ? '+' : ''}${pct}% vs last week)`;
}

export function SummaryStats({ summary, unit }: { summary: Summary; unit: Unit }) {
  const { thisWeek, lastWeek } = summary;
  return (
    <div className="stats">
      <div className="card stat">
        <div className="muted small">Workouts this week</div>
        <strong>{thisWeek.workouts}</strong>
        <div className="muted small">{lastWeek.workouts} last week</div>
      </div>
      <div className="card stat">
        <div className="muted small">Sets this week</div>
        <strong>{thisWeek.sets}</strong>
        <div className="muted small">{delta(thisWeek.sets, lastWeek.sets) || ' '}</div>
      </div>
      <div className="card stat">
        <div className="muted small">Volume this week</div>
        <strong>{formatVolume(thisWeek.volumeKg, unit)}</strong>
      </div>
      <div className="card stat">
        <div className="muted small">Week streak</div>
        <strong>{summary.weekStreak}</strong>
        <div className="muted small">{summary.totalWorkouts} workouts in total</div>
      </div>
    </div>
  );
}

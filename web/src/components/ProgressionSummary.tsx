import type { NewRecord, ProgressionChange } from '../api/types';
import { formatWeight, type Unit } from '../lib/units';

interface Props {
  changes: ProgressionChange[];
  records: NewRecord[];
  unit: Unit;
  onClose(): void;
}

function describe(c: ProgressionChange, unit: Unit): string {
  switch (c.outcome) {
    case 'increased':
      return `Next time: ${formatWeight(c.weightKg, unit)} (up from ${formatWeight(c.previousWeightKg, unit)})`;
    case 'deloaded':
      return `Deload: next time ${formatWeight(c.weightKg, unit)}`;
    default:
      return c.failures > 0
        ? `Stays at ${formatWeight(c.weightKg, unit)} (${c.failures} missed in a row)`
        : `Stays at ${formatWeight(c.weightKg, unit)}`;
  }
}

export function ProgressionSummary({ changes, records, unit, onClose }: Props) {
  return (
    <div className="card highlight progression">
      <div className="page-header">
        <h2>Workout finished</h2>
        <button className="icon-btn" aria-label="Close" onClick={onClose}>
          ×
        </button>
      </div>
      {records.length > 0 && (
        <ul className="plain records-list">
          {records.map((r) => (
            <li key={`${r.exerciseId}-${r.kind}`}>
              <span className="badge pr">PR</span> <strong>{r.exerciseName}</strong>:{' '}
              {r.kind === 'heaviest' ? 'heaviest set' : 'estimated 1RM'}{' '}
              {formatWeight(r.value, unit)}{' '}
              <span className="muted">(was {formatWeight(r.previous, unit)})</span>
            </li>
          ))}
        </ul>
      )}
      {changes.length === 0 ? (
        <p className="muted">Nice work. Nothing in your template changed.</p>
      ) : (
        <ul className="plain">
          {changes.map((c) => (
            <li key={c.templateExerciseId} className={`outcome-${c.outcome}`}>
              <strong>{c.exerciseName}</strong>: {describe(c, unit)}
            </li>
          ))}
        </ul>
      )}
    </div>
  );
}

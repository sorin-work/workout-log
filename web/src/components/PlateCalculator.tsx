import { useState } from 'react';
import { loadBar } from '../lib/plates';
import { trimNumber, type Unit } from '../lib/units';

interface Props {
  initial: number;
  unit: Unit;
  onClose(): void;
}

export function PlateCalculator({ initial, unit, onClose }: Props) {
  const [target, setTarget] = useState(trimNumber(initial, 2));
  const load = loadBar(Number(target.replace(',', '.')) || 0, unit);
  return (
    <div className="card plate-calc" role="dialog" aria-label="Plate calculator">
      <div className="page-header">
        <h2>Plates</h2>
        <button className="icon-btn" aria-label="Close" onClick={onClose}>
          ×
        </button>
      </div>
      <label>
        Target ({unit})
        <input inputMode="decimal" value={target} onChange={(e) => setTarget(e.target.value)} />
      </label>
      {load.perSide.length === 0 ? (
        <p className="muted">Just the bar.</p>
      ) : (
        <p>
          Each side: <strong>{load.perSide.map((p) => trimNumber(p, 2)).join(' + ')}</strong>
        </p>
      )}
      {load.remainder > 0 && (
        <p className="muted small">
          Closest you can load is {trimNumber(load.total, 2)} {unit}.
        </p>
      )}
    </div>
  );
}

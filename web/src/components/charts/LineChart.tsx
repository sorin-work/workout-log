import { useState } from 'react';
import { linearScale, niceTicks } from './scale';

export interface LinePoint {
  x: number; // timestamp in ms
  y: number;
  label?: string;
}

interface Props {
  points: LinePoint[];
  height?: number;
  formatY?(v: number): string;
  formatX?(v: number): string;
  ariaLabel: string;
}

const W = 640;
const PAD = { top: 12, right: 16, bottom: 28, left: 52 };

export function LineChart({
  points,
  height = 240,
  formatY = String,
  formatX = (v) => new Date(v).toLocaleDateString('en-GB', { day: 'numeric', month: 'short' }),
  ariaLabel,
}: Props) {
  const [hover, setHover] = useState<number | null>(null);
  if (points.length === 0) return <p className="muted">No data yet.</p>;

  const xs = points.map((p) => p.x);
  const ys = points.map((p) => p.y);
  const ticks = niceTicks(Math.min(...ys), Math.max(...ys), 4);
  const yMin = Math.min(ticks[0], ...ys);
  const yMax = Math.max(ticks[ticks.length - 1], ...ys);
  const x = linearScale(Math.min(...xs), Math.max(...xs), PAD.left, W - PAD.right);
  const y = linearScale(yMin, yMax, height - PAD.bottom, PAD.top);
  const path = points.map((p, i) => `${i === 0 ? 'M' : 'L'}${x(p.x)},${y(p.y)}`).join(' ');
  const h = hover === null ? null : points[hover];

  return (
    <svg
      className="chart"
      viewBox={`0 0 ${W} ${height}`}
      role="img"
      aria-label={ariaLabel}
      onMouseLeave={() => setHover(null)}
    >
      {ticks.map((t) => (
        <g key={t}>
          <line x1={PAD.left} x2={W - PAD.right} y1={y(t)} y2={y(t)} className="grid-line" />
          <text x={PAD.left - 6} y={y(t)} className="axis-label" textAnchor="end" dy="0.32em">
            {formatY(t)}
          </text>
        </g>
      ))}
      <text x={PAD.left} y={height - 8} className="axis-label">
        {formatX(xs[0])}
      </text>
      {points.length > 1 && (
        <text x={W - PAD.right} y={height - 8} className="axis-label" textAnchor="end">
          {formatX(xs[xs.length - 1])}
        </text>
      )}
      <path d={path} className="line" />
      {points.map((p, i) => (
        <circle
          key={i}
          cx={x(p.x)}
          cy={y(p.y)}
          r={hover === i ? 5 : 3}
          className="dot"
          onMouseEnter={() => setHover(i)}
        />
      ))}
      {h && (
        <text x={x(h.x)} y={y(h.y) - 10} textAnchor="middle" className="tooltip">
          {h.label ?? formatY(h.y)}
        </text>
      )}
    </svg>
  );
}

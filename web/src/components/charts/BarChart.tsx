import { linearScale, niceTicks } from './scale';

export interface Bar {
  label: string;
  value: number;
  title?: string;
}

interface Props {
  bars: Bar[];
  height?: number;
  formatY?(v: number): string;
  ariaLabel: string;
}

const W = 640;
const PAD = { top: 12, right: 8, bottom: 28, left: 40 };

export function BarChart({ bars, height = 200, formatY = String, ariaLabel }: Props) {
  if (bars.length === 0) return <p className="muted">No data yet.</p>;
  const max = Math.max(0, ...bars.map((b) => b.value));
  const ticks = niceTicks(0, max || 1, 4);
  const top = Math.max(ticks[ticks.length - 1], max);
  const y = linearScale(0, top, height - PAD.bottom, PAD.top);
  const slot = (W - PAD.left - PAD.right) / bars.length;
  const barW = Math.max(2, slot * 0.7);
  const every = Math.ceil(bars.length / 8);

  return (
    <svg className="chart" viewBox={`0 0 ${W} ${height}`} role="img" aria-label={ariaLabel}>
      {ticks.map((t) => (
        <g key={t}>
          <line x1={PAD.left} x2={W - PAD.right} y1={y(t)} y2={y(t)} className="grid-line" />
          <text x={PAD.left - 6} y={y(t)} className="axis-label" textAnchor="end" dy="0.32em">
            {formatY(t)}
          </text>
        </g>
      ))}
      {bars.map((b, i) => {
        const x = PAD.left + i * slot + (slot - barW) / 2;
        return (
          <g key={b.label}>
            <rect
              x={x}
              y={y(b.value)}
              width={barW}
              height={Math.max(0, y(0) - y(b.value))}
              className="bar"
            >
              <title>{b.title ?? `${b.label}: ${formatY(b.value)}`}</title>
            </rect>
            {i % every === 0 && (
              <text x={x + barW / 2} y={height - 8} className="axis-label" textAnchor="middle">
                {b.label}
              </text>
            )}
          </g>
        );
      })}
    </svg>
  );
}

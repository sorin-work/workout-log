import { Link } from 'react-router';
import { useFetch } from '../lib/useFetch';

interface Day {
  date: string;
  workouts: { id: number; name: string }[];
}

interface Props {
  month: string; // YYYY-MM
  onMonthChange(month: string): void;
}

function shift(month: string, delta: number): string {
  const [y, m] = month.split('-').map(Number);
  const d = new Date(y, m - 1 + delta, 1);
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}`;
}

/** Month grid starting on Monday, with links to each day's workouts. */
export function MonthCalendar({ month, onMonthChange }: Props) {
  const { data } = useFetch<Day[]>(`/stats/calendar?month=${month}`);
  const [y, m] = month.split('-').map(Number);
  const first = new Date(y, m - 1, 1);
  const daysInMonth = new Date(y, m, 0).getDate();
  const lead = (first.getDay() + 6) % 7;
  const byDate = new Map(data?.map((d) => [d.date, d.workouts]) ?? []);
  const cells: (number | null)[] = [
    ...Array<null>(lead).fill(null),
    ...Array.from({ length: daysInMonth }, (_, i) => i + 1),
  ];
  const title = first.toLocaleDateString('en-GB', { month: 'long', year: 'numeric' });

  return (
    <section className="card calendar">
      <div className="page-header">
        <button
          className="icon-btn"
          aria-label="Previous month"
          onClick={() => onMonthChange(shift(month, -1))}
        >
          ‹
        </button>
        <h2>{title}</h2>
        <button
          className="icon-btn"
          aria-label="Next month"
          onClick={() => onMonthChange(shift(month, 1))}
        >
          ›
        </button>
      </div>
      <div className="calendar-grid">
        {['Mo', 'Tu', 'We', 'Th', 'Fr', 'Sa', 'Su'].map((d) => (
          <div key={d} className="calendar-head">
            {d}
          </div>
        ))}
        {cells.map((day, i) => {
          if (day === null) return <div key={`e${i}`} />;
          const date = `${month}-${String(day).padStart(2, '0')}`;
          const workouts = byDate.get(date);
          return (
            <div key={date} className={workouts ? 'calendar-day trained' : 'calendar-day'}>
              {workouts ? (
                <Link
                  to={`/workouts/${workouts[0].id}`}
                  title={workouts.map((w) => w.name).join(', ')}
                >
                  {day}
                </Link>
              ) : (
                day
              )}
            </div>
          );
        })}
      </div>
    </section>
  );
}

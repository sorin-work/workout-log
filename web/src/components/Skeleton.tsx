/** Grey placeholder rows shown while a list loads. */
export function Skeleton({ rows = 4 }: { rows?: number }) {
  return (
    <div className="card skeleton" aria-busy="true" aria-label="Loading">
      {Array.from({ length: rows }, (_, i) => (
        <div key={i} className="skeleton-row" style={{ width: `${90 - ((i * 17) % 35)}%` }} />
      ))}
    </div>
  );
}

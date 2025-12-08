import { Link } from 'react-router';

export default function NotFoundPage() {
  return (
    <div className="empty">
      <h1>Page not found</h1>
      <Link to="/">Back to the start</Link>
    </div>
  );
}

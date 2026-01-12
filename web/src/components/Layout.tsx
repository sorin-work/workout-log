import { NavLink, Outlet } from 'react-router';
import { useAuth } from '../auth/context';

const links = [
  { to: '/', label: 'Home', end: true },
  { to: '/workouts', label: 'Workouts' },
  { to: '/templates', label: 'Templates' },
  { to: '/exercises', label: 'Exercises' },
  { to: '/settings', label: 'Settings' },
];

export function Layout() {
  const { user, logout } = useAuth();
  return (
    <div className="app">
      <header className="topbar">
        <NavLink to="/" className="brand">
          Workout Log
        </NavLink>
        <nav>
          {links.map((l) => (
            <NavLink key={l.to} to={l.to} end={l.end}>
              {l.label}
            </NavLink>
          ))}
        </nav>
        <div className="topbar-user">
          <span className="muted">{user?.displayName || user?.email}</span>
          <button className="btn small" onClick={logout}>
            Log out
          </button>
        </div>
      </header>
      <main className="container">
        <Outlet />
      </main>
    </div>
  );
}

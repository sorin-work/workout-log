import { useAuth } from '../auth/context';

export default function HomePage() {
  const { user } = useAuth();
  return (
    <>
      <h1>Welcome back{user?.displayName ? `, ${user.displayName}` : ''}</h1>
      <p className="muted">Start a workout or look through the exercise library.</p>
    </>
  );
}

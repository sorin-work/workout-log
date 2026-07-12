import { Link } from 'react-router';

/** Shown on the home page until the first workout is logged. */
export function GettingStarted({ hasTemplates }: { hasTemplates: boolean }) {
  return (
    <section className="card getting-started">
      <h2>Getting started</h2>
      <ol>
        <li className={hasTemplates ? 'done' : ''}>
          <Link to="/templates/new">Create a template</Link> with the lifts you do, sets, reps and
          the weight to add each session.
        </li>
        <li>
          <Link to="/programs">Put templates in a program</Link> if you alternate days, e.g. A/B.
        </li>
        <li>Start a workout below and tick off sets as you go.</li>
      </ol>
    </section>
  );
}

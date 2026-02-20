import { act, render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { formatClock, RestTimer } from './RestTimer';

describe('formatClock', () => {
  it('formats minutes and seconds', () => {
    expect(formatClock(125)).toBe('2:05');
    expect(formatClock(59.2)).toBe('1:00');
    expect(formatClock(-3)).toBe('0:00');
  });
});

describe('RestTimer', () => {
  beforeEach(() => {
    vi.useFakeTimers({ shouldAdvanceTime: true });
  });
  afterEach(() => {
    vi.useRealTimers();
  });

  it('counts down and calls onDone', () => {
    const onDone = vi.fn();
    render(<RestTimer startedAt={Date.now()} seconds={90} onDone={onDone} onDismiss={() => {}} />);
    expect(screen.getByRole('timer')).toHaveTextContent('1:30');
    act(() => {
      vi.advanceTimersByTime(30_000);
    });
    expect(screen.getByRole('timer')).toHaveTextContent('1:00');
    act(() => {
      vi.advanceTimersByTime(61_000);
    });
    expect(screen.getByRole('timer')).toHaveTextContent('Go!');
    expect(onDone).toHaveBeenCalledTimes(1);
  });

  it('adds thirty seconds', async () => {
    const user = userEvent.setup({ advanceTimers: vi.advanceTimersByTime });
    render(<RestTimer startedAt={Date.now()} seconds={60} onDismiss={() => {}} />);
    await user.click(screen.getByRole('button', { name: '+30 s' }));
    expect(screen.getByRole('timer')).toHaveTextContent('1:30');
  });
});

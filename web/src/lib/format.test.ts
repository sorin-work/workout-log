import { describe, expect, it } from 'vitest';
import { formatDuration } from './format';

describe('formatDuration', () => {
  it('is empty for unfinished workouts', () => {
    expect(formatDuration('2026-01-03T18:00:00Z', null)).toBe('');
  });

  it('shows minutes below an hour', () => {
    expect(formatDuration('2026-01-03T18:00:00Z', '2026-01-03T18:45:20Z')).toBe('45 min');
  });

  it('pads minutes after hours', () => {
    expect(formatDuration('2026-01-03T18:00:00Z', '2026-01-03T19:05:00Z')).toBe('1 h 05 min');
  });

  it('never goes negative', () => {
    expect(formatDuration('2026-01-03T18:00:00Z', '2026-01-03T17:00:00Z')).toBe('0 min');
  });
});

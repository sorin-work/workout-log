import { describe, expect, it } from 'vitest';
import { localMonth } from './format';

describe('localMonth', () => {
  it('uses the local calendar month', () => {
    expect(localMonth(new Date(2026, 6, 1, 0, 30))).toBe('2026-07');
    expect(localMonth(new Date(2026, 11, 31, 23, 59))).toBe('2026-12');
  });
});

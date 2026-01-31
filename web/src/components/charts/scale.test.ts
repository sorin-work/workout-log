import { describe, expect, it } from 'vitest';
import { linearScale, niceTicks } from './scale';

describe('linearScale', () => {
  it('maps the domain onto the range', () => {
    const s = linearScale(0, 100, 0, 200);
    expect(s(0)).toBe(0);
    expect(s(50)).toBe(100);
    expect(s(100)).toBe(200);
  });

  it('handles an inverted range for y axes', () => {
    const s = linearScale(0, 10, 100, 0);
    expect(s(10)).toBe(0);
  });

  it('does not divide by zero for a single value', () => {
    expect(linearScale(5, 5, 0, 10)(5)).toBe(0);
  });
});

describe('niceTicks', () => {
  it('produces round steps', () => {
    expect(niceTicks(97, 143, 4)).toEqual([90, 100, 110, 120, 130, 140]);
  });

  it('spreads a flat series', () => {
    const t = niceTicks(100, 100);
    expect(t[0]).toBeLessThan(100);
    expect(t[t.length - 1]).toBeGreaterThan(100);
  });

  it('returns nothing for empty input', () => {
    expect(niceTicks(Infinity, -Infinity)).toEqual([]);
  });
});

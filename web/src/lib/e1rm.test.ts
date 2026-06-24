import { describe, expect, it } from 'vitest';
import { estimateOneRepMax } from './e1rm';

describe('estimateOneRepMax', () => {
  it('matches the server formulas', () => {
    expect(estimateOneRepMax(100, 5, 'epley')).toBe(116.7);
    expect(estimateOneRepMax(100, 5, 'brzycki')).toBe(112.5);
  });

  it('treats a single as the max', () => {
    expect(estimateOneRepMax(180, 1, 'epley')).toBe(180);
  });

  it('ignores sets it cannot estimate', () => {
    expect(estimateOneRepMax(0, 5, 'epley')).toBe(0);
    expect(estimateOneRepMax(60, 15, 'epley')).toBe(0);
  });
});

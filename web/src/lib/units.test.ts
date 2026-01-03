import { describe, expect, it } from 'vitest';
import { formatVolume, formatWeight, fromKg, trimNumber } from './units';

describe('units', () => {
  it('converts kilograms to pounds', () => {
    expect(fromKg(100, 'lb')).toBeCloseTo(220.462, 3);
    expect(fromKg(100, 'kg')).toBe(100);
  });

  it('trims trailing zeros', () => {
    expect(trimNumber(100)).toBe('100');
    expect(trimNumber(102.5)).toBe('102.5');
    expect(trimNumber(61.249)).toBe('61.2');
  });

  it('formats weights in the chosen unit', () => {
    expect(formatWeight(102.5, 'kg')).toBe('102.5 kg');
    expect(formatWeight(102.0583, 'lb')).toBe('225 lb');
  });

  it('formats volume with separators', () => {
    expect(formatVolume(12450, 'kg')).toBe('12,450 kg');
    expect(formatVolume(1000, 'lb')).toBe('2,205 lb');
  });
});

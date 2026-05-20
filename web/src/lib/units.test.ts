import { describe, expect, it } from 'vitest';
import { formatVolume, formatWeight, fromKg, toKg, trimNumber } from './units';

describe('units', () => {
  it('converts kilograms to pounds', () => {
    expect(fromKg(100, 'lb')).toBeCloseTo(220.462, 3);
    expect(fromKg(100, 'kg')).toBe(100);
  });

  it('converts typed weights back to kilograms', () => {
    expect(toKg(80, 'kg')).toBe(80);
    expect(toKg(220.462, 'lb')).toBeCloseTo(100, 1);
    expect(toKg(0, 'lb')).toBe(0);
  });

  it('keeps pound weights intact through a save', () => {
    for (const lb of [135, 185, 225, 315, 47.5, 2.5]) {
      const stored = Math.round(toKg(lb, 'lb') * 1000) / 1000; // numeric(7, 3)
      expect(trimNumber(fromKg(stored, 'lb'), 2)).toBe(String(lb));
    }
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

import { describe, expect, it } from 'vitest';
import { loadBar } from './plates';

describe('loadBar', () => {
  it('loads a common kg weight', () => {
    expect(loadBar(100, 'kg')).toEqual({ perSide: [25, 15], remainder: 0, total: 100 });
  });

  it('uses small plates', () => {
    expect(loadBar(62.5, 'kg').perSide).toEqual([20, 1.25]);
  });

  it('loads pounds with 45 lb plates', () => {
    expect(loadBar(225, 'lb').perSide).toEqual([45, 45]);
    expect(loadBar(135, 'lb').perSide).toEqual([45]);
  });

  it('reports what cannot be loaded', () => {
    const load = loadBar(101, 'kg');
    expect(load.total).toBe(100);
    expect(load.remainder).toBe(1);
  });

  it('handles just the bar', () => {
    expect(loadBar(20, 'kg')).toEqual({ perSide: [], remainder: 0, total: 20 });
  });
});

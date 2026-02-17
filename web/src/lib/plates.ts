import type { Unit } from './units';

export const BAR_WEIGHT: Record<Unit, number> = { kg: 20, lb: 45 };

export const PLATES: Record<Unit, number[]> = {
  kg: [25, 20, 15, 10, 5, 2.5, 1.25],
  lb: [45, 35, 25, 10, 5, 2.5],
};

export interface PlateLoad {
  /** Plates for one side of the bar, heaviest first. */
  perSide: number[];
  /** Weight that could not be loaded with the available plates. */
  remainder: number;
  total: number;
}

/** Works out the plates per side for a target weight, in the given unit. */
export function loadBar(
  target: number,
  unit: Unit,
  bar = BAR_WEIGHT[unit],
  plates = PLATES[unit],
): PlateLoad {
  if (target <= bar) return { perSide: [], remainder: Math.max(0, target - bar), total: bar };
  let side = (target - bar) / 2;
  const perSide: number[] = [];
  for (const p of [...plates].sort((a, b) => b - a)) {
    while (side + 1e-9 >= p) {
      perSide.push(p);
      side -= p;
    }
  }
  const loaded = perSide.reduce((a, b) => a + b, 0);
  const total = bar + loaded * 2;
  return { perSide, remainder: Math.round((target - total) * 100) / 100, total };
}

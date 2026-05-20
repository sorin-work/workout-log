export type Unit = 'kg' | 'lb';

export const KG_PER_LB = 0.45359237;

/** Converts kilograms to the display unit. */
export function fromKg(kg: number, unit: Unit): number {
  return unit === 'lb' ? kg / KG_PER_LB : kg;
}

/**
 * Converts a weight typed in the display unit to kilograms for the API. The
 * database keeps three decimals, which is enough for pounds to survive the
 * round trip; rounding harder here made 135 lb come back as 134.92 lb.
 */
export function toKg(value: number, unit: Unit): number {
  if (unit === 'kg') return value;
  return Math.round(value * KG_PER_LB * 1000) / 1000;
}

/** Trims a number to at most `decimals` places without trailing zeros. */
export function trimNumber(n: number, decimals = 1): string {
  const f = 10 ** decimals;
  return String(Math.round(n * f) / f);
}

/** "102.5 kg", "225 lb". */
export function formatWeight(kg: number, unit: Unit): string {
  return `${trimNumber(fromKg(kg, unit))} ${unit}`;
}

/** Tonnage with thousands separators: "12,450 kg". */
export function formatVolume(kg: number, unit: Unit): string {
  return `${Math.round(fromKg(kg, unit)).toLocaleString('en-US')} ${unit}`;
}

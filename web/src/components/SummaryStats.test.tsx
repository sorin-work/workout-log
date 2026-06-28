import { render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import { SummaryStats } from './SummaryStats';

const summary = {
  thisWeek: { workouts: 2, sets: 30, volumeKg: 9000 },
  lastWeek: { workouts: 3, sets: 40, volumeKg: 12000 },
  weekStreak: 6,
  totalWorkouts: 120,
};

describe('SummaryStats', () => {
  it('compares this week with last week', () => {
    render(<SummaryStats summary={summary} unit="kg" />);
    expect(screen.getByText('(-25% vs last week)')).toBeInTheDocument();
    expect(screen.getByText('9,000 kg')).toBeInTheDocument();
    expect(screen.getByText('120 workouts in total')).toBeInTheDocument();
  });

  it('converts volume to pounds', () => {
    render(<SummaryStats summary={summary} unit="lb" />);
    expect(screen.getByText('19,842 lb')).toBeInTheDocument();
  });
});

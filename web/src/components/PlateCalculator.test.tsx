import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, expect, it } from 'vitest';
import { PlateCalculator } from './PlateCalculator';

describe('PlateCalculator', () => {
  it('shows the plates per side and updates while typing', async () => {
    const user = userEvent.setup();
    render(<PlateCalculator initial={100} unit="kg" onClose={() => {}} />);
    expect(screen.getByText('25 + 15')).toBeInTheDocument();
    const input = screen.getByLabelText('Target (kg)');
    await user.clear(input);
    await user.type(input, '61');
    expect(screen.getByText('20')).toBeInTheDocument();
    expect(screen.getByText(/Closest you can load is 60 kg/)).toBeInTheDocument();
  });

  it('works in pounds', () => {
    render(<PlateCalculator initial={225} unit="lb" onClose={() => {}} />);
    expect(screen.getByText('45 + 45')).toBeInTheDocument();
  });
});

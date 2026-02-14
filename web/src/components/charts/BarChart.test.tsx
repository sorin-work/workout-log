import { render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import { BarChart } from './BarChart';

describe('BarChart', () => {
  it('draws one bar per value with a tooltip', () => {
    const { container } = render(
      <BarChart
        ariaLabel="sets per week"
        bars={[
          { label: 'W1', value: 10 },
          { label: 'W2', value: 0 },
          { label: 'W3', value: 25 },
        ]}
      />,
    );
    expect(screen.getByRole('img', { name: 'sets per week' })).toBeInTheDocument();
    const rects = container.querySelectorAll('rect.bar');
    expect(rects).toHaveLength(3);
    expect(rects[1].getAttribute('height')).toBe('0');
    expect(rects[2].querySelector('title')?.textContent).toBe('W3: 25');
  });

  it('shows an empty state', () => {
    render(<BarChart ariaLabel="none" bars={[]} />);
    expect(screen.getByText('No data yet.')).toBeInTheDocument();
  });
});

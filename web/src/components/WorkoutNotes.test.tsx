import { render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import { WorkoutNotes } from './WorkoutNotes';

describe('WorkoutNotes', () => {
  it('shows notes as text, not html', () => {
    const notes = 'PR day!\n<img src=x onerror="alert(1)"><b>bold</b>';
    const { container } = render(<WorkoutNotes notes={notes} />);
    expect(container.querySelector('img')).toBeNull();
    expect(container.querySelector('b')).toBeNull();
    expect(screen.getByText(/onerror/)).toBeInTheDocument();
  });

  it('keeps line breaks', () => {
    const { container } = render(<WorkoutNotes notes={'line one\nline two'} />);
    expect(container.textContent).toBe('line one\nline two');
  });

  it('renders nothing for empty notes', () => {
    const { container } = render(<WorkoutNotes notes="  " />);
    expect(container).toBeEmptyDOMElement();
  });
});

/** Workout notes as plain text; line breaks are kept by CSS. */
export function WorkoutNotes({ notes }: { notes: string }) {
  if (!notes.trim()) return null;
  return <div className="card notes">{notes}</div>;
}

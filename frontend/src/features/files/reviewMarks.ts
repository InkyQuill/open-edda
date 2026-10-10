import { Decoration, EditorView } from '@codemirror/view';
import type { ReviewRecord } from './pocketApi';

// Wire anchors count UTF-8 bytes; CodeMirror counts UTF-16 and normalizes newlines.
export function reviewPositions(source: string, records: ReviewRecord[]) {
  const offsets = new Set(records.flatMap(r => r.resolution.range ? [r.resolution.range.from, r.resolution.range.to] : []));
  const mapped = new Map<number, number>();
  let bytes = 0, position = 0, previousCR = false;
  for (const char of source) {
    if (offsets.has(bytes)) mapped.set(bytes, position);
    const code = char.codePointAt(0)!;
    bytes += code < 128 ? 1 : code < 2048 ? 2 : code < 65536 ? 3 : 4;
    if (!(previousCR && char === '\n')) position += char.length;
    previousCR = char === '\r';
  }
  if (offsets.has(bytes)) mapped.set(bytes, position);
  return records.flatMap(record => {
    const range = record.resolution.range;
    if (!range || record.resolution.kind !== 'resolved') return [];
    const from = mapped.get(range.from), to = mapped.get(range.to);
    return from === undefined || to === undefined || from >= to ? [] : [{ record, from, to }];
  });
}
export function reviewMarks(source: string, records: ReviewRecord[], selected: string, reveal: (record: ReviewRecord) => void) {
  const ranges = reviewPositions(source, records);
  return [EditorView.decorations.of(Decoration.set(ranges.map(({ record, from, to }) => Decoration.mark({ class: `pocket-mark ${selected === record.id ? 'pocket-mark-active' : ''}`, attributes: { 'data-review-id': record.id } }).range(from, to)), true)),
    EditorView.domEventHandlers({ click(event) { const id = (event.target as HTMLElement).closest('[data-review-id]')?.getAttribute('data-review-id'); const record = records.find(r => r.id === id); if (record) reveal(record); return false; } })];
}

export function reviewSelection(source: string, from: number, to: number) {
  let position = 0, start = -1, end = -1;
  for (let i = 0; i <= source.length; i++) {
    if (position === from) start = i;
    if (position === to) { end = i; break; }
    if (source[i] === '\r' && source[i + 1] === '\n') i++;
    position++;
  }
  if (start < 0 || end <= start) return null;
  const selected = source.slice(start, end), encoder = new TextEncoder();
  return { from: encoder.encode(source.slice(0, start)).length, to: encoder.encode(source.slice(0, end)).length, selected };
}

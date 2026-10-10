import { expect, test } from 'vitest';
import { reviewPositions } from './reviewMarks';
import type { ReviewRecord } from './pocketApi';

test('byte anchors navigate Cyrillic, emoji and CRLF in both editor surfaces', () => {
  const source = '🐦\r\nПривет 日本\r\n';
  const from = new TextEncoder().encode('🐦\r\nПривет ').length;
  const records: ReviewRecord[] = [{ id: 'one', resolution: { kind: 'resolved', range: { from, to: from + 6 } } }];
  expect(reviewPositions(source, records)[0]).toMatchObject({ from: '🐦\nПривет '.length, to: '🐦\nПривет 日本'.length });
  expect(reviewPositions(source, [{ id: 'bad', resolution: { kind: 'resolved', range: { from: 1, to: 3 } } }])).toEqual([]);
});

test('editor selections map back to exact CRLF/emoji source bytes for new reviews', async () => {
  const { reviewSelection } = await import('./reviewMarks');
  const source = '🐦\r\nТихо\r\n';
  expect(reviewSelection(source, 3, 7)).toEqual({ from: 6, to: 14, selected: 'Тихо' });
  expect(reviewSelection(source, 2, 3)).toEqual({ from: 4, to: 6, selected: '\r\n' });
  expect(reviewSelection(source, 4, 4)).toBeNull();
});

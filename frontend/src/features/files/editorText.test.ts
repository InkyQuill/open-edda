import { expect, test } from 'vitest';
import { editorText, sourceText } from './editorText';

test('opening and changing modes preserves CRLF, CR and mixed source bytes', () => {
  for (const original of ['🐦\r\nПривет\r\n', 'a\rb\nccc\r\n', '', '\r\n']) expect(sourceText(original, editorText(original))).toBe(original);
});
test('author edits preserve unchanged mixed newlines and use the existing convention for inserted lines', () => {
  expect(sourceText('🐦\r\nПривет\r\n', '🐦\nДобрый день\n')).toBe('🐦\r\nДобрый день\r\n');
  expect(sourceText('a\rb\nccc\r\n', 'a\nb\nновое\nccc\n')).toBe('a\rb\nновое\rccc\r\n');
  expect(sourceText('a\r\nb\r\n', 'a\n')).toBe('a\r\n');
  expect(sourceText('a\r\nb', 'a\nb\nc')).toBe('a\r\nb\r\nc');
  expect(sourceText('abc', '')).toBe('');
});

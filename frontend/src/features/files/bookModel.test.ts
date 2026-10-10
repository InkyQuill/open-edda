import { describe, expect, it } from 'vitest';
import { decodeBook, moveChapter } from './bookModel';
const id = '0b4f1cad-c846-4551-a497-a745087f5de2';
const chapter = { id, path: 'one.md', extra: { custom: true } };
const book = { schema_version: 2, book_id: id, chapters: [chapter, { id: id.replace(/^0/, '1'), path: 'missing.md' }], future: { keep: true } };
describe('author controlled book spine', () => {
 it('keeps missing chapters and extension properties while moving chapters', () => {
  const result = moveChapter(decodeBook(JSON.stringify(book)), 1, -1);
  expect(result.chapters.map(c => c.path)).toEqual(['missing.md', 'one.md']);
  expect(result.chapters[1].extra).toEqual({ custom: true });
  expect(result.future).toEqual({ keep: true });
  expect(result.schema_version).toBe(2);
 });
 it('accepts an empty book without workflow gates', () => expect(decodeBook(JSON.stringify({ schema_version: 2, book_id: id })).chapters).toEqual([]));
 it.each([
  { ...book, chapters: [chapter, chapter] },
  { ...book, chapters: [{ ...chapter, path: '../one.md' }] },
  { ...book, ignored_files: ['one.md'] },
  { ...book, schema_version: 3 },
  { ...book, chapters: null },
 ])('refuses unsafe manifests instead of rewriting them', value => expect(() => decodeBook(JSON.stringify(value))).toThrow());
});

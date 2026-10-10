export type BookChapter = { id: string; path: string; title?: string; [key: string]: unknown };
export type BookManifest = { schema_version: 1 | 2; book_id: string; title?: string; chapters: BookChapter[]; ignored_files?: string[]; [key: string]: unknown };
const uuid = /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i;
const child = (s: unknown): s is string => typeof s === 'string' && !!s && s !== '.' && s !== '..' && !/[\\/\x00-\x1f]/.test(s);
const object = (v: unknown): v is Record<string, unknown> => !!v && typeof v === 'object' && !Array.isArray(v);
export function decodeBook(text: string): BookManifest {
  if (new TextEncoder().encode(text).length > 1024 * 1024) throw new Error('Оглавление превышает 1 МиБ.');
  const v: unknown = JSON.parse(text);
  if (!object(v) || ![1, 2].includes(Number(v.schema_version)) || typeof v.schema_version !== 'number' || typeof v.book_id !== 'string' || !uuid.test(v.book_id)) throw new Error('Неизвестный или повреждённый формат книги.');
  if ((v.schema_version === 1 || 'title' in v) && typeof v.title !== 'string') throw new Error('Некорректное название книги.');
  const chapters = v.chapters ?? (v.schema_version === 2 && !('chapters' in v) ? [] : null);
  if (!Array.isArray(chapters)) throw new Error('Некорректный список глав.');
  const ids = new Set<string>(), paths = new Set<string>(), sidecars = new Set<string>();
  for (const c of chapters) {
    if (!object(c) || typeof c.id !== 'string' || !uuid.test(c.id) || !child(c.path) || !c.path.endsWith('.md') || ((v.schema_version === 1 || 'title' in c) && typeof c.title !== 'string')) throw new Error('Некорректная глава в оглавлении.');
    const reviews = [c.path.slice(0, -3) + '.review.json', c.path + '.review.json'];
    if (ids.has(c.id) || paths.has(c.path) || reviews.some(p => sidecars.has(p))) throw new Error('Повтор главы или пересечение имён рецензий.');
    ids.add(c.id); paths.add(c.path); reviews.forEach(p => sidecars.add(p));
  }
  if ('ignored_files' in v && (!Array.isArray(v.ignored_files) || v.ignored_files.some(p => !child(p) || paths.has(p)) || new Set(v.ignored_files).size !== v.ignored_files.length)) throw new Error('Некорректный список исключённых файлов.');
  return { ...v, chapters } as BookManifest;
}
export function moveChapter(book: BookManifest, index: number, delta: -1 | 1): BookManifest {
  const next = index + delta;
  if (index < 0 || index >= book.chapters.length || next < 0 || next >= book.chapters.length) return book;
  const chapters = [...book.chapters];
  [chapters[index], chapters[next]] = [chapters[next], chapters[index]];
  return { ...book, chapters };
}
export const bookRoot = (path: string) => path.slice(0, Math.max(0, path.lastIndexOf('/') + 1));

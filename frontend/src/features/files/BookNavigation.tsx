import { useEffect, useRef, useState } from 'react';
import { changeBook } from './pocketApi';
import { ApiError } from '../../api';
import { Modal } from '../../shared/ui/edda';
import { loadFile, type ProjectVersion, type TreeEntry } from './fileApi';
import { bookRoot, decodeBook, moveChapter, type BookManifest } from './bookModel';

export function BookNavigation({ version, busy, dirtyIds, run, onSaved, onOpen, error }: {
  version: ProjectVersion; busy: boolean; dirtyIds: Set<string>; run: (work: () => Promise<void>) => Promise<void>;
  onSaved: () => Promise<void>; onOpen: (entry: TreeEntry) => void; error: string;
}) {
  const manifests = version.entries.filter(e => e.path === '.pocket-editor.json' || e.path.endsWith('/.pocket-editor.json'));
  const [selected, setSelected] = useState('');
  const [retry, setRetry] = useState(0);
  const entry = manifests.find(e => e.id === selected) ?? manifests[0];
  const key = `${version.projectId}:${entry?.id}:${entry?.sha256}:${entry?.kind}`;
  const [loaded, setLoaded] = useState<{ key: string; book?: BookManifest; error?: string }>();
  const [draft, setDraft] = useState<{ base: ProjectVersion; entry: TreeEntry; book: BookManifest; createPaths: string[] } | null>(null);
  const [newPath, setNewPath] = useState('');
  const pending = useRef<{ body: string; operation: string } | null>(null);
  useEffect(() => {
    let cancelled = false;
    if (!entry) return;
    void (async () => {
      if (entry.kind !== 'file' || entry.bytes > 1024 * 1024) throw new Error('Оглавление должно быть JSON-файлом до 1 МиБ.');
      const book = decodeBook(await (await loadFile(version, entry)).text());
      if (!cancelled) setLoaded({ key, book });
    })().catch(cause => { if (!cancelled) setLoaded({ key, error: cause instanceof Error ? cause.message : 'Не удалось прочитать книгу.' }); });
    return () => { cancelled = true; };
  }, [key, retry]);
  const current = loaded?.key === key ? loaded : undefined;
  if (!entry) return null;
  const root = bookRoot(entry.path);
  const draftRoot = draft ? bookRoot(draft.entry.path) : '';
  const candidates = draft?.base.entries.filter(e => e.kind === 'file' && e.path.endsWith('.md') && bookRoot(e.path) === draftRoot && !e.path.slice(draftRoot.length).startsWith('.') && !draft.book.chapters.some(c => draftRoot + c.path === e.path) && !draft.book.ignored_files?.includes(e.path.slice(draftRoot.length))) ?? [];
  async function addExisting(name: string) {
    if (!draft || !name) return;
    const reviews = draft.base.entries.filter(e => e.path === draftRoot + name.slice(0, -3) + '.review.json' || e.path === draftRoot + name + '.review.json');
    if (reviews.length > 1) throw new Error('У главы два файла рецензий. Сначала устраните конфликт имён.');
    let id = crypto.randomUUID() as string;
    if (reviews[0]) {
      if (reviews[0].kind !== 'file' || reviews[0].bytes > 1024 * 1024) throw new Error('Не удалось проверить файл рецензии.');
      const review: unknown = JSON.parse(await (await loadFile(draft.base, reviews[0])).text());
      if (!review || typeof review !== 'object' || !('chapter_id' in review) || typeof review.chapter_id !== 'string' || !('source_path' in review) || review.source_path !== name) throw new Error('Рецензия не принадлежит выбранному файлу.');
      id = review.chapter_id;
    }
    const book = { ...draft.book, chapters: [...draft.book.chapters, { id, path: name, ...(draft.book.schema_version === 1 ? { title: name.slice(0, -3) } : {}) }] };
    decodeBook(JSON.stringify(book)); setDraft({ ...draft, book });
  }
  function createChapter() {
    if (!draft) return;
    const name = newPath.trim();
    const book = { ...draft.book, chapters: [...draft.book.chapters, { id: crypto.randomUUID(), path: name, title: name.slice(0, -3) }] };
    decodeBook(JSON.stringify(book));
    if (name.startsWith('.') || draft.base.entries.some(e => e.path === draftRoot + name || e.path === draftRoot + name.slice(0, -3) + '.review.json' || e.path === draftRoot + name + '.review.json')) throw new Error('Имя файла уже занято или скрыто.');
    setDraft({ ...draft, book, createPaths: [...new Set([...draft.createPaths, name])] }); setNewPath('');
  }
  async function save() {
    if (!draft || dirtyIds.has(draft.entry.id)) return;
    const body = JSON.stringify(draft.book, null, 2) + '\n'; decodeBook(body);
    if (!draft.book.title?.trim() || /[\r\n]/.test(draft.book.title)) throw new Error('Введите название книги в одну строку.');
    if (pending.current?.body !== body) pending.current = { body, operation: crypto.randomUUID() };
    try { await changeBook(draft.base.projectId, { expectedVersion: draft.base.id, operationId: pending.current.operation, entryId: draft.entry.id, title: draft.book.title, createPaths: draft.createPaths.filter(p => draft.book.chapters.some(c => c.path === p)), chapters: draft.book.chapters.map(({ id, path, title }) => ({ id, path, ...(title === undefined ? {} : { title }) })) }); }
    catch (cause) { if (cause instanceof ApiError && cause.status === 409) pending.current = null; throw cause; }
    await onSaved(); pending.current = null; setDraft(null);
  }
  return <section className="workspace-sections" aria-label="Оглавление книги">
    {manifests.length > 1 && <label className="field-label">Книга<select disabled={busy} value={entry.id} onChange={e => setSelected(e.target.value)}>{manifests.map(m => <option key={m.id} value={m.id}>{bookRoot(m.path) || 'Корень проекта'}</option>)}</select></label>}
    {current?.error && <><p className="form-error">{current.error}</p><button className="subtle-action" disabled={busy} onClick={() => setRetry(n => n + 1)}>Повторить загрузку книги</button></>}
    {current?.book && <><strong>{current.book.title || root || 'Книга'}</strong><nav aria-label="Главы книги">{current.book.chapters.map((c, i) => {
      const file = version.entries.find(e => e.path === root + c.path && e.kind === 'file');
      return <button key={c.id} disabled={!file || busy} onClick={() => file && onOpen(file)}>{i + 1}. {c.title || c.path}{!file && ' · файл не найден'}</button>;
    })}</nav><button className="subtle-action" disabled={busy || dirtyIds.has(entry.id)} onClick={() => { pending.current = null; setNewPath(''); setDraft({ base: version, entry, book: structuredClone(current.book!), createPaths: [] }); }}>Управление книгой</button></>}
    <Modal open={!!draft} busy={busy} onClose={() => setDraft(null)} title="Управление книгой" description="Порядок и названия в оглавлении независимы от расположения файлов и готовности текста. Удаление из оглавления сохраняет текст и рецензии." wide>
      {error && <p role="alert" className="form-error">{error}</p>}
      {draft && <form onSubmit={e => { e.preventDefault(); void run(save); }}>
        {version.id !== draft.base.id && <p className="fine-print">Версия проекта обновилась. Повтор сохранения проверит прежнюю операцию; для новых изменений закройте окно и откройте книгу заново.</p>}
        <label className="field-label">Название книги<input value={draft.book.title ?? ''} required disabled={busy} onChange={e => setDraft({ ...draft, book: { ...draft.book, title: e.target.value } })} /></label>
        <ol className="book-chapters">{draft.book.chapters.map((c, i) => <li key={c.id}>
          <label className="field-label">Название главы {i + 1}<input aria-label={`Название главы ${i + 1}`} disabled={busy} value={c.title ?? ''} placeholder={c.path} onChange={e => setDraft({ ...draft, book: { ...draft.book, chapters: draft.book.chapters.map(ch => ch.id === c.id ? { ...ch, title: e.target.value } : ch) } })} /></label>
          <span>{c.path}{!draft.base.entries.some(e => e.path === draftRoot + c.path && e.kind === 'file') && (draft.createPaths.includes(c.path) ? ' · новая глава' : ' · файл не найден')}</span>
          <div><button type="button" className="subtle-action" aria-label={`Поднять главу ${i + 1}`} disabled={busy || i === 0} onClick={() => setDraft({ ...draft, book: moveChapter(draft.book, i, -1) })}>↑</button><button type="button" className="subtle-action" aria-label={`Опустить главу ${i + 1}`} disabled={busy || i === draft.book.chapters.length - 1} onClick={() => setDraft({ ...draft, book: moveChapter(draft.book, i, 1) })}>↓</button><button type="button" className="subtle-action" disabled={busy} onClick={() => setDraft({ ...draft, book: { ...draft.book, chapters: draft.book.chapters.filter(ch => ch.id !== c.id) } })}>Убрать из оглавления</button></div>
        </li>)}</ol>
        <label className="field-label">Добавить существующий текст<select aria-label="Добавить существующий текст" value="" disabled={busy} onChange={e => { const name = e.target.value; void run(() => addExisting(name)); }}><option value="">Выберите файл</option>{candidates.map(e => <option key={e.id} value={e.path.slice(draftRoot.length)}>{e.path.slice(draftRoot.length)}</option>)}</select></label>
        <p className="fine-print">Здесь доступны Markdown-файлы рядом с оглавлением. При добавлении проверяется принадлежность рецензий.</p>
        <label className="field-label">Файл новой главы<input aria-label="Файл новой главы" placeholder="chapter.md" value={newPath} disabled={busy} onChange={e => setNewPath(e.target.value)} /></label><button type="button" className="subtle-action" disabled={busy || !newPath.trim()} onClick={() => void run(async () => createChapter())}>Создать главу</button>
        <div className="dialog-actions"><button className="quiet-button" disabled={busy || dirtyIds.has(draft.entry.id)}>Сохранить книгу</button></div>
      </form>}
    </Modal>
  </section>;
}

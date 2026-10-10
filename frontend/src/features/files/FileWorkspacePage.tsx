import { frontmatter } from './chapterMetadata';
import { ChapterMetadata } from './ChapterMetadata';
import { BookNavigation } from './BookNavigation';
import { WorkspaceSections } from './WorkspaceSections';
import { ReviewComposer, type ReviewComposition } from './ReviewComposer';
import { editorText, sourceText } from './editorText';
import type { GalleyHandle } from '@inkyquill/galley-editor';
import { usePocketReview } from './usePocketReview';
import { ReviewPanel } from './ReviewPanel';
import { reviewMarks, reviewPositions, reviewSelection } from './reviewMarks';
import type { ReviewRecord } from './pocketApi';
import { useEffect, useLayoutEffect, useMemo, useRef, useState } from 'react';
import { Link, useParams } from 'react-router-dom';
import { ArrowDownToLine, Check, ChevronRight, FileText, FolderPlus, History, Languages, Maximize2, Menu, Pencil, Plus, RefreshCw, Save, Settings2, Upload, X } from 'lucide-react';
import { IconButton, Modal, downloadFile } from '../../shared/ui/edda';
import { AppearanceButton, useAppearance } from '../appearance/Appearance';
import { FileTree } from './FileTree';
import { MarkdownEditor } from './MarkdownEditor';
import { FileHistory } from './FileHistory';
import { FilePreview } from './FilePreview';
import { basename, dirname, dirtyDraft, droppedItems, excludedPath, moveEntries, validatePath, type ImportItem } from './fileModel';
import { stageArchive, loadFile, publishTree, restoreVersion, uploadBlob, uploadText, type ProjectVersion, type TreeEntry } from './fileApi';
import { useFileWorkspace } from './useFileWorkspace';

export function FileWorkspacePage() { const { projectId = '' } = useParams(); return <FileWorkspace key={projectId} projectId={projectId} />; }
function FileWorkspace({ projectId }: { projectId: string }) {
  const w = useFileWorkspace(projectId), appearance = useAppearance();
  const review = usePocketReview(w);
  const [workspaceMode, setWorkspaceMode] = useState<'edit' | 'review'>('edit');
  const [reviewOpen, setReviewOpen] = useState<boolean | null>(null), [selectedReview, setSelectedReview] = useState('');
  const showReview = reviewOpen ?? !!review.review;
  const [composition, setComposition] = useState<ReviewComposition | null>(null);
  const galley = useRef<GalleyHandle>(null);
  const [rawMarkdown, setRawMarkdown] = useState(() => { try { return localStorage.getItem('edda.editor-mode') === 'source'; } catch { return false; } });
  const [sectionScope, setSectionScope] = useState('');
  const [treeOpen, setTreeOpen] = useState(false), [focus, setFocus] = useState(false), [historyOpen, setHistoryOpen] = useState(false), [query, setQuery] = useState('');
  const [translate, setTranslate] = useState(false), [sourceTab, setSourceTab] = useState(false), [source, setSource] = useState<{ id: string; body: string } | null>(null), [split, setSplit] = useState(50);
  const [modal, setModal] = useState<'create' | 'manage' | 'delete' | 'import' | 'restore' | null>(null);
  const [kind, setKind] = useState<'file' | 'directory'>('file'), [name, setName] = useState(''), [parent, setParent] = useState('');
  const [managed, setManaged] = useState<TreeEntry | null>(null), [past, setPast] = useState<ProjectVersion | null>(null);
  const [imports, setImports] = useState<ImportItem[]>([]), [omitted, setOmitted] = useState<string[]>([]);
  const editor = useRef<HTMLTextAreaElement>(null), sidebar = useRef<HTMLElement>(null), search = useRef<HTMLInputElement>(null), fileInput = useRef<HTMLInputElement>(null), folderInput = useRef<HTMLInputElement>(null);
  const archiveInput = useRef<HTMLInputElement>(null);
  const pending = useRef<{ signature: string; base: ProjectVersion; entries: TreeEntry[]; operation: string } | null>(null);
  const restoreOperation = useRef<{ signature: string; operation: string } | null>(null);
  const version = w.version, current = w.view, draft = w.draft;
  const entry = current && (version?.entries.find(e => e.id === current.entry.id) ?? current.entry);
  const folders = version?.entries.filter(e => e.kind === 'directory') ?? [];
  const scopePath = folders.find(folder => folder.id === sectionScope)?.path ?? '';
  const markdown = /\.(md|markdown)$/i.test(entry?.path ?? '');
  const isDirty = !!draft && dirtyDraft(draft);
  const reviewRecords = useMemo(() => review.review && !isDirty && draft?.body === review.review.source ? [...review.review.edits, ...review.review.signals] : [], [review.review, isDirty, draft?.body]);
  function revealReview(record: ReviewRecord) {
    setSelectedReview(record.id); setSourceTab(false);
    const range = reviewPositions(draft?.body ?? '', [record])[0];
    if (!range) return;
    if (!rawMarkdown) { galley.current?.select(range.from, range.to); galley.current?.scrollSelectionIntoView(); }
    else { editor.current?.focus(); editor.current?.setSelectionRange(range.from, range.to); }
    requestAnimationFrame(() => document.querySelector(`[data-record-id="${record.id}"]`)?.scrollIntoView({ block: 'nearest', inline: 'nearest' }));
  }
  function composeReview() {
    if (!review.review || isDirty) return;
    const selection = rawMarkdown ? { from: editor.current?.selectionStart ?? 0, to: editor.current?.selectionEnd ?? 0 } : galley.current?.getSelection();
    const selected = selection && reviewSelection(review.review.source, selection.from, selection.to);
    if (!selected) { w.setError('Выделите фрагмент текста для рецензии.'); return; }
    setComposition({ mode: 'add', selection: selected });
  }
  const reviewExtensions = useMemo(() => reviewMarks(draft?.body ?? '', reviewRecords, selectedReview, revealReview), [draft?.body, reviewRecords, selectedReview, rawMarkdown]);
  const words = (draft?.body ?? '').trim().split(/\s+/).filter(Boolean).length;
  useLayoutEffect(() => { const element = editor.current; if (!element) return; const resize = () => { element.style.height = 'auto'; element.style.height = `${element.scrollHeight}px`; }; resize(); const observer = new ResizeObserver(resize); observer.observe(element.parentElement!); return () => observer.disconnect(); }, [current?.entry.id, draft?.body, appearance.fontSize, appearance.serif, translate, sourceTab, rawMarkdown]);
  useEffect(() => {
    if (!treeOpen) return;
    const previous = document.activeElement instanceof HTMLElement ? document.activeElement : null;
    const element = sidebar.current; element?.querySelector<HTMLButtonElement>('button')?.focus();
    const key = (event: KeyboardEvent) => { if (document.querySelector('[role="dialog"]')) return; if (event.key === 'Escape') setTreeOpen(false); if (event.key !== 'Tab' || !element) return; const controls = Array.from(element.querySelectorAll<HTMLElement>('button:not(:disabled),input,a[href]')).filter(e => e.offsetParent !== null); if (event.shiftKey && document.activeElement === controls[0]) { event.preventDefault(); controls.at(-1)?.focus(); } else if (!event.shiftKey && document.activeElement === controls.at(-1)) { event.preventDefault(); controls[0]?.focus(); } };
    document.addEventListener('keydown', key); return () => { document.removeEventListener('keydown', key); previous?.isConnected && previous.focus(); };
  }, [treeOpen]);
  useEffect(() => { const key = (event: KeyboardEvent) => { if ((event.ctrlKey || event.metaKey) && event.key.toLowerCase() === 's') { event.preventDefault(); if (draft && dirtyDraft(draft)) void w.run(() => w.saveDraft(draft)); } if ((event.ctrlKey || event.metaKey) && event.key.toLowerCase() === 'k') { event.preventDefault(); if (matchMedia('(max-width:800px)').matches) setTreeOpen(true); setTimeout(() => search.current?.focus(), 0); } if (event.key === 'Escape') setFocus(false); }; window.addEventListener('keydown', key); return () => window.removeEventListener('keydown', key); }, [draft]);
  useEffect(() => { const media = matchMedia('(max-width:800px)'); const change = () => { if (!media.matches) setTreeOpen(false); }; media.addEventListener('change', change); return () => media.removeEventListener('change', change); }, []);
  function beginCreate(nextKind: 'file' | 'directory') { setKind(nextKind); setName(''); setParent(scopePath); w.setError(''); setModal('create'); }
  function open(next: TreeEntry) { if (version) void w.run(async () => { await w.openFile(next, version); setTreeOpen(false); }); }
  async function publishChange(key: string, build: () => Promise<TreeEntry[]>): Promise<ProjectVersion | null> {
    if (!version) return null;
    const signature = `${version.id}:${key}`;
    if (pending.current?.signature !== signature) pending.current = { signature, base: version, entries: await build(), operation: crypto.randomUUID() };
    const transaction = pending.current;
    const saved = await publishTree(transaction.base, transaction.entries, transaction.operation);
    if (!w.alive.current) return null;
    w.setVersion(saved); pending.current = null; w.setNotice('Изменения сохранены на сервере.'); return saved;
  }
  async function submitCreate() {
    if (!version) return;
    const path = parent ? `${parent}/${name.trim()}` : name.trim(); validatePath(path);
    if (version.entries.some(e => e.path === path)) throw new Error('Это имя уже занято.');
    if (dirname(path) && !folders.some(e => e.path === dirname(path))) throw new Error('Сначала создайте родительскую папку.');
    const saved = await publishChange(`create:${kind}:${path}`, async () => [...version.entries, { id: crypto.randomUUID(), path, kind, bytes: 0, ...(kind === 'file' ? await uploadText(projectId, '') : {}) }]);
    if (saved) { setModal(null); if (kind === 'file') { await w.openFile(saved.entries.find(e => e.path === path)!, saved); setTreeOpen(false); } }
  }
  async function move(id: string, target: string) { if (!version) return; await publishChange(`move:${id}:${target}`, async () => moveEntries(version.entries, id, target)); }
  function prepareImport(items: ImportItem[], destination: string) {
    setParent(destination); setImports(items.filter(item => !excludedPath(item.path))); setOmitted(items.filter(item => excludedPath(item.path)).map(item => item.path)); w.setError(''); setModal('import');
  }
  function drop(event: React.DragEvent, destination: string) {
    event.preventDefault(); event.stopPropagation(); if (w.busy || !version) return;
    const id = event.dataTransfer.getData('application/x-edda-entry');
    if (id) { const selected = version.entries.find(e => e.id === id); if (selected) void w.run(() => move(id, destination ? `${destination}/${basename(selected.path)}` : basename(selected.path))); return; }
    const collected = droppedItems(event.dataTransfer.items, event.dataTransfer.files);
    void w.run(async () => prepareImport(await collected, destination));
  }
  async function importFiles() {
    if (!version) return;
    const saved = await publishChange(`import:${parent}:${imports.map(i => i.id).join(',')}`, async () => {
      const entries = [...version.entries], paths = new Map(entries.map(e => [e.path, e]));
      // Validate the complete plan before uploading objects; publish only once.
      const planned: { item: ImportItem; path: string }[] = [];
      for (const item of imports) {
        const path = parent ? `${parent}/${item.path}` : item.path; validatePath(path);
        if (paths.has(path) && (item.file || item.staged?.kind === 'file' || paths.get(path)?.kind !== 'directory')) throw new Error(`«${path}» уже существует. Измените папку назначения или имя.`);
        const parts = path.split('/'); parts.pop();
        for (let index = 1; index <= parts.length; index++) { const folder = parts.slice(0, index).join('/'); if (paths.get(folder)?.kind === 'file') throw new Error(`Файл «${folder}» мешает созданию папки.`); if (!paths.has(folder)) { const directory: TreeEntry = { id: crypto.randomUUID(), path: folder, kind: 'directory', bytes: 0 }; paths.set(folder, directory); entries.push(directory); } }
        if (item.staged?.kind === 'file') { const staged = { ...item.staged, path }; paths.set(path, staged); entries.push(staged); }
        else if (!item.file) { if (!paths.has(path)) { const directory: TreeEntry = { id: item.id, path, kind: 'directory', bytes: 0 }; paths.set(path, directory); entries.push(directory); } }
        else { if (item.file.size > 64 * 1024 * 1024) throw new Error(`«${path}» превышает 64 МиБ.`); const placeholder: TreeEntry = { id: item.id, path, kind: 'file', bytes: item.file.size }; paths.set(path, placeholder); planned.push({ item, path }); }
      }
      for (const { item, path } of planned) entries.push({ id: item.id, path, kind: 'file', ...await uploadBlob(projectId, item.file!) });
      return entries;
    });
    if (saved) { setModal(null); setImports([]); const firstId = imports.find(item => item.file || item.staged?.kind === 'file')?.id; const first = saved.entries.find(e => e.id === firstId); if (first) { await w.openFile(first, saved); setTreeOpen(false); } }
  }
  function download() { if (!current) return; downloadFile(draft ? new Blob([draft.body], { type: 'text/plain;charset=utf-8' }) : current.blob, basename(entry?.path ?? current.entry.path)); }
  async function resolve(choice: 'mine' | 'theirs' | 'both') {
    const conflict = w.conflict; if (!conflict) return;
    const remote = conflict.head.entries.find(e => e.id === conflict.draft.entry.id);
    if (choice === 'theirs') { w.forget(conflict.draft.entry.id); if (remote?.kind === "file") await w.openFile(remote, conflict.head); else w.setView(null); w.setConflict(null); return; }
    if (choice === 'mine' && remote?.kind === 'file') { const rebased = { ...conflict.draft, version: conflict.head, entry: remote, operationId: crypto.randomUUID() }; w.keep(rebased); await w.saveDraft(rebased, false); return; }
    const baseName = basename(conflict.draft.entry.path), dot = baseName.lastIndexOf('.'), extension = dot > 0 ? baseName.slice(dot) : '', stem = dot > 0 ? baseName.slice(0, dot) : baseName;
    const path = `${stem} — мой черновик-${conflict.draft.operationId.slice(0, 8)}${extension}`;
    const id = conflict.draft.operationId;
    if (conflict.head.entries.some(e => e.path === path && e.id !== id)) throw new Error('Имя копии уже занято. Скачайте черновик и выберите другое имя.');
    const object = await uploadText(projectId, conflict.draft.body);
    const copy: TreeEntry = { id, path, kind: 'file', ...object };
    const saved = await publishTree(conflict.head, [...conflict.head.entries, copy], `${id}:keep-both:${conflict.head.id}`);
    w.setVersion(saved); w.forget(conflict.draft.entry.id); await w.openFile(copy, saved); w.setConflict(null); w.setNotice('Ваш текст сохранён отдельным файлом.');
  }
  return <div className={`edda ${focus ? 'focus-mode' : ''}`}>
    <input hidden type="file" multiple ref={fileInput} onChange={event => { if (event.target.files) prepareImport(Array.from(event.target.files).map(file => ({ file, path: file.name, id: crypto.randomUUID() })), scopePath); event.target.value = ''; }} />
    <input hidden type="file" multiple ref={node => { folderInput.current = node; node?.setAttribute('webkitdirectory', ''); }} onChange={event => { if (event.target.files) prepareImport(Array.from(event.target.files).map(file => ({ file, path: file.webkitRelativePath || file.name, id: crypto.randomUUID() })), scopePath); event.target.value = ''; }} />
    <input hidden type="file" accept=".zip,application/zip" aria-label="ZIP-архив" ref={archiveInput} onChange={event => { const file = event.target.files?.[0]; event.target.value = ''; if (file) void w.run(async () => { const plan = await stageArchive(projectId, file); if (!w.alive.current) return; prepareImport(plan.entries.map(staged => ({ id: staged.id, path: staged.path, staged })), scopePath); setOmitted(plan.omitted); }); }} />
    {treeOpen && <button className="tree-backdrop" aria-label="Закрыть список файлов" onClick={() => setTreeOpen(false)} />}
    <aside ref={sidebar} className={`sidebar ${treeOpen ? 'mobile-open' : ''}`} aria-label="Навигация проекта"><div className="sidebar-brand"><Link to="/projects" className="wordmark">edda<span className="brand-dot" /></Link><IconButton icon={X} label="Закрыть файлы" className="mobile-only" onClick={() => setTreeOpen(false)} /></div><div className="project-switch"><Link to="/projects" title={w.title}>{w.title}</Link><Link className="icon-button" aria-label="Настройки проекта" title="Настройки проекта" to={`/settings?projectId=${encodeURIComponent(projectId)}`}><Settings2 size={16} /></Link></div><label className="search-box"><input ref={search} value={query} onChange={event => setQuery(event.target.value)} aria-label="Найти файл" placeholder="Найти файл" /><kbd>Ctrl K</kbd></label><div className="tree-heading"><span>Файлы проекта</span><div><IconButton icon={Plus} label="Создать файл" disabled={w.busy || !version} onClick={() => beginCreate('file')} /><IconButton icon={FolderPlus} label="Создать папку" disabled={w.busy || !version} onClick={() => beginCreate('directory')} /><IconButton icon={Upload} label="Добавить файлы" disabled={w.busy || !version} onClick={() => fileInput.current?.click()} /></div></div>
      {version && <BookNavigation version={version} busy={w.busy} error={w.error} dirtyIds={new Set(w.dirty.map(d => d.entry.id))} run={w.run} onSaved={w.refresh} onOpen={open} />}
      {version && <WorkspaceSections version={version} busy={w.busy} error={w.error} dirtyIds={new Set(w.dirty.map(d => d.entry.id))} run={w.run} onSaved={w.refresh} scope={sectionScope} onScope={setSectionScope} />}
      {version ? <FileTree root={scopePath} entries={version.entries} selected={entry?.id} busy={w.busy} query={query} dirtyIds={new Set(w.dirty.map(d => d.entry.id))} onOpen={open} onManage={item => { setManaged(item); setName(basename(item.path)); setParent(dirname(item.path)); w.setError(''); setModal('manage'); }} onDrop={drop} /> : <p className="tree-empty" role="status">{w.error ? 'Не удалось загрузить файлы.' : 'Загружаем файлы…'}</p>}
      {w.dirty.filter(d => !version?.entries.some(e => e.id === d.entry.id)).map(d => <div className="recovered-drafts" key={d.entry.id}>Черновик удалённого файла<button className="open-draft" onClick={() => version && void w.run(() => w.openFile(d.entry, version))}>{d.entry.path}</button></div>)}
      <div className="sidebar-footer"><button className="subtle-action" disabled={w.busy || !version} onClick={() => folderInput.current?.click()}>Добавить папку</button><button className="subtle-action" disabled={w.busy || !version} onClick={() => archiveInput.current?.click()}>Импорт ZIP</button><AppearanceButton /></div></aside>
    <div className="work-area" inert={treeOpen} aria-busy={w.busy}><header className="workspace-header"><IconButton icon={Menu} label="Файлы проекта" className="mobile-only" onClick={() => setTreeOpen(true)} /><div className="breadcrumbs"><Link to="/projects">{w.title}</Link><ChevronRight size={13} /><span title={entry?.path}>{entry?.path ?? w.title}</span></div><div className="header-actions"><IconButton icon={RefreshCw} label="Обновить файлы" disabled={w.busy} onClick={() => void w.run(w.refresh)} />{current && <IconButton icon={ArrowDownToLine} label={isDirty ? 'Скачать черновик' : 'Скачать файл'} disabled={w.busy} onClick={download} />}{draft && <><IconButton icon={Languages} label="Оригинал и перевод" active={translate} aria-pressed={translate} onClick={() => { setTranslate(!translate); setSourceTab(false); }} /><IconButton icon={Maximize2} label={focus ? 'Выйти из режима сосредоточения' : 'Сосредоточиться на тексте'} aria-pressed={focus} onClick={() => { setFocus(!focus); setHistoryOpen(false); }} /></>}<IconButton icon={History} label="История проекта" active={historyOpen} disabled={!version} onClick={() => { setHistoryOpen(!historyOpen); if (!historyOpen) setReviewOpen(false); }} /></div></header>
      {w.error && !modal && <div className="error-banner" role="alert"><span>{w.error}</span><IconButton icon={X} label="Закрыть сообщение" onClick={() => w.setError('')} /></div>}
      {w.storageError && <div className="error-banner" role="alert">Не удалось сохранить копию в браузере. Сохраните текст на сервере или скачайте черновик перед выходом.</div>}
      <div className="work-body"><main className={`document-stage ${translate && draft ? 'translation-mode' : ''}`}>
        {current && entry ? draft && current.kind === 'text' ? <><div className="document-toolbar"><span><Pencil size={14} />{translate ? 'Перевод' : 'Рукопись'}</span>{translate && <label className="split-control">Ширина оригинала<input aria-label="Ширина оригинала" type="range" min={30} max={70} value={split} onChange={event => setSplit(Number(event.target.value))} /></label>}<div className="workspace-mode" role="group" aria-label="Режим работы"><button className="subtle-action" aria-pressed={workspaceMode === 'edit'} onClick={() => setWorkspaceMode('edit')}>Редактирование</button><button className="subtle-action" aria-pressed={workspaceMode === 'review'} onClick={() => { setWorkspaceMode('review'); setReviewOpen(true); setHistoryOpen(false); }}>Рецензирование</button></div>{review.available && <button className="subtle-action" aria-pressed={showReview} onClick={() => { setReviewOpen(!showReview); setHistoryOpen(false); }}>Рецензия</button>}{markdown && version && <ChapterMetadata key={entry.id} source={draft.body} version={version} entryId={entry.id} busy={w.busy} dirty={!!w.dirty.length} error={w.error} run={w.run} onSaved={async () => { await w.refresh(); review.clearHistory(); }} />}{markdown && <button className="subtle-action" aria-pressed={rawMarkdown} onClick={() => { setRawMarkdown(!rawMarkdown); try { localStorage.setItem('edda.editor-mode', rawMarkdown ? 'live' : 'source'); } catch { /* Optional preference. */ } }}>{rawMarkdown ? 'Живой Markdown' : 'Исходный Markdown'}</button>}<span className="save-label">{w.busy ? 'Работаем…' : isDirty ? 'Есть изменения' : 'Сохранено'}</span><IconButton icon={Save} label="Сохранить файл" disabled={w.busy || !isDirty} onClick={() => void w.run(() => w.saveDraft(draft))} /></div>{translate && <div className="translation-tabs" role="tablist" aria-label="Сторона перевода"><button role="tab" aria-selected={!sourceTab} onClick={() => setSourceTab(false)}>Перевод</button><button role="tab" aria-selected={sourceTab} onClick={() => setSourceTab(true)}>Оригинал</button></div>}
          <div className={`writing-columns ${sourceTab ? 'show-source' : ''}`} style={translate ? { gridTemplateColumns: `${split}fr ${100 - split}fr` } : undefined}>{translate && <section className="source-column" aria-label="Оригинал"><label className="column-label" htmlFor="source-file">Оригинал</label><select id="source-file" className="source-select" disabled={w.busy} value={source?.id ?? ''} onChange={event => { const item = version?.entries.find(e => e.id === event.target.value); if (!item || !version) { setSource(null); return; } void w.run(async () => { if (item.bytes > 1024 * 1024) throw new Error('Для сопоставления выберите текст до 1 МиБ.'); const text = w.drafts[item.id]?.body ?? new TextDecoder('utf-8', { fatal: true }).decode(await (await loadFile(version, item)).arrayBuffer()); if (text.includes('\0')) throw new Error('Для оригинала выберите текстовый файл.'); setSource({ id: item.id, body: text }); }); }}><option value="">Выберите исходный текст…</option>{version?.entries.filter(e => e.kind === 'file' && e.id !== entry.id).map(e => <option value={e.id} key={e.id}>{e.path}</option>)}</select><div className="source-prose" style={{ fontSize: appearance.fontSize }}>{source?.body ?? 'Оригинал останется перед глазами, пока вы работаете над переводом.'}</div></section>}<section className="prose-column" aria-label="Текст документа">{translate && <div className="column-label">Перевод</div>}<h1>{(markdown && frontmatter(draft.body)?.title) || basename(entry.path).replace(/\.(md|txt)$/i, '').replace(/^\d+\s*[—-]\s*/, '')}</h1><>{markdown && !rawMarkdown ? <MarkdownEditor key={entry.id} value={draft.body} onChange={w.edit} busy={w.busy} readOnly={workspaceMode === 'review'} editorRef={galley} extensions={reviewExtensions} /> : <textarea ref={editor} key={entry.id} className={`prose-editor ${appearance.serif ? '' : 'sans-prose'}`} style={{ fontSize: appearance.fontSize }} aria-label="Текст файла" value={editorText(draft.body)} readOnly={workspaceMode === 'review'} disabled={w.busy} onChange={event => w.edit(sourceText(draft.body, event.target.value))} placeholder="Начните с первой строки…" spellCheck />}</></section></div>
        </> : <FilePreview key={current.url} view={{ ...current, entry }} /> : <div className="empty-workspace"><FileText size={32} strokeWidth={1.2} /><h1>{w.title}</h1><p>{w.busy ? 'Открываем проект…' : 'Откройте файл слева или создайте первый текст.'}</p><button className="quiet-button" disabled={w.busy || !version} onClick={() => beginCreate('file')}><Plus size={17} />Новый текст</button></div>}
      </main>{draft && review.available && showReview && <ReviewPanel review={review.review} loading={review.loading} error={review.error} busy={w.busy} dirty={w.dirty.length > 0} canSave={isDirty} selected={selectedReview} onSelect={revealReview} onDecide={(action, id) => void w.run(() => review.decide(action, id))} onSave={() => void w.run(() => w.saveDraft(draft))} onClose={() => setReviewOpen(false)} onRetry={review.retry} onCompose={composeReview} onEdit={record => setComposition({ mode: 'update', record })} onNote={() => setComposition({ mode: 'note', note: review.review?.note ?? '' })} onUndo={direction => void w.run(() => review.undo(direction))} canUndo={review.canUndo} canRedo={review.canRedo} onStart={() => void w.run(review.start)} />}{historyOpen && version && <FileHistory version={version} busy={w.busy} run={w.run} onClose={() => setHistoryOpen(false)} onRestore={value => { setPast(value); setModal('restore'); w.setError(''); }} />}</div>
      <footer className="status-bar"><span role="status"><span className="status-dot" />{w.busy ? 'Работаем…' : !version ? 'Проект не загружен' : isDirty ? w.storageError ? 'Черновик только в памяти' : 'Черновик сохранён в этой вкладке' : 'Сохранено на сервере'}</span><span>{draft ? `${words} ${new Intl.PluralRules('ru').select(words) === 'one' ? 'слово' : new Intl.PluralRules('ru').select(words) === 'few' ? 'слова' : 'слов'}` : 'Edda'}</span></footer>
    </div>
    {w.notice && <div className="toast" role="status"><Check size={16} />{w.notice}</div>}
    <Modal open={composition !== null} title="Рецензия" description="Предложения и комментарии сохраняются отдельно от авторского текста." onClose={() => setComposition(null)} busy={w.busy}>{composition && <>{w.error && <p className="form-error" role="alert">{w.error}</p>}<ReviewComposer initial={composition} busy={w.busy} onSubmit={change => void w.run(async () => { await review.mutate(change); setComposition(null); })} /></>}</Modal>
    <Modal open={modal !== null} onClose={() => { setModal(null); w.setError(''); }} busy={w.busy} title={modal === 'create' ? kind === 'directory' ? 'Новая папка' : 'Новый текст' : modal === 'manage' ? 'Файл или папка' : modal === 'delete' ? 'Удалить из проекта?' : modal === 'restore' ? 'Восстановить проект?' : 'Добавить в проект'} description={modal === 'restore' ? 'Все файлы проекта будут возвращены к выбранной версии. Текущее состояние останется в истории, несохранённые черновики — в этой вкладке.' : modal === 'delete' ? 'Файл и содержимое папки исчезнут из текущего дерева. Предыдущие версии останутся доступны в истории.' : modal === 'import' ? 'Проверьте список. Существующие файлы не будут заменены.' : 'Структура проекта остаётся свободной. Перемещения сохраняют историю файла.'}>
      {w.error && <p className="form-error" role="alert">{w.error}</p>}
      {(modal === 'create' || modal === 'manage') && <form onSubmit={event => { event.preventDefault(); void w.run(async () => { if (modal === 'create') await submitCreate(); else if (managed) { await move(managed.id, parent ? `${parent}/${name.trim()}` : name.trim()); setModal(null); } }); }}><label className="field-label">Имя<input autoFocus required value={name} disabled={w.busy} onChange={event => setName(event.target.value)} /></label><label className="field-label">Папка<select value={parent} disabled={w.busy} onChange={event => setParent(event.target.value)}><option value="">Корень проекта</option>{folders.filter(e => modal !== 'manage' || !managed || (e.id !== managed.id && !e.path.startsWith(managed.path + '/'))).map(e => <option key={e.id} value={e.path}>{e.path}</option>)}</select></label><div className="dialog-actions">{modal === 'manage' && <button type="button" className="subtle-action danger-action" disabled={w.busy} onClick={() => setModal('delete')}>Удалить…</button>}<button className="quiet-button" type="submit" disabled={w.busy}>{modal === 'create' ? 'Создать' : 'Сохранить изменения'}</button></div></form>}
      {modal === 'delete' && managed && <><p>{managed.path}</p><div className="dialog-actions"><button className="quiet-button danger-action" disabled={w.busy || w.dirty.some(d => d.entry.id === managed.id || d.entry.path.startsWith(managed.path + '/'))} onClick={() => void w.run(async () => { if (!version) return; await publishChange(`delete:${managed.id}`, async () => version.entries.filter(e => e.id !== managed.id && !e.path.startsWith(managed.path + '/'))); setModal(null); await w.refresh(); })}>Удалить из проекта</button></div>{w.dirty.some(d => d.entry.id === managed.id || d.entry.path.startsWith(managed.path + '/')) && <p className="fine-print">Сначала сохраните черновики внутри этой папки или файла.</p>}</>}
      {modal === 'import' && <><label className="field-label">Папка назначения<select value={parent} disabled={w.busy} onChange={event => setParent(event.target.value)}><option value="">Корень проекта</option>{folders.map(e => <option key={e.id} value={e.path}>{e.path}</option>)}</select></label><ul className="pending-list">{imports.map(item => <li key={item.id}>{item.path}{item.file || item.staged?.kind === 'file' ? '' : '/'}</li>)}</ul>{omitted.length > 0 && <details><summary>Исключено локальных настроек и кэшей: {omitted.length}</summary><ul className="pending-list">{omitted.map(path => <li key={path}>{path}</li>)}</ul></details>}<div className="dialog-actions"><button className="quiet-button" disabled={w.busy || !imports.length} onClick={() => void w.run(importFiles)}>Добавить файлы</button></div></>}
      {modal === 'restore' && past && <><p>{new Date(past.createdAt).toLocaleString('ru')} · {past.entries.length} файлов и папок</p><div className="dialog-actions"><button className="quiet-button" disabled={w.busy} onClick={() => void w.run(async () => { if (!version) return; const signature = `${version.id}:${past.id}`; if (restoreOperation.current?.signature !== signature) restoreOperation.current = { signature, operation: crypto.randomUUID() }; const restored = await restoreVersion(version, past.id, restoreOperation.current.operation); if (!w.alive.current) return; w.setVersion(restored); setModal(null); setHistoryOpen(false); await w.refresh(); w.setNotice('Проект восстановлен. Предыдущее состояние сохранено в истории.'); })}>Восстановить весь проект</button></div></>}
    </Modal>
    <Modal open={w.conflict !== null} onClose={() => w.setConflict(null)} title="Две версии одного текста" description="Этот файл изменился на другом устройстве. Ваш черновик сохранён. Выберите, как продолжить." wide busy={w.busy}>{w.conflict && <><div className="conflict-columns"><section><h3>Ваш черновик</h3><p>{w.conflict.draft.body}</p></section><section><h3>Версия на сервере</h3><p>{w.conflict.remote ?? 'Файл удалён, слишком велик или больше не является текстовым.'}</p></section></div>{w.error && <p role="alert" className="form-error">{w.error}</p>}<div className="dialog-actions"><button className="subtle-action" onClick={download}>Скачать мой черновик</button><button className="quiet-button" disabled={w.busy || !w.conflict.head.entries.some(e => e.id === w.conflict?.draft.entry.id && e.kind === 'file')} onClick={() => void w.run(() => resolve('mine'))}>Сохранить мой текст</button><button className="quiet-button" disabled={w.busy} onClick={() => void w.run(() => resolve('both'))}>Сохранить оба</button><button className="subtle-action" disabled={w.busy} onClick={() => void w.run(() => resolve('theirs'))}>Отказаться от моего черновика</button></div></>}</Modal>
  </div>;
}

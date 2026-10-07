import { useEffect, useRef, useState } from 'react';
import { ApiError, listProjects } from '../../api';
import { loadFile, loadVersion, publishTree, uploadText, type ProjectVersion, type TreeEntry } from './fileApi';
import { dirtyDraft, inspectFile, readDrafts, type Draft, type FileView } from './fileModel';

export function useFileWorkspace(projectId: string) {
  const [version, setVersion] = useState<ProjectVersion | null>(null), [title, setTitle] = useState('Проект');
  const [view, setView] = useState<FileView | null>(null), [drafts, setDrafts] = useState(() => readDrafts(projectId));
  const draftsRef = useRef(drafts);
  const [busy, setBusy] = useState(false), [error, setError] = useState(''), [notice, setNotice] = useState(''), [storageError, setStorageError] = useState(false);
  const [conflict, setConflict] = useState<{ draft: Draft; head: ProjectVersion; remote: string | null } | null>(null);
  const busyRef = useRef(false), alive = useRef(true), epoch = useRef(0);
  const dirty = Object.values(drafts).filter(dirtyDraft);
  const draft = view ? drafts[view.entry.id] : undefined;
  function persist(next: Record<string, Draft>) {
    draftsRef.current = next; setDrafts(next);
    try { sessionStorage.setItem(`edda.drafts:${projectId}`, JSON.stringify(Object.fromEntries(Object.entries(next).filter(([, d]) => dirtyDraft(d))))); sessionStorage.removeItem(`open_edda_file_draft:${projectId}`); setStorageError(false); }
    catch { setStorageError(true); }
  }
  function keep(next: Draft) { persist({ ...draftsRef.current, [next.entry.id]: next }); }
  function forget(id: string) { const next = { ...draftsRef.current }; delete next[id]; persist(next); }
  async function run(work: () => Promise<void>) {
    if (busyRef.current || !alive.current) return;
    busyRef.current = true; setBusy(true); setError('');
    try { await work(); } catch (cause) { if (alive.current) setError(cause instanceof ApiError && cause.status === 409 ? 'Проект изменился на другом устройстве. Обновите файлы и повторите действие; черновики сохранятся.' : cause instanceof Error ? cause.message : 'Не удалось выполнить действие. Повторите попытку.'); }
    finally { busyRef.current = false; if (alive.current) setBusy(false); }
  }
  async function openFile(entry: TreeEntry, base: ProjectVersion) {
    const generation = epoch.current;
    const cached = draftsRef.current[entry.id];
    const blob = cached && dirtyDraft(cached) ? new Blob([cached.body]) : await loadFile(base, entry);
    const next = await inspectFile(entry, blob);
    if (!alive.current || generation !== epoch.current) { URL.revokeObjectURL(next.url); return; }
    if (next.kind === 'text' && (!cached || !dirtyDraft(cached))) keep({ version: base, entry, original: next.body ?? '', body: next.body ?? '', operationId: crypto.randomUUID() });
    setView(next);
    try { localStorage.setItem(`edda.resume:${projectId}`, entry.id); localStorage.setItem('edda.last-project', projectId); } catch { /* Opening never depends on preferences storage. */ }
  }
  useEffect(() => {
    alive.current = true; let cancelled = false;
    setBusy(true); busyRef.current = true;
    void (async () => {
      const [head, projects] = await Promise.all([loadVersion(projectId), listProjects()]);
      if (cancelled) return;
      setVersion(head); setTitle(projects.find(project => project.id === projectId)?.title ?? 'Проект');
      let resume = ''; try { resume = localStorage.getItem(`edda.resume:${projectId}`) ?? ''; localStorage.setItem('edda.last-project', projectId); } catch { /* Optional. */ }
      const recovered = Object.values(draftsRef.current).find(dirtyDraft);
      const selected = head.entries.find(entry => entry.id === resume && entry.kind === 'file') ?? recovered?.entry ?? head.entries.find(entry => entry.kind === 'file');
      if (selected) await openFile(selected, head);
      if (recovered) setNotice('Восстановлены черновики из этой вкладки.');
    })().catch(cause => { if (!cancelled) setError(cause instanceof Error ? cause.message : 'Не удалось загрузить проект.'); }).finally(() => { if (!cancelled) {busyRef.current = false; setBusy(false);} });
    return () => { cancelled = true; alive.current = false; busyRef.current = false; epoch.current++; };
  }, [projectId]);
  useEffect(() => () => { if (view) URL.revokeObjectURL(view.url); }, [view?.url]);
  useEffect(() => { if (!notice) return; const timer = setTimeout(() => setNotice(''), 4000); return () => clearTimeout(timer); }, [notice]);
  useEffect(() => { if (!dirty.length) return; const warn = (event: BeforeUnloadEvent) => event.preventDefault(); window.addEventListener('beforeunload', warn); return () => window.removeEventListener('beforeunload', warn); }, [dirty.length]);
  function edit(body: string) { if (draft && !busyRef.current) keep({ ...draft, body, operationId: crypto.randomUUID() }); }
  async function detectConflict(d: Draft, head: ProjectVersion) {
    const remoteEntry = head.entries.find(entry => entry.id === d.entry.id);
    let remote: string | null = null;
    if (remoteEntry?.kind === 'file' && remoteEntry.bytes <= 1024 * 1024) {
      try { remote = new TextDecoder('utf-8', { fatal: true }).decode(await (await loadFile(head, remoteEntry)).arrayBuffer()); } catch { remote = null; }
    }
    if (alive.current) { setVersion(head); setConflict({ draft: d, head, remote }); }
  }
  async function saveDraft(d: Draft, allowRebase = true) {
    if (new TextEncoder().encode(d.body).byteLength > 1024 * 1024) throw new Error('Текст превышает 1 МиБ. Скачайте черновик и продолжите локально.');
    const object = await uploadText(projectId, d.body);
    try {
      const saved = await publishTree(d.version, d.version.entries.map(entry => entry.id === d.entry.id ? { ...entry, ...object } : entry), d.operationId);
      if (!alive.current) return;
      const entry = saved.entries.find(entry => entry.id === d.entry.id)!;
      keep({ ...d, version: saved, entry, original: d.body, operationId: crypto.randomUUID() }); setVersion(saved); setConflict(null); setNotice('Сохранено на сервере. Предыдущая версия доступна в истории.');
      setView(previous => previous?.entry.id === entry.id ? { ...previous, entry, body: d.body } : previous);
    } catch (cause) {
      if (!(cause instanceof ApiError) || cause.status !== 409) throw cause;
      const head = await loadVersion(projectId), current = head.entries.find(entry => entry.id === d.entry.id);
      if (allowRebase && current?.kind === 'file' && current.sha256 === d.entry.sha256) {
        const rebased = { ...d, version: head, entry: current, operationId: crypto.randomUUID() }; keep(rebased); await saveDraft(rebased, false);
      } else await detectConflict(d, head);
    }
  }
  async function refresh() { const head = await loadVersion(projectId); if (!alive.current) return; setVersion(head); if (view) { const entry = head.entries.find(e => e.id === view.entry.id); if (entry) await openFile(entry, head); else if (!draft || !dirtyDraft(draft)) setView(null); } }
  return { version, setVersion, title, view, setView, draft, drafts, dirty, busy, error, setError, notice, setNotice, storageError, conflict, setConflict, run, openFile, keep, forget, edit, saveDraft, refresh, alive };
}

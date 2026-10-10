import { ApiError } from '../../api';
import { useEffect, useRef, useState } from 'react';
import { loadReview, changeReview, type PocketReview, type ReviewAction, type ReviewChange } from './pocketApi';
import type { useFileWorkspace } from './useFileWorkspace';

export function usePocketReview(w: ReturnType<typeof useFileWorkspace>) {
  const version = w.version, entry = w.view?.entry;
  const available = !!entry?.path.endsWith('.md');
  const key = available ? `${version?.id}:${entry?.id}` : '';
  const [state, setState] = useState<{ key: string; review: PocketReview | null; error: string }>({ key: '', review: null, error: '' });
  const [retry, setRetry] = useState(0);
  const pending = useRef<{ signature: string; action: ReviewAction } | null>(null);
  useEffect(() => {
    if (!available || !version || !entry) return;
    const abort = new AbortController();
    void loadReview(version, entry.id, abort.signal).then(review => {
      if (!abort.signal.aborted) setState({ key, review, error: '' });
    }).catch(cause => {
      if (!abort.signal.aborted) setState({ key, review: null, error: cause instanceof Error ? cause.message : 'Не удалось прочитать рецензию.' });
    });
    return () => abort.abort();
  }, [key, retry]);
  const review = state.key === key ? state.review : null;
  const [history, setHistory] = useState<Record<string, { undo: string[]; redo: string[] }>>({});
  const chapterHistory = (Object.hasOwn(history, entry?.id ?? '') ? history[entry?.id ?? ''] : undefined) ?? { undo: [], redo: [] };
  async function mutate(change: ReviewChange, direction?: 'undo' | 'redo') {
    if (!review || !version || w.dirty.length) return;
    const signature = `${review.chapter.id}:${JSON.stringify(change)}`;
    if (pending.current?.signature !== signature) pending.current = { signature, action: { ...change, expectedVersion: review.versionId, operationId: crypto.randomUUID(), entryId: review.chapter.id } };
    let saved;
    try { saved = await changeReview(version.projectId, pending.current.action); }
    catch (cause) { if (cause instanceof ApiError && cause.status === 409) pending.current = null; throw cause; }
    await w.refresh();
    pending.current = null;
    setHistory(previous => {
      const old = (Object.hasOwn(previous, review.chapter.id) ? previous[review.chapter.id] : undefined) ?? { undo: [], redo: [] };
      const next = direction === 'undo' ? { undo: old.undo.slice(0, -1), redo: [...old.redo, saved.id] } : direction === 'redo' ? { undo: [...old.undo, saved.id], redo: old.redo.slice(0, -1) } : { undo: [...old.undo, saved.id].slice(-100), redo: [] };
      return { ...previous, [review.chapter.id]: next };
    });
    w.setNotice(direction === 'undo' ? 'Действие с рецензией отменено.' : direction === 'redo' ? 'Действие повторено.' : 'Рецензия сохранена.');
  }
  async function start() {
    if (!version || !entry || w.dirty.length) return;
    const signature = `${entry.id}:start`;
    if (pending.current?.signature !== signature) pending.current = { signature, action: { action: 'start', expectedVersion: version.id, operationId: crypto.randomUUID(), entryId: entry.id } };
    try { await changeReview(version.projectId, pending.current.action); }
    catch (cause) { if (cause instanceof ApiError && cause.status === 409) pending.current = null; throw cause; }
    await w.refresh(); pending.current = null;
  }
  async function decide(action: 'apply' | 'remove', recordId: string) {
    if (!review || !version || w.dirty.length) return;
    await mutate({ action, recordId });
  }
  async function undo(direction: 'undo' | 'redo') {
    const actionVersion = chapterHistory[direction].at(-1);
    if (actionVersion) await mutate({ action: 'undo', actionVersion }, direction);
  }

  return { clearHistory: () => { if (entry) setHistory(previous => ({ ...previous, [entry.id]: { undo: [], redo: [] } })); }, available, review, start, mutate, undo, canUndo: !!chapterHistory.undo.length, canRedo: !!chapterHistory.redo.length, loading: available && state.key !== key, error: state.key === key ? state.error : '', retry: () => { setState({ key: '', review: null, error: '' }); setRetry(n => n + 1); }, decide };
}

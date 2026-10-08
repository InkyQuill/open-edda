import { useEffect, useMemo, useState } from 'react';
import { Modal, downloadFile } from '../../shared/ui/edda';
import { diffLines } from '../review/lineDiff';
import { loadFile, type ProjectVersion, type TreeEntry } from './fileApi';
import { compareTrees, type EntryChange } from './versionDiff';

type Content = { text: string | null; blob?: Blob; entry?: TreeEntry };
async function read(version: ProjectVersion, entry?: TreeEntry): Promise<Content> {
  if (!entry || entry.kind === 'directory') return { text: '', entry };
  if (entry.bytes > 1024 * 1024) return { text: null, entry };
  const blob = await loadFile(version, entry);
  try { const text = new TextDecoder('utf-8', { fatal: true, ignoreBOM: true }).decode(await blob.arrayBuffer()); return { text: text.includes('\0') ? null : text, blob, entry }; }
  catch { return { text: null, blob, entry }; }
}
export function VersionComparison({ before, after, onClose }: { before: ProjectVersion; after: ProjectVersion; onClose: () => void }) {
  const changes = useMemo(() => compareTrees(before.entries, after.entries), [before, after]);
  const [selected, setSelected] = useState<EntryChange | null>(changes[0] ?? null);
  const [content, setContent] = useState<{ before: Content; after: Content } | null>(null);
  const [error, setError] = useState('');
  useEffect(() => {
    let cancelled = false; setContent(null); setError('');
    if (selected) void Promise.all([read(before, selected.before), read(after, selected.after)]).then(([a, b]) => { if (!cancelled) setContent({ before: a, after: b }); }).catch(cause => { if (!cancelled) setError(cause instanceof Error ? cause.message : 'Не удалось сравнить файлы.'); });
    return () => { cancelled = true; };
  }, [selected, before, after]);
  const lines = useMemo(() => content && content.before.text !== null && content.after.text !== null ? diffLines(content.before.text, content.after.text) : null, [content]);
  async function download(version: ProjectVersion, value: Content) {
    if (!value.entry) return;
    try { downloadFile(value.blob ?? await loadFile(version, value.entry), value.entry.path.split('/').at(-1)!); }
    catch (cause) { setError(cause instanceof Error ? cause.message : 'Не удалось скачать файл.'); }
  }
  return <Modal open onClose={onClose} title="Сравнение версий" description="Из выбранной версии в текущую сохранённую версию. Черновики не участвуют в сравнении." wide>
    <p className="fine-print">{new Date(before.createdAt).toLocaleString('ru')} → {new Date(after.createdAt).toLocaleString('ru')}</p>
    {!changes.length ? <p>Деревья проекта совпадают.</p> : <><label className="field-label">Изменения проекта<select value={selected?.id ?? ''} onChange={event => setSelected(changes.find(change => change.id === event.target.value) ?? null)}>{changes.map(change => <option key={change.id} value={change.id}>{change.label}: {(change.after ?? change.before)!.path}</option>)}</select></label>
    {selected && <p className="comparison-path">{selected.label}: {selected.before?.path ?? '∅'} → {selected.after?.path ?? '∅'}</p>}
    {error && <p role="alert" className="form-error">{error}</p>}
    {!content && !error && <p role="status">Сравниваем…</p>}
    {content && <>{lines ? <><p className="fine-print">− удалено · + добавлено</p><pre className="version-diff" aria-label="Изменения текста">{lines.map((line, index) => <span key={index} className={`diff-${line.kind}`}>{line.kind === 'added' ? '+ ' : line.kind === 'removed' ? '− ' : '  '}{line.text}{'\n'}</span>)}</pre></> : <p>Бинарный файл или текст больше 1 МиБ. Скачайте версии для сравнения.</p>}
    <div className="dialog-actions">{content.before.entry?.kind === 'file' && <button className="subtle-action" onClick={() => void download(before, content.before)}>Скачать прежний файл</button>}{content.after.entry?.kind === 'file' && <button className="subtle-action" onClick={() => void download(after, content.after)}>Скачать текущий файл</button>}</div></>}
    </>}
  </Modal>;
}

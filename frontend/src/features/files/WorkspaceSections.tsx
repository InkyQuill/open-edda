import { ApiError } from '../../api';
import { useEffect, useRef, useState } from 'react';
import { Modal } from '../../shared/ui/edda';
import { loadFile, publishTree, uploadText, type ProjectVersion } from './fileApi';
import { decodeSections, type WorkspaceSection, type WorkspaceConfig } from './workspaceSectionsModel';

const configPath = 'edda.workspace.json';
const defaults = () => ['База знаний', 'Черновики', 'Отрывки', 'Основной текст'].map((label, index) => ({ id: crypto.randomUUID(), label, folderId: '-', purpose: ['knowledge', 'drafts', 'fragments', 'manuscript'][index] }));
export function WorkspaceSections({ version, busy, dirtyIds, error, run, onSaved, scope, onScope }: {
  version: ProjectVersion; busy: boolean; dirtyIds: Set<string>; error: string; run: (work: () => Promise<void>) => Promise<void>; onSaved: () => Promise<void>; scope: string; onScope: (id: string) => void;
}) {
  const entry = version.entries.find(e => e.path === configPath);
  const key = `${version.projectId}:${entry?.id ?? ''}:${entry?.kind ?? ''}:${entry?.sha256 ?? 'absent'}`;
  const [loaded, setLoaded] = useState<{ key: string; config: WorkspaceConfig; error: string } | null>(null);
  const [open, setOpen] = useState(false), [rows, setRows] = useState<WorkspaceSection[]>([]);
  const pending = useRef<{ signature: string; operation: string; id: string; base: ProjectVersion } | null>(null);
  useEffect(() => {
    let cancelled = false;
    void (async () => {
      if (entry && (entry.kind !== 'file' || entry.bytes > 65536)) throw new Error('Настройки разделов должны быть JSON-файлом до 64 КиБ.');
      const config = entry ? decodeSections(await (await loadFile(version, entry)).text()) : { schemaVersion: 1 as const, sections: [] };
      if (!cancelled) setLoaded({ key, config, error: '' });
    })().catch(cause => { if (!cancelled) setLoaded({ key, config: { schemaVersion: 1, sections: [] }, error: cause instanceof Error ? cause.message : 'Не удалось прочитать разделы.' }); });
    return () => { cancelled = true; };
  }, [key]);
  const current = loaded?.key === key ? loaded : null;
  const folders = version.entries.filter(e => e.kind === 'directory');
  async function save() {
    if (!current || current.error || (entry && dirtyIds.has(entry.id))) return;
    const body = JSON.stringify({ ...current.config, sections: rows.filter(row => row.folderId !== '-').map(row => ({ ...row, label: row.label.trim() })) }, null, 2) + '\n';
    decodeSections(body);
    const signature = `${version.projectId}:${body}`;
    if (pending.current?.signature !== signature) pending.current = { signature, operation: crypto.randomUUID(), id: entry?.id ?? crypto.randomUUID(), base: version };
    const upload = await uploadText(version.projectId, body);
    try { await publishTree(pending.current.base, [...pending.current.base.entries.filter(e => e.path !== configPath), { id: pending.current.id, path: configPath, kind: 'file', ...upload }], pending.current.operation); }
    catch (cause) { if (cause instanceof ApiError && cause.status === 409) pending.current = null; throw cause; }
    await onSaved(); pending.current = null; setOpen(false);
  }
  return <section className="workspace-sections" aria-label="Разделы книги">
    <button className="subtle-action" disabled={busy || !current} onClick={() => { setRows(current?.config.sections.length ? current.config.sections : defaults()); setOpen(true); }}>Настроить разделы книги</button>
    {current?.error && <p className="form-error">{current.error}</p>}
    {!!current?.config.sections.length && <nav aria-label="Назначение папок"><button aria-pressed={!scope} onClick={() => onScope('')}>Все файлы</button>{current.config.sections.map(section => {
      const folder = folders.find(e => e.id === section.folderId), missing = section.folderId !== '' && !folder;
      return <button key={section.id} disabled={missing} title={missing ? 'Папка не найдена — измените назначение' : folder?.path ?? 'Весь проект'} aria-pressed={scope === section.folderId} onClick={() => onScope(section.folderId)}>{section.label}{missing ? ' · папка не найдена' : ''}</button>;
    })}</nav>}
    <Modal open={open} busy={busy} onClose={() => setOpen(false)} title="Разделы книги" description="Выберите папки и назовите разделы так, как удобно вам. Никаких обязательных этапов: текст может оставаться в работе столько, сколько вы решите." wide>
      {(current?.error || error) && <p className="form-error" role="alert">{current?.error || error}</p>}
      {entry && dirtyIds.has(entry.id) && <p className="fine-print">Сначала сохраните открытый черновик настроек разделов.</p>}
      <form onSubmit={event => { event.preventDefault(); void run(save); }}><div className="section-settings">{rows.map((row, i) => <div key={row.id}>
        <label className="field-label">Название раздела<input aria-label={`Название раздела ${i + 1}`} maxLength={100} value={row.label} disabled={busy} onChange={event => setRows(rows.map(r => r.id === row.id ? { ...r, label: event.target.value } : r))} /></label>
        <label className="field-label">Назначение<select aria-label={`Назначение раздела ${i + 1}`} value={row.purpose ?? 'custom'} disabled={busy} onChange={event => setRows(rows.map(r => r.id === row.id ? { ...r, purpose: event.target.value } : r))}><option value="knowledge">База знаний</option><option value="drafts">Черновики</option><option value="fragments">Отрывки</option><option value="manuscript">Основной текст</option><option value="custom">Другое</option>{row.purpose && !['knowledge', 'drafts', 'fragments', 'manuscript', 'custom'].includes(row.purpose) && <option value={row.purpose}>{row.purpose}</option>}</select></label>
        <label className="field-label">Папка<select aria-label={`Папка раздела ${i + 1}`} value={row.folderId} disabled={busy} onChange={event => setRows(rows.map(r => r.id === row.id ? { ...r, folderId: event.target.value } : r))}><option value="-">Не использовать</option><option value="">Весь проект</option>{row.folderId !== '-' && row.folderId !== '' && !folders.some(f => f.id === row.folderId) && <option value={row.folderId}>Папка не найдена</option>}{folders.map(folder => <option key={folder.id} value={folder.id}>{folder.path}</option>)}</select></label>
        <button className="subtle-action" type="button" disabled={busy} aria-label={`Удалить раздел ${i + 1}`} onClick={() => setRows(rows.filter(r => r.id !== row.id))}>Убрать</button>
      </div>)}</div><button className="subtle-action" type="button" disabled={busy || rows.length >= 50} onClick={() => setRows([...rows, { id: crypto.randomUUID(), label: '', folderId: '-', purpose: 'custom' }])}>Добавить раздел</button><div className="dialog-actions"><button className="quiet-button" disabled={busy || !current || !!current.error || (!!entry && dirtyIds.has(entry.id))}>Сохранить разделы</button></div></form>
    </Modal>
  </section>;
}

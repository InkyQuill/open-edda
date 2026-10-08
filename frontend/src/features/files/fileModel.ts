import type { ProjectVersion, TreeEntry } from './fileApi';
export const basename = (path: string) => path.split('/').at(-1) ?? path;
export const dirname = (path: string) => path.slice(0, Math.max(0, path.lastIndexOf('/')));
export const excludedPath = (path: string) => path.split('/').some(part => ['.git', '.edda', 'node_modules', '__pycache__', '.DS_Store', '.env'].includes(part) || part.startsWith('.env.'));
export function validatePath(path: string) {
  if (!path || path.startsWith('/') || path.endsWith('/') || path.includes('\\') || /[\u0000-\u001f\u007f]/.test(path) || path.split('/').some(part => !part || part === '.' || part === '..')) throw new Error('Введите корректное имя или путь без пустых частей, . и ..');
  if (excludedPath(path)) throw new Error('Этот путь предназначен для локальных настроек или кэша.');
}
export function moveEntries(entries: TreeEntry[], id: string, target: string): TreeEntry[] {
  validatePath(target);
  const source = entries.find(entry => entry.id === id);
  if (!source) throw new Error('Файл больше не существует. Обновите список.');
  if (source.path === target) return entries;
  if (source.kind === 'directory' && target.startsWith(source.path + '/')) throw new Error('Нельзя переместить папку внутрь самой себя.');
  const parent = dirname(target);
  if (parent && !entries.some(entry => entry.path === parent && entry.kind === 'directory')) throw new Error('Папка назначения не существует.');
  const result = entries.map(entry => entry.id === source.id || entry.path.startsWith(source.path + '/') ? { ...entry, path: target + entry.path.slice(source.path.length) } : entry);
  if (new Set(result.map(entry => entry.path)).size !== result.length) throw new Error('В папке назначения уже есть файл или папка с таким именем.');
  return result;
}
export type Draft = { version: ProjectVersion; entry: TreeEntry; original: string; body: string; operationId: string };
export const dirtyDraft = (draft: Draft) => draft.original !== draft.body;
function isDraft(value: unknown, projectId: string): value is Draft {
  if (!value || typeof value !== 'object') return false;
  const d = value as Partial<Draft>;
  return typeof d.body === 'string' && typeof d.original === 'string' && typeof d.operationId === 'string' && typeof d.entry?.id === 'string' && typeof d.entry?.path === 'string' && d.version?.projectId === projectId && Array.isArray(d.version.entries);
}
export function readDrafts(projectId: string): Record<string, Draft> {
  try {
    const value: unknown = JSON.parse(sessionStorage.getItem(`edda.drafts:${projectId}`) ?? 'null');
    const drafts: Record<string, Draft> = {};
    if (value && typeof value === 'object') for (const d of Object.values(value)) if (isDraft(d, projectId) && dirtyDraft(d)) drafts[d.entry.id] = d;
    const legacy: unknown = JSON.parse(sessionStorage.getItem(`open_edda_file_draft:${projectId}`) ?? 'null');
    if (isDraft(legacy, projectId) && dirtyDraft(legacy) && !drafts[legacy.entry.id]) drafts[legacy.entry.id] = legacy;
    return drafts;
  } catch { return {}; }
}
export type FileView = { entry: TreeEntry; kind: 'text' | 'image' | 'audio' | 'video' | 'pdf' | 'binary'; blob: Blob; body?: string; url: string };
const mimeByExtension: Record<string, string> = { png: 'image/png', jpg: 'image/jpeg', jpeg: 'image/jpeg', gif: 'image/gif', webp: 'image/webp', avif: 'image/avif', svg: 'image/svg+xml', bmp: 'image/bmp', ico: 'image/x-icon', pdf: 'application/pdf', mp3: 'audio/mpeg', wav: 'audio/wav', ogg: 'audio/ogg', flac: 'audio/flac', m4a: 'audio/mp4', mp4: 'video/mp4', webm: 'video/webm', mov: 'video/quicktime' };
export async function inspectFile(entry: TreeEntry, blob: Blob): Promise<FileView> {
  const extension = entry.path.split('.').at(-1)?.toLowerCase() ?? '';
  const mime = mimeByExtension[extension];
  let kind: FileView['kind'] = mime?.startsWith('image/') ? 'image' : mime?.startsWith('audio/') ? 'audio' : mime?.startsWith('video/') ? 'video' : mime === 'application/pdf' ? 'pdf' : 'binary';
  let body: string | undefined;
  if (!mime && blob.size <= 1024 * 1024 && !['zip', 'gz', '7z', 'rar', 'docx', 'xlsx', 'epub', 'woff', 'woff2', 'exe'].includes(extension)) {
    try { const text = new TextDecoder('utf-8', { fatal: true, ignoreBOM: true }).decode(await blob.arrayBuffer()); if (!text.includes('\0')) { body = text; kind = 'text'; } } catch { /* Binary content is a normal file page. */ }
  }
  const typed = new Blob([blob], { type: mime ?? (kind === 'text' ? 'text/plain;charset=utf-8' : 'application/octet-stream') });
  return { entry, kind, blob: typed, body, url: URL.createObjectURL(typed) };
}
export type ImportItem = { path: string; file?: File; staged?: TreeEntry; id: string };
export async function droppedItems(items: DataTransferItemList, fallback: FileList): Promise<ImportItem[]> {
  const roots = Array.from(items).filter(item => item.kind === 'file').map(item => item.webkitGetAsEntry?.()).filter((entry): entry is FileSystemEntry => !!entry);
  if (!roots.length) return Array.from(fallback).map(file => ({ path: file.name, file, id: crypto.randomUUID() }));
  const result: ImportItem[] = [];
  async function walk(entry: FileSystemEntry, parent: string) {
    const path = parent + entry.name;
    if (entry.isFile) { const file = await new Promise<File>((resolve, reject) => (entry as FileSystemFileEntry).file(resolve, reject)); result.push({ path, file, id: crypto.randomUUID() }); }
    else if (entry.isDirectory) { result.push({ path, id: crypto.randomUUID() }); const reader = (entry as FileSystemDirectoryEntry).createReader(); for (;;) { const children = await new Promise<FileSystemEntry[]>((resolve, reject) => reader.readEntries(resolve, reject)); if (!children.length) break; for (const child of children) await walk(child, path + '/'); } }
  }
  for (const entry of roots) await walk(entry, '');
  return result;
}

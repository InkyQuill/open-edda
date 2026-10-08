import type { TreeEntry } from './fileApi';
export type EntryChange = { id: string; before?: TreeEntry; after?: TreeEntry; label: string };
export function compareTrees(before: TreeEntry[], after: TreeEntry[]): EntryChange[] {
  const old = new Map(before.map(entry => [entry.id, entry]));
  const current = new Map(after.map(entry => [entry.id, entry]));
  return [...new Set([...old.keys(), ...current.keys()])].flatMap(id => {
    const a = old.get(id), b = current.get(id);
    const labels = !a ? ['Добавлено'] : !b ? ['Удалено'] : [a.path !== b.path ? 'Перемещено' : '', a.kind !== b.kind || a.sha256 !== b.sha256 || a.bytes !== b.bytes ? 'Изменено' : ''].filter(Boolean);
    return labels.length ? [{ id, before: a, after: b, label: labels.join(' · ') }] : [];
  }).sort((a, b) => (a.after ?? a.before)!.path.localeCompare((b.after ?? b.before)!.path));
}

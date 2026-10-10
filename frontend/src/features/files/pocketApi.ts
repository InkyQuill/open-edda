import { apiError, authFetch } from '../../api';
import type { ProjectVersion, TreeEntry } from './fileApi';

export type ReviewRecord = {
  id: string; type?: 'note' | 'change_required' | 'warning' | 'review';
  before?: string; after?: string; selected_text?: string; comment?: string;
  resolution: { kind: 'resolved' | 'stale' | 'ambiguous' | 'conflicting'; range?: { from: number; to: number } };
};
export type PocketReview = { versionId: string; chapter: TreeEntry; sidecar: TreeEntry; note: string; source: string; edits: ReviewRecord[]; signals: ReviewRecord[] };
export type ReviewAction = { expectedVersion: string; operationId: string; entryId: string; action: 'apply' | 'remove' | 'add' | 'update' | 'note' | 'undo' | 'start' | 'metadata'; recordId?: string; from?: number; to?: number; selected?: string; after?: string; signalType?: string; comment?: string; note?: string; metadata?: string; actionVersion?: string };
export type ReviewChange = Omit<ReviewAction, 'expectedVersion' | 'operationId' | 'entryId'>;
const root = (project: string) => `/api/projects/${encodeURIComponent(project)}/files`;
export async function loadReview(version: ProjectVersion, entryId: string, signal: AbortSignal): Promise<PocketReview | null> {
  const response = await authFetch(`${root(version.projectId)}/versions/${encodeURIComponent(version.id)}/review/${encodeURIComponent(entryId)}`, { signal });
  if (response.status === 404) return null;
  if (!response.ok) throw await apiError('Не удалось прочитать рецензию', response);
  return response.json();
}
export async function changeReview(project: string, action: ReviewAction): Promise<ProjectVersion> {
  const response = await authFetch(`${root(project)}/review`, { method: 'POST', body: JSON.stringify(action) });
  if (!response.ok) throw await apiError('Не удалось сохранить решение', response);
  return response.json();
}

export async function changeBook(project: string, input: { expectedVersion: string; operationId: string; entryId: string; title: string; createPaths?: string[]; chapters: { id: string; path: string; title?: string }[] }): Promise<ProjectVersion> {
 const response = await authFetch(`${root(project)}/book`, { method: 'POST', body: JSON.stringify(input) });
 if (!response.ok) throw await apiError('Не удалось сохранить книгу', response);
 return response.json();
}

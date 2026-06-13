import type { ContentItem, ContentKind, StoryProject } from "./types";

export async function listProjects(): Promise<StoryProject[]> {
  const response = await fetch("/api/projects");
  if (!response.ok) {
    throw new Error(`list projects failed: ${response.status}`);
  }
  return response.json() as Promise<StoryProject[]>;
}

export async function listContent(projectId: string, kind: ContentKind, signal?: AbortSignal): Promise<ContentItem[]> {
  const params = new URLSearchParams({ kind });
  const response = await fetch(`/api/projects/${encodeURIComponent(projectId)}/content?${params}`, { signal });
  if (!response.ok) {
    throw new Error(`list content failed: ${response.status}`);
  }
  return response.json() as Promise<ContentItem[]>;
}

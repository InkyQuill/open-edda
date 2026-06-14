import { getToken } from "./authApi";
import type { ContentItem, ContentKind, StoryProject } from "./types";

async function authFetch(path: string, init?: RequestInit): Promise<Response> {
  const token = getToken();
  const headers = new Headers(init?.headers);
  if (!headers.has("Content-Type") && init?.method && init.method !== "GET") {
    headers.set("Content-Type", "application/json");
  }
  if (token) headers.set("Authorization", `Bearer ${token}`);
  return fetch(path, { ...init, headers });
}

export async function listProjects(): Promise<StoryProject[]> {
  const response = await authFetch("/api/projects");
  if (!response.ok) {
    throw new Error(`list projects failed: ${response.status}`);
  }
  return response.json() as Promise<StoryProject[]>;
}

export async function listContent(projectId: string, kind: ContentKind, signal?: AbortSignal): Promise<ContentItem[]> {
  const params = new URLSearchParams({ kind });
  const response = await authFetch(`/api/projects/${encodeURIComponent(projectId)}/content?${params}`, { signal });
  if (!response.ok) {
    throw new Error(`list content failed: ${response.status}`);
  }
  return response.json() as Promise<ContentItem[]>;
}

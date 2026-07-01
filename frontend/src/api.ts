import { getToken } from "./authApi";
import type { ContentItem, ContentKind, Revision, StoryProject } from "./types";

async function authFetch(path: string, init?: RequestInit): Promise<Response> {
  const token = getToken();
  const headers = new Headers(init?.headers);
  if (!headers.has("Content-Type") && init?.body) {
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

export type CreateProjectInput = {
  title: string;
  language: string;
};

export async function createProject(input: CreateProjectInput): Promise<StoryProject> {
  const response = await authFetch("/api/projects", {
    method: "POST",
    body: JSON.stringify(input),
  });
  if (!response.ok) {
    throw new Error(`create project failed: ${response.status}`);
  }
  return response.json() as Promise<StoryProject>;
}

export async function importElysiumProject(file: File): Promise<StoryProject> {
  const response = await authFetch("/api/projects/import/elysium", {
    method: "POST",
    body: file,
    headers: new Headers({ "Content-Type": "application/zip" }),
  });
  if (!response.ok) {
    throw new Error(`import Elysium project failed: ${response.status}`);
  }
  return response.json() as Promise<StoryProject>;
}

export type CreateContentInput = {
  kind: ContentKind;
  title: string;
  bodyMarkdown: string;
  metadataJson: string;
  sortOrder: number;
  reason: string;
};

export async function createContent(
  projectId: string,
  input: CreateContentInput,
  signal?: AbortSignal,
): Promise<ContentItem> {
  const response = await authFetch(`/api/projects/${encodeURIComponent(projectId)}/content`, {
    method: "POST",
    body: JSON.stringify(input),
    signal,
  });
  if (!response.ok) {
    throw new Error(`create content failed: ${response.status}`);
  }
  return response.json() as Promise<ContentItem>;
}

export async function listContent(projectId: string, kind: ContentKind, signal?: AbortSignal): Promise<ContentItem[]> {
  const params = new URLSearchParams({ kind });
  const response = await authFetch(`/api/projects/${encodeURIComponent(projectId)}/content?${params}`, { signal });
  if (!response.ok) {
    throw new Error(`list content failed: ${response.status}`);
  }
  return response.json() as Promise<ContentItem[]>;
}

export type UpdateContentInput = {
  expectedRevision: number;
  bodyMarkdown: string;
  metadataJson: string;
  reason: string;
};

export async function updateContent(
  projectId: string,
  contentId: string,
  input: UpdateContentInput,
  signal?: AbortSignal,
): Promise<ContentItem> {
  const response = await authFetch(
    `/api/projects/${encodeURIComponent(projectId)}/content/${encodeURIComponent(contentId)}`,
    {
      method: "PUT",
      body: JSON.stringify(input),
      signal,
    },
  );
  if (!response.ok) {
    throw new Error(`update content failed: ${response.status}`);
  }
  return response.json() as Promise<ContentItem>;
}

export async function listRevisions(projectId: string, contentId: string, signal?: AbortSignal): Promise<Revision[]> {
  const response = await authFetch(
    `/api/projects/${encodeURIComponent(projectId)}/content/${encodeURIComponent(contentId)}/revisions`,
    { signal },
  );
  if (!response.ok) {
    throw new Error(`list revisions failed: ${response.status}`);
  }
  return response.json() as Promise<Revision[]>;
}

export type RestoreRevisionInput = {
  expectedRevision: number;
  reason: string;
};

export async function restoreRevision(
  projectId: string,
  contentId: string,
  revisionNumber: number,
  input: RestoreRevisionInput,
  signal?: AbortSignal,
): Promise<ContentItem> {
  const response = await authFetch(
    `/api/projects/${encodeURIComponent(projectId)}/content/${encodeURIComponent(contentId)}/revisions/${encodeURIComponent(String(revisionNumber))}/restore`,
    {
      method: "POST",
      body: JSON.stringify(input),
      signal,
    },
  );
  if (!response.ok) {
    throw new Error(`restore revision failed: ${response.status}`);
  }
  return response.json() as Promise<ContentItem>;
}

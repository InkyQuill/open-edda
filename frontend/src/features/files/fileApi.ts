import { apiError, authFetch } from "../../api";

export type TreeEntry = { id: string; path: string; kind: "file" | "directory"; sha256?: string; bytes: number };
export type ProjectVersion = { id: string; projectId: string; parentId: string; operationId: string; message: string; createdAt: string; entries: TreeEntry[] };
const root = (projectId: string) => `/api/projects/${encodeURIComponent(projectId)}/files`;
export async function loadVersion(projectId: string, versionId = "current"): Promise<ProjectVersion> {
  const response = await authFetch(`${root(projectId)}/versions/${encodeURIComponent(versionId)}`);
  if (!response.ok) throw await apiError("Load project files", response);
  return response.json() as Promise<ProjectVersion>;
}
export async function loadFile(version: ProjectVersion, entry: TreeEntry): Promise<Blob> {
  const response = await authFetch(`${root(version.projectId)}/versions/${encodeURIComponent(version.id)}/entries/${encodeURIComponent(entry.id)}`);
  if (!response.ok) throw await apiError("Open file", response);
  return response.blob();
}
export async function uploadText(projectId: string, body: string): Promise<{ sha256: string; bytes: number }> {
  const data = new TextEncoder().encode(body);
  const digest = await crypto.subtle.digest("SHA-256", data);
  const sha256 = Array.from(new Uint8Array(digest), byte => byte.toString(16).padStart(2, "0")).join("");
  const response = await authFetch(`${root(projectId)}/objects/${sha256}`, {
    method: "PUT", headers: { "Content-Type": "application/octet-stream" }, body: data,
  });
  if (!response.ok) throw await apiError("Upload file", response);
  return { sha256, bytes: data.byteLength };
}
export async function publishTree(base: ProjectVersion, entries: TreeEntry[], operationId: string): Promise<ProjectVersion> {
  const response = await authFetch(`${root(base.projectId)}/versions`, {
    method: "POST", body: JSON.stringify({ expectedVersion: base.id, operationId, entries }),
  });
  if (!response.ok) throw await apiError("Save project", response);
  return response.json() as Promise<ProjectVersion>;
}
export type VersionPage = { versions: ProjectVersion[]; next?: string };
export async function loadHistory(projectId: string, cursor = ""): Promise<VersionPage> {
  const response = await authFetch(`${root(projectId)}/versions?cursor=${encodeURIComponent(cursor)}`);
  if (!response.ok) throw await apiError("Load history", response);
  return response.json() as Promise<VersionPage>;
}
export async function restoreVersion(base: ProjectVersion, versionId: string, operationId: string): Promise<ProjectVersion> {
  const response = await authFetch(`${root(base.projectId)}/restore`, { method: "POST", body: JSON.stringify({ expectedVersion: base.id, versionId, operationId }) });
  if (!response.ok) throw await apiError("Restore version", response);
  return response.json() as Promise<ProjectVersion>;
}
export async function uploadBlob(projectId: string, file: Blob): Promise<{ sha256: string; bytes: number }> {
  if (file.size > 64 * 1024 * 1024) throw new Error("Files must be 64 MiB or smaller.");
  const data = await file.arrayBuffer();
  const digest = await crypto.subtle.digest("SHA-256", data);
  const sha256 = Array.from(new Uint8Array(digest), byte => byte.toString(16).padStart(2, "0")).join("");
  const response = await authFetch(`${root(projectId)}/objects/${sha256}`, { method: "PUT", headers: { "Content-Type": "application/octet-stream" }, body: file });
  if (!response.ok) throw await apiError("Upload file", response);
  return { sha256, bytes: file.size };
}

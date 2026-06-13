import type { WriterSkill } from "./skillTypes";

async function requestJSON<T>(path: string, init?: RequestInit): Promise<T> {
  const response = await fetch(path, init);
  if (!response.ok) {
    throw new Error(`${init?.method ?? "GET"} ${path} failed: ${response.status}`);
  }
  return response.json() as Promise<T>;
}

function projectSkillPath(projectId: string, suffix: string): string {
  return `/api/projects/${encodeURIComponent(projectId)}/${suffix}`;
}

export function listSkills(projectId: string, signal?: AbortSignal): Promise<WriterSkill[]> {
  return requestJSON<WriterSkill[]>(projectSkillPath(projectId, "skills"), { signal });
}

export function getSkill(projectId: string, skillId: string, signal?: AbortSignal): Promise<WriterSkill> {
  return requestJSON<WriterSkill>(projectSkillPath(projectId, `skills/${encodeURIComponent(skillId)}`), { signal });
}

export function importSkill(projectId: string, file: File): Promise<WriterSkill> {
  return requestJSON<WriterSkill>(projectSkillPath(projectId, "skills/import"), {
    method: "POST",
    body: file,
    headers: { "Content-Type": "application/zip" },
  });
}

export function listSessionSkills(projectId: string, sessionId: string, signal?: AbortSignal): Promise<WriterSkill[]> {
  return requestJSON<WriterSkill[]>(
    `/api/projects/${encodeURIComponent(projectId)}/agent/sessions/${encodeURIComponent(sessionId)}/skills`,
    { signal },
  );
}

export function selectSessionSkills(projectId: string, sessionId: string, skillIds: string[]): Promise<WriterSkill[]> {
  return requestJSON<WriterSkill[]>(
    `/api/projects/${encodeURIComponent(projectId)}/agent/sessions/${encodeURIComponent(sessionId)}/skills`,
    {
      method: "PUT",
      body: JSON.stringify({ skillIds }),
      headers: { "Content-Type": "application/json" },
    },
  );
}

import type { ScriptApproval, ScriptAudit, ScriptRun } from "./scriptRuntimeTypes";

async function requestJSON<T>(path: string, init?: RequestInit): Promise<T> {
  const response = await fetch(path, init);
  if (!response.ok) {
    let message = `${init?.method ?? "GET"} ${path} failed: ${response.status}`;
    try {
      const body = (await response.json()) as { error?: unknown };
      if (typeof body.error === "string" && body.error.trim()) {
        message = body.error;
      }
    } catch {
      // Keep the status-based fallback when the server returns a non-JSON error.
    }
    throw new Error(message);
  }
  return response.json() as Promise<T>;
}

function projectPath(projectId: string, suffix: string): string {
  return `/api/projects/${encodeURIComponent(projectId)}/${suffix}`;
}

export function listScriptAudits(projectId: string, skillId: string, signal?: AbortSignal): Promise<ScriptAudit[]> {
  return requestJSON<ScriptAudit[]>(projectPath(projectId, `skills/${encodeURIComponent(skillId)}/scripts`), { signal });
}

export function updateScriptApproval(
  projectId: string,
  skillId: string,
  skillFileId: string,
  body: {
    enabled: boolean;
    runtimeCommand: string;
    timeoutMs: number;
    maxStdoutBytes: number;
    maxStderrBytes: number;
    allowNetwork: boolean;
    allowProjectFiles: boolean;
  },
): Promise<ScriptApproval> {
  return requestJSON<ScriptApproval>(
    projectPath(projectId, `skills/${encodeURIComponent(skillId)}/scripts/${encodeURIComponent(skillFileId)}/approval`),
    {
      method: "PUT",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(body),
    },
  );
}

export function listProjectScriptRuns(projectId: string, signal?: AbortSignal): Promise<ScriptRun[]> {
  return requestJSON<ScriptRun[]>(projectPath(projectId, "skill-script-runs"), { signal });
}

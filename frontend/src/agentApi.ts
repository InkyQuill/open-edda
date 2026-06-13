import type {
  AcceptCandidateResult,
  ActivityEvent,
  AgentSession,
  ChatTurnResult,
  ContinuationRequest,
  ContinuationResult,
  CreateSessionRequest,
  GenerationCandidate,
  ModelVariant,
  ModelVariantRequest,
  PromptProfile,
  PromptProfileRequest,
  PromptRecord,
  ProviderConfigRequest,
  ProviderConfigSummary,
  ProviderConfigUpdateRequest,
  ReadAndCheckResult,
  RewriteRequest,
  RewriteResult,
} from "./agentTypes";

async function requestJSON<T>(path: string, init?: RequestInit): Promise<T> {
  const response = await fetch(path, {
    ...init,
    headers: {
      ...(init?.body ? { "Content-Type": "application/json" } : {}),
      ...init?.headers,
    },
  });
  if (!response.ok) {
    throw new Error(`${init?.method ?? "GET"} ${path} failed: ${response.status}`);
  }
  return response.json() as Promise<T>;
}

function projectAgentPath(projectId: string, suffix: string): string {
  return `/api/projects/${encodeURIComponent(projectId)}/agent/${suffix}`;
}

function withLimit(path: string, limit?: number): string {
  if (!limit) {
    return path;
  }
  const params = new URLSearchParams({ limit: String(limit) });
  return `${path}?${params}`;
}

export function listProviderConfigs(signal?: AbortSignal): Promise<ProviderConfigSummary[]> {
  return requestJSON<ProviderConfigSummary[]>("/api/provider-configs", { signal });
}

export function createProviderConfig(input: ProviderConfigRequest): Promise<ProviderConfigSummary> {
  return requestJSON<ProviderConfigSummary>("/api/provider-configs", {
    method: "POST",
    body: JSON.stringify(input),
  });
}

export function updateProviderConfig(providerId: string, input: ProviderConfigUpdateRequest): Promise<ProviderConfigSummary> {
  return requestJSON<ProviderConfigSummary>(`/api/provider-configs/${encodeURIComponent(providerId)}`, {
    method: "PUT",
    body: JSON.stringify(input),
  });
}

export function listModelVariants(providerId: string, signal?: AbortSignal): Promise<ModelVariant[]> {
  return requestJSON<ModelVariant[]>(`/api/provider-configs/${encodeURIComponent(providerId)}/model-variants`, { signal });
}

export function createModelVariant(providerId: string, input: ModelVariantRequest): Promise<ModelVariant> {
  return requestJSON<ModelVariant>(`/api/provider-configs/${encodeURIComponent(providerId)}/model-variants`, {
    method: "POST",
    body: JSON.stringify(input),
  });
}

export function getPromptProfile(projectId: string, signal?: AbortSignal): Promise<PromptProfile> {
  return requestJSON<PromptProfile>(projectAgentPath(projectId, "prompt-profile"), { signal });
}

export function upsertPromptProfile(projectId: string, input: PromptProfileRequest): Promise<PromptProfile> {
  return requestJSON<PromptProfile>(projectAgentPath(projectId, "prompt-profile"), {
    method: "PUT",
    body: JSON.stringify(input),
  });
}

export function listSessions(projectId: string, limit?: number, signal?: AbortSignal): Promise<AgentSession[]> {
  return requestJSON<AgentSession[]>(withLimit(projectAgentPath(projectId, "sessions"), limit), { signal });
}

export function createSession(
  projectId: string,
  input: CreateSessionRequest,
): Promise<AgentSession> {
  return requestJSON<AgentSession>(projectAgentPath(projectId, "sessions"), {
    method: "POST",
    body: JSON.stringify(input),
  });
}

export function createSessionMessage(projectId: string, sessionId: string, bodyMarkdown: string): Promise<ChatTurnResult> {
  return requestJSON<ChatTurnResult>(
    `${projectAgentPath(projectId, "sessions")}/${encodeURIComponent(sessionId)}/messages`,
    {
      method: "POST",
      body: JSON.stringify({ bodyMarkdown }),
    },
  );
}

export function runContinuation(projectId: string, input: ContinuationRequest): Promise<ContinuationResult> {
  return requestJSON<ContinuationResult>(projectAgentPath(projectId, "actions/continuation"), {
    method: "POST",
    body: JSON.stringify(input),
  });
}

export function runRewrite(projectId: string, input: RewriteRequest): Promise<RewriteResult> {
  return requestJSON<RewriteResult>(projectAgentPath(projectId, "actions/rewrite"), {
    method: "POST",
    body: JSON.stringify(input),
  });
}

export function runReadAndCheck(projectId: string, input: RewriteRequest): Promise<ReadAndCheckResult> {
  return requestJSON<ReadAndCheckResult>(projectAgentPath(projectId, "actions/read-check"), {
    method: "POST",
    body: JSON.stringify(input),
  });
}

export function acceptCandidate(projectId: string, candidateId: string): Promise<AcceptCandidateResult> {
  return requestJSON<AcceptCandidateResult>(
    `${projectAgentPath(projectId, "candidates")}/${encodeURIComponent(candidateId)}/accept`,
    { method: "POST", body: "{}" },
  );
}

export function rejectCandidate(projectId: string, candidateId: string): Promise<GenerationCandidate> {
  return requestJSON<GenerationCandidate>(
    `${projectAgentPath(projectId, "candidates")}/${encodeURIComponent(candidateId)}/reject`,
    { method: "POST", body: "{}" },
  );
}

export function listActivity(projectId: string, limit?: number, signal?: AbortSignal): Promise<ActivityEvent[]> {
  return requestJSON<ActivityEvent[]>(withLimit(projectAgentPath(projectId, "activity"), limit), { signal });
}

export function listPromptRecords(projectId: string, limit?: number, signal?: AbortSignal): Promise<PromptRecord[]> {
  return requestJSON<PromptRecord[]>(withLimit(projectAgentPath(projectId, "prompt-records"), limit), { signal });
}

export function prunePromptRecords(projectId: string): Promise<{ deleted: number }> {
  return requestJSON<{ deleted: number }>(projectAgentPath(projectId, "prompt-records/prune"), {
    method: "POST",
    body: "{}",
  });
}

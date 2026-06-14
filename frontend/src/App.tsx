import { useEffect, useMemo, useRef, useState } from "react";
import {
  acceptCandidate,
  createModelVariant,
  createProviderConfig,
  createSession,
  createSessionMessage,
  getPromptProfile,
  listActivity,
  listModelVariants,
  listPromptRecords,
  listProviderConfigs,
  listSessions,
  rejectCandidate,
  runContinuation,
  runReadAndCheck,
  runRewrite,
  updateProviderConfig,
  upsertPromptProfile,
} from "./agentApi";
import { getSkill, importSkill, listSessionSkills, listSkills, selectSessionSkills } from "./skillApi";
import { login, register, clearToken, getToken } from "./authApi";
import type {
  ActivityEvent,
  AgentMessage,
  AgentSession,
  ApplyMode,
  GenerationCandidate,
  ModelVariant,
  PromptProfileRequest,
  PromptRecord,
  ProviderConfigSummary,
} from "./agentTypes";
import { listContent, listProjects } from "./api";
import type { WriterSkill } from "./skillTypes";
import type { ContentItem, ContentKind, StoryProject } from "./types";
import "./styles.css";

const contentKinds: Array<{ kind: ContentKind; label: string }> = [
  { kind: "chapter", label: "Chapters" },
  { kind: "story_bible_entry", label: "Story Bible" },
  { kind: "writing_brief", label: "Briefs" },
  { kind: "project_note", label: "Notes" },
];

const emptyPromptProfile: PromptProfileRequest = {
  genre: "",
  tense: "",
  pov: "",
  voice: "",
  instructionsMarkdown: "",
  promptRecordRetentionDays: 30,
};

const emptyProviderForm = { name: "", baseUrl: "", apiKey: "" };

const emptyModelForm = {
  name: "",
  model: "",
  temperature: 0.4,
  maxOutputTokens: 4096,
  contextWindowTokens: 64000,
  inputPricePerMillion: 0,
  outputPricePerMillion: 0,
  cacheReadPricePerMillion: 0,
  cacheWritePricePerMillion: 0,
  requestTokenField: "max_tokens",
  reasoningFormat: "",
  compatibilityJson: "{}",
};

const textEncoder = new TextEncoder();

function byteLength(value: string): number {
  return textEncoder.encode(value).length;
}

function formatMoney(value: number): string {
  if (value <= 0) {
    return "$0.0000";
  }
  return `$${value.toFixed(value < 0.01 ? 6 : 4)}`;
}

function modelLabel(model: ModelVariant | null, providers: ProviderConfigSummary[]): string {
  if (!model) {
    return "No model selected";
  }
  const provider = providers.find((entry) => entry.id === model.providerConfigId);
  return `${provider?.name ?? "Provider"} / ${model.name} (${model.model})`;
}

function parseCompatibilityJson(raw: string): unknown {
  const trimmed = raw.trim();
  if (!trimmed) {
    return {};
  }
  return JSON.parse(trimmed) as unknown;
}

function updateContentInList(items: ContentItem[], updated: ContentItem): ContentItem[] {
  return items.map((item) => (item.id === updated.id ? updated : item));
}

function sortProvidersByName(items: ProviderConfigSummary[]): ProviderConfigSummary[] {
  return [...items].sort((a, b) => a.name.localeCompare(b.name));
}

function sortModelsByName(items: ModelVariant[]): ModelVariant[] {
  return [...items].sort((a, b) => a.name.localeCompare(b.name));
}

function sortSkillsByName(items: WriterSkill[]): WriterSkill[] {
  return [...items].sort((a, b) => a.displayName.localeCompare(b.displayName));
}

type DollarSkillToken = {
  start: number;
  end: number;
  query: string;
};

const skillMentionListboxId = "chat-skill-mention-listbox";
const tokenBoundaryPattern = /[\s,.;:!?()[\]{}"'`<>]/;

function isTokenBoundary(value: string): boolean {
  return tokenBoundaryPattern.test(value);
}

function skillMentionOptionId(skillId: string): string {
  return `${skillMentionListboxId}-option-${skillId}`;
}

function findDollarSkillToken(value: string, cursorPosition: number): DollarSkillToken | null {
  const cursor = Math.max(0, Math.min(cursorPosition, value.length));
  let start = cursor;
  while (start > 0 && !isTokenBoundary(value[start - 1])) {
    start -= 1;
  }

  let end = cursor;
  while (end < value.length && !isTokenBoundary(value[end])) {
    end += 1;
  }

  const token = value.slice(start, end);
  const prefix = value.slice(start, cursor);
  if (!token.startsWith("$") || token.includes("/") || !prefix.startsWith("$")) {
    return null;
  }

  return { start, end, query: prefix.slice(1).toLowerCase() };
}

function matchesSkillQuery(skill: WriterSkill, query: string): boolean {
  if (!query) {
    return true;
  }
  return [skill.name, skill.displayName, skill.description].some((value) => value.toLowerCase().includes(query));
}

function usageForSession(records: PromptRecord[], sessionId: string | null): { tokens: number; cost: number } {
  if (!sessionId) {
    return { tokens: 0, cost: 0 };
  }
  return records.reduce(
    (total, record) => {
      if (record.sessionId !== sessionId) {
        return total;
      }
      return {
        tokens: total.tokens + record.usage.totalTokens,
        cost: total.cost + record.usage.totalCost,
      };
    },
    { tokens: 0, cost: 0 },
  );
}

function newestRecordForSession(records: PromptRecord[], sessionId: string): PromptRecord | null {
  return records.find((record) => record.sessionId === sessionId) ?? null;
}

export function App() {
  const [projects, setProjects] = useState<StoryProject[]>([]);
  const [selectedProjectId, setSelectedProjectId] = useState<string | null>(null);
  const [selectedKind, setSelectedKind] = useState<ContentKind>("chapter");
  const [contentItems, setContentItems] = useState<ContentItem[]>([]);
  const [selectedContentId, setSelectedContentId] = useState<string | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [isContentLoading, setIsContentLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [isAuthenticated, setIsAuthenticated] = useState(() => getToken() !== null);
  const [authEmail, setAuthEmail] = useState("");
  const [authPassword, setAuthPassword] = useState("");
  const [authError, setAuthError] = useState<string | null>(null);
  const [isRegistering, setIsRegistering] = useState(false);
  const [isAuthLoading, setIsAuthLoading] = useState(false);
  const [contentError, setContentError] = useState<string | null>(null);

  const [providers, setProviders] = useState<ProviderConfigSummary[]>([]);
  const [models, setModels] = useState<ModelVariant[]>([]);
  const [selectedProviderId, setSelectedProviderId] = useState<string | null>(null);
  const [activeModelVariantId, setActiveModelVariantId] = useState<string | null>(null);
  const [promptProfile, setPromptProfile] = useState<PromptProfileRequest>(emptyPromptProfile);
  const [agentSessions, setAgentSessions] = useState<AgentSession[]>([]);
  const [activeChatSessionId, setActiveChatSessionId] = useState<string | null>(null);
  const [chatMessages, setChatMessages] = useState<AgentMessage[]>([]);
  const [activityEvents, setActivityEvents] = useState<ActivityEvent[]>([]);
  const [promptRecords, setPromptRecords] = useState<PromptRecord[]>([]);
  const [agentError, setAgentError] = useState<string | null>(null);
  const [skills, setSkills] = useState<WriterSkill[]>([]);
  const [skillError, setSkillError] = useState<string | null>(null);
  const [textSelection, setTextSelection] = useState<{ start: number; end: number } | null>(null);
  const latestProjectIdRef = useRef<string | null>(selectedProjectId);
  latestProjectIdRef.current = selectedProjectId;

  useEffect(() => {
    void listProjects()
      .then((items) => {
        setProjects(items);
        setSelectedProjectId((current) => current ?? items[0]?.id ?? null);
        setError(null);
      })
      .catch((cause: unknown) => {
        setError(cause instanceof Error ? cause.message : "Project list failed");
      })
      .finally(() => {
        setIsLoading(false);
      });
  }, []);

  useEffect(() => {
    if (!selectedProjectId) {
      setContentItems([]);
      setSelectedContentId(null);
      return;
    }

    const abortController = new AbortController();
    setIsContentLoading(true);
    setContentItems([]);
    setSelectedContentId(null);
    setContentError(null);
    void listContent(selectedProjectId, selectedKind, abortController.signal)
      .then((items) => {
        setContentItems(items);
        setSelectedContentId(items[0]?.id ?? null);
        setContentError(null);
      })
      .catch((cause: unknown) => {
        if (cause instanceof DOMException && cause.name === "AbortError") {
          return;
        }
        setContentItems([]);
        setSelectedContentId(null);
        setContentError(cause instanceof Error ? cause.message : "Content list failed");
      })
      .finally(() => {
        if (!abortController.signal.aborted) {
          setIsContentLoading(false);
        }
      });

    return () => {
      abortController.abort();
    };
  }, [selectedProjectId, selectedKind]);

  useEffect(() => {
    const abortController = new AbortController();
    void listProviderConfigs(abortController.signal)
      .then((items) => {
        setProviders(items);
        setSelectedProviderId((current) => current ?? items[0]?.id ?? null);
        setAgentError(null);
      })
      .catch((cause: unknown) => {
        if (cause instanceof DOMException && cause.name === "AbortError") {
          return;
        }
        setAgentError(cause instanceof Error ? cause.message : "Provider list failed");
      });
    return () => {
      abortController.abort();
    };
  }, []);

  const providerIds = useMemo(() => providers.map((provider) => provider.id).join("\n"), [providers]);

  useEffect(() => {
    const ids = providerIds ? providerIds.split("\n") : [];
    if (ids.length === 0) {
      setModels([]);
      setActiveModelVariantId(null);
      return;
    }

    const abortController = new AbortController();
    void Promise.all(ids.map((providerId) => listModelVariants(providerId, abortController.signal)))
      .then((groups) => {
        const nextModels = groups.flat();
        setModels(nextModels);
        setActiveModelVariantId((current) => current ?? nextModels[0]?.id ?? null);
      })
      .catch((cause: unknown) => {
        if (cause instanceof DOMException && cause.name === "AbortError") {
          return;
        }
        setAgentError(cause instanceof Error ? cause.message : "Model list failed");
      });
    return () => {
      abortController.abort();
    };
  }, [providerIds]);

  useEffect(() => {
    if (!selectedProjectId) {
      setPromptProfile(emptyPromptProfile);
      setAgentSessions([]);
      setActiveChatSessionId(null);
      setChatMessages([]);
      setActivityEvents([]);
      setPromptRecords([]);
      return;
    }

    const abortController = new AbortController();
    setChatMessages([]);
    setAgentError(null);
    void Promise.allSettled([
      getPromptProfile(selectedProjectId, abortController.signal),
      listSessions(selectedProjectId, 12, abortController.signal),
      listActivity(selectedProjectId, 30, abortController.signal),
      listPromptRecords(selectedProjectId, 30, abortController.signal),
    ]).then(([profileResult, sessionsResult, activityResult, recordsResult]) => {
      if (abortController.signal.aborted) {
        return;
      }
      setPromptProfile(profileResult.status === "fulfilled" ? profileResult.value : emptyPromptProfile);
      const sessions = sessionsResult.status === "fulfilled" ? sessionsResult.value : [];
      setAgentSessions(sessions);
      setActiveChatSessionId(sessions.find((session) => session.actionKind === "chat")?.id ?? null);
      setActivityEvents(activityResult.status === "fulfilled" ? activityResult.value : []);
      setPromptRecords(recordsResult.status === "fulfilled" ? recordsResult.value : []);
      if (sessionsResult.status === "rejected" || activityResult.status === "rejected" || recordsResult.status === "rejected") {
        setAgentError("Some agent data could not be loaded.");
      }
    });
    return () => {
      abortController.abort();
    };
  }, [selectedProjectId]);

  useEffect(() => {
    if (!selectedProjectId) {
      setSkills([]);
      setSkillError(null);
      return;
    }

    const abortController = new AbortController();
    setSkills([]);
    setSkillError(null);
    void listSkills(selectedProjectId, abortController.signal)
      .then((items) => {
        setSkills(sortSkillsByName(items));
        setSkillError(null);
      })
      .catch((cause: unknown) => {
        if (cause instanceof DOMException && cause.name === "AbortError") {
          return;
        }
        setSkills([]);
        setSkillError(cause instanceof Error ? cause.message : "Skill list failed");
      });

    return () => {
      abortController.abort();
    };
  }, [selectedProjectId]);

  const selectedProject = useMemo(
    () => projects.find((project) => project.id === selectedProjectId) ?? null,
    [projects, selectedProjectId],
  );
  const selectedContent = useMemo(
    () => contentItems.find((item) => item.id === selectedContentId) ?? null,
    [contentItems, selectedContentId],
  );
  const activeModel = useMemo(
    () => models.find((model) => model.id === activeModelVariantId) ?? null,
    [activeModelVariantId, models],
  );
  const selectedProviderModels = useMemo(
    () => models.filter((model) => model.providerConfigId === selectedProviderId),
    [models, selectedProviderId],
  );
  const currentModelLabel = modelLabel(activeModel, providers);
  const sessionSpend = useMemo(
    () => usageForSession(promptRecords, activeChatSessionId),
    [activeChatSessionId, promptRecords],
  );

  function refreshAgentTrail(projectId: string): void {
    void Promise.all([listActivity(projectId, 30), listPromptRecords(projectId, 30), listSessions(projectId, 12)])
      .then(([events, records, sessions]) => {
        if (latestProjectIdRef.current !== projectId) {
          return;
        }
        setActivityEvents(events);
        setPromptRecords(records);
        setAgentSessions(sessions);
      })
      .catch((cause: unknown) => {
        if (latestProjectIdRef.current !== projectId) {
          return;
        }
        setAgentError(cause instanceof Error ? cause.message : "Agent refresh failed");
      });
  }

  function refreshSkills(projectId: string): void {
    void listSkills(projectId)
      .then((items) => {
        if (latestProjectIdRef.current !== projectId) {
          return;
        }
        setSkills(sortSkillsByName(items));
        setSkillError(null);
      })
      .catch((cause: unknown) => {
        if (latestProjectIdRef.current !== projectId) {
          return;
        }
        setSkillError(cause instanceof Error ? cause.message : "Skill list failed");
      });
  }

  function setUpdatedContent(content: ContentItem): void {
    setContentItems((items) => updateContentInList(items, content));
    setSelectedContentId(content.id);
  }

  async function handleAuth(): Promise<void> {
    setAuthError(null);
    setIsAuthLoading(true);
    try {
      const fn = isRegistering ? register : login;
      await fn(authEmail, authPassword);
      setIsAuthenticated(true);
      setAuthPassword("");
      setAuthError(null);
    } catch (cause: unknown) {
      setAuthError(cause instanceof Error ? cause.message : "Authentication failed");
    } finally {
      setIsAuthLoading(false);
    }
  }

  if (!isAuthenticated) {
    return (
      <main className="app-shell">
        <section className="auth-form-section">
          <header>
            <h1>Writer</h1>
            <p>Self-hosted AI writing studio</p>
          </header>
          <form
            className="auth-form"
            onSubmit={(e) => {
              e.preventDefault();
              void handleAuth();
            }}
          >
            <label>
              Email
              <input
                type="email"
                value={authEmail}
                onChange={(e) => setAuthEmail(e.target.value)}
                required
                autoComplete="email"
              />
            </label>
            <label>
              Password (min 8 characters)
              <input
                type="password"
                value={authPassword}
                onChange={(e) => setAuthPassword(e.target.value)}
                required
                minLength={8}
                autoComplete={isRegistering ? "new-password" : "current-password"}
              />
            </label>
            {authError ? <p className="auth-error" role="alert">{authError}</p> : null}
            <button type="submit" disabled={isAuthLoading}>
              {isAuthLoading ? "Please wait..." : isRegistering ? "Register" : "Login"}
            </button>
            <button type="button" className="pill-button" onClick={() => { setAuthError(null); setIsRegistering((v) => !v); }}>
              {isRegistering ? "Already have an account? Login" : "New author? Register"}
            </button>
          </form>
        </section>
      </main>
    );
  }

  return (
    <main className="app-shell">
      <section className="project-dashboard">
        <header>
          <h1>Edda</h1>
          <p>self-hosted AI writing studio</p>
          <button
            className="pill-button"
            type="button"
            onClick={() => {
              clearToken();
              setIsAuthenticated(false);
              setSelectedProjectId(null);
            }}
          >
            Logout
          </button>
        </header>

        <div className="project-list">
          {isLoading ? (
            <p>Loading story projects...</p>
          ) : error ? (
            <p role="alert">Could not load story projects.</p>
          ) : projects.length === 0 ? (
            <p>No story projects yet.</p>
          ) : (
            projects.map((project) => (
              <button
                key={project.id}
                className="project-row"
                type="button"
                data-active={project.id === selectedProjectId}
                aria-pressed={project.id === selectedProjectId}
                onClick={() => setSelectedProjectId(project.id)}
              >
                <span className="project-row-title">{project.title}</span>
                <span className="project-row-meta">{project.language || "Language not set"}</span>
              </button>
            ))
          )}
        </div>
      </section>

      {selectedProject ? (
        <>
          <AgentSettings
            activeModelVariantId={activeModelVariantId}
            models={models}
            promptProfile={promptProfile}
            providers={providers}
            selectedProjectId={selectedProject.id}
            selectedProviderId={selectedProviderId}
            selectedProviderModels={selectedProviderModels}
            onActiveModelChange={setActiveModelVariantId}
            onModelsChange={setModels}
            onProfileChange={setPromptProfile}
            onProviderChange={setSelectedProviderId}
            onProvidersChange={setProviders}
            onError={setAgentError}
          />

          <SkillBrowser
            projectId={selectedProject.id}
            skills={skills}
            skillError={skillError}
            onError={setSkillError}
            onSkillsChange={setSkills}
            onRefreshSkills={refreshSkills}
          />

          <section className="workspace-shell" aria-label={`${selectedProject.title} workspace`}>
            <aside className="workspace-nav" aria-label="Content type">
              {contentKinds.map((entry) => (
                <button
                  key={entry.kind}
                  type="button"
                  data-active={entry.kind === selectedKind}
                  aria-pressed={entry.kind === selectedKind}
                  onClick={() => setSelectedKind(entry.kind)}
                >
                  {entry.label}
                </button>
              ))}
            </aside>

            <section className="content-list" aria-label="Project content">
              <header>
                <h2>{contentKinds.find((entry) => entry.kind === selectedKind)?.label}</h2>
                <p>{selectedProject.title}</p>
              </header>
              {isContentLoading ? (
                <p>Loading content...</p>
              ) : contentError ? (
                <p role="alert">Could not load content.</p>
              ) : contentItems.length === 0 ? (
                <p>No content in this section yet.</p>
              ) : (
                contentItems.map((item) => (
                  <button
                    key={item.id}
                    className="content-row"
                    type="button"
                    data-active={item.id === selectedContentId}
                    aria-pressed={item.id === selectedContentId}
                     onClick={() => { setSelectedContentId(item.id); setTextSelection(null); }}
                  >
                    <span>{item.title}</span>
                    <small>Revision {item.currentRevision}</small>
                  </button>
                ))
              )}
            </section>

            <section className="detail-panel" aria-label="Content detail">
              {selectedContent ? (
                <>
                  <header>
                    <h2>{selectedContent.title}</h2>
                    <p>{selectedContent.kind.replaceAll("_", " ")}</p>
                  </header>
                   <textarea
                     readOnly
                     value={selectedContent.bodyMarkdown}
                     aria-label={`${selectedContent.title} markdown`}
                     onSelect={(e) => {
                       const target = e.target as HTMLTextAreaElement;
                       if (target.selectionStart !== target.selectionEnd) {
                         setTextSelection({
                           start: byteLength(target.value.slice(0, target.selectionStart)),
                           end: byteLength(target.value.slice(0, target.selectionEnd)),
                         });
                       } else {
                         setTextSelection(null);
                       }
                     }}
                   />
                </>
              ) : (
                <p>Select an item to preview its Markdown.</p>
              )}
            </section>

            <AgentPanel
              activeSessionId={activeChatSessionId}
              activeModel={activeModel}
              activityEvents={activityEvents}
              agentError={agentError}
              chatMessages={chatMessages}
              currentModelLabel={currentModelLabel}
              hasConfiguredModelVariant={models.length > 0}
              promptRecords={promptRecords}
              projectId={selectedProject.id}
              selectedContent={selectedContent}
              skills={skills}
              sessionSpend={sessionSpend}
              sessions={agentSessions}
              onActiveSessionChange={setActiveChatSessionId}
              onChatMessagesChange={setChatMessages}
              onContentUpdate={setUpdatedContent}
              onError={setAgentError}
              onRefreshTrail={refreshAgentTrail}
              onSessionsChange={setAgentSessions}
              textSelection={textSelection}
              onTextSelectionClear={() => setTextSelection(null)}
            />
          </section>
        </>
      ) : null}
    </main>
  );
}

type AgentSettingsProps = {
  activeModelVariantId: string | null;
  models: ModelVariant[];
  promptProfile: PromptProfileRequest;
  providers: ProviderConfigSummary[];
  selectedProjectId: string;
  selectedProviderId: string | null;
  selectedProviderModels: ModelVariant[];
  onActiveModelChange: (modelId: string | null) => void;
  onModelsChange: (models: ModelVariant[]) => void;
  onProfileChange: (profile: PromptProfileRequest) => void;
  onProviderChange: (providerId: string | null) => void;
  onProvidersChange: (providers: ProviderConfigSummary[]) => void;
  onError: (message: string | null) => void;
};

type SkillBrowserProps = {
  projectId: string;
  skills: WriterSkill[];
  skillError: string | null;
  onError: (message: string | null) => void;
  onSkillsChange: React.Dispatch<React.SetStateAction<WriterSkill[]>>;
  onRefreshSkills: (projectId: string) => void;
};

function SkillBrowser({ projectId, skills, skillError, onError, onSkillsChange, onRefreshSkills }: SkillBrowserProps) {
  const [selectedSkillId, setSelectedSkillId] = useState<string | null>(null);
  const [skillDetail, setSkillDetail] = useState<WriterSkill | null>(null);
  const [isImporting, setIsImporting] = useState(false);
  const fileInputRef = useRef<HTMLInputElement | null>(null);
  const currentProjectIdRef = useRef(projectId);
  const importRequestRef = useRef(0);
  currentProjectIdRef.current = projectId;

  useEffect(() => {
    currentProjectIdRef.current = projectId;
    importRequestRef.current += 1;
    setIsImporting(false);
  }, [projectId]);

  useEffect(() => {
    setSelectedSkillId((current) => {
      if (current && skills.some((skill) => skill.id === current)) {
        return current;
      }
      return skills[0]?.id ?? null;
    });
  }, [skills]);

  useEffect(() => {
    if (!selectedSkillId) {
      setSkillDetail(null);
      return;
    }

    const abortController = new AbortController();
    setSkillDetail(null);
    void getSkill(projectId, selectedSkillId, abortController.signal)
      .then((skill) => {
        setSkillDetail(skill);
        onError(null);
      })
      .catch((cause: unknown) => {
        if (cause instanceof DOMException && cause.name === "AbortError") {
          return;
        }
        setSkillDetail(null);
        onError(cause instanceof Error ? cause.message : "Skill detail failed");
      });

    return () => {
      abortController.abort();
    };
  }, [projectId, selectedSkillId, onError]);

  function handleImport(event: React.ChangeEvent<HTMLInputElement>): void {
    const file = event.target.files?.[0] ?? null;
    event.target.value = "";
    if (!file) {
      return;
    }
    const importProjectId = projectId;
    const requestId = importRequestRef.current + 1;
    importRequestRef.current = requestId;
    setIsImporting(true);
    void importSkill(importProjectId, file)
      .then((skill) => {
        if (currentProjectIdRef.current !== importProjectId || importRequestRef.current !== requestId) {
          return;
        }
        onSkillsChange((current) => sortSkillsByName([skill, ...current.filter((entry) => entry.id !== skill.id)]));
        setSelectedSkillId(skill.id);
        onError(null);
        onRefreshSkills(importProjectId);
      })
      .catch((cause: unknown) => {
        if (currentProjectIdRef.current !== importProjectId || importRequestRef.current !== requestId) {
          return;
        }
        onError(cause instanceof Error ? cause.message : "Skill import failed");
      })
      .finally(() => {
        if (currentProjectIdRef.current === importProjectId && importRequestRef.current === requestId) {
          setIsImporting(false);
        }
      });
  }

  return (
    <section className="skill-browser" aria-label="Skills">
      <header className="skill-browser-header">
        <div>
          <h2>Skills</h2>
          <p>{skills.length} installed</p>
        </div>
        <button className="pill-button" type="button" onClick={() => fileInputRef.current?.click()} disabled={isImporting}>
          {isImporting ? "Importing..." : "Import .zip"}
        </button>
        <input
          ref={fileInputRef}
          className="visually-hidden"
          type="file"
          accept=".zip,application/zip"
          onChange={handleImport}
          disabled={isImporting}
          tabIndex={-1}
        />
      </header>

      {skillError ? <p className="agent-error" role="alert">{skillError}</p> : null}

      <div className="skill-browser-grid">
        <div className="skill-list">
          {skills.length === 0 ? (
            <p className="muted">No skills installed.</p>
          ) : (
            skills.map((skill) => (
              <button
                key={skill.id}
                className="skill-row"
                type="button"
                data-active={skill.id === selectedSkillId}
                onClick={() => setSelectedSkillId(skill.id)}
              >
                <span className="skill-row-title">{skill.displayName}</span>
                <span className="skill-row-description">{skill.description}</span>
                <SkillChips skill={skill} />
              </button>
            ))
          )}
        </div>

        <SkillDetail skill={skillDetail} />
      </div>
    </section>
  );
}

function SkillChips({ skill, showScriptStatus = true }: { skill: WriterSkill; showScriptStatus?: boolean }) {
  const hints = skill.routingHints ?? [];
  return (
    <span className="skill-chip-row">
      {hints.slice(0, 4).map((hint, index) => (
        <span key={`${hint.actionKind}-${hint.contentKind}-${hint.tag}-${index}`} className="skill-badge">
          {[hint.actionKind, hint.contentKind, hint.tag].filter(Boolean).join(" / ")}
        </span>
      ))}
      {hints.length > 4 ? <span className="skill-badge">+{hints.length - 4}</span> : null}
      {showScriptStatus && skill.scriptCount > 0 ? <span className="skill-badge skill-badge-warning">Scripts disabled</span> : null}
    </span>
  );
}

function SkillDetail({ skill }: { skill: WriterSkill | null }) {
  if (!skill) {
    return (
      <section className="skill-detail" aria-label="Skill detail">
        <p className="muted">No skill selected.</p>
      </section>
    );
  }

  return (
    <section className="skill-detail" aria-label={`${skill.displayName} detail`}>
      <header>
        <div>
          <h3>{skill.displayName}</h3>
          <p>{skill.name}</p>
        </div>
        {skill.scriptCount > 0 ? <span className="skill-badge skill-badge-warning">Scripts disabled</span> : null}
      </header>
      {(skill.routingHints ?? []).length > 0 ? <SkillChips skill={skill} showScriptStatus={false} /> : null}
      <div className="skill-detail-body">
        <section>
          <h4>Instructions</h4>
          <pre>{skill.instructionsMarkdown || "No instructions."}</pre>
        </section>
        <section>
          <h4>Files</h4>
          <div className="skill-file-list">
            {(skill.files ?? []).length === 0 ? (
              <p className="muted">No files.</p>
            ) : (
              (skill.files ?? []).map((file) => (
                <div key={file.id} className="skill-file-row">
                  <div>
                    <strong>{file.relativePath}</strong>
                    <span>
                      {file.purpose} · {file.bytes} bytes
                    </span>
                  </div>
                  {file.scriptDisabled ? <span className="skill-badge skill-badge-warning">Script disabled</span> : null}
                </div>
              ))
            )}
          </div>
        </section>
      </div>
    </section>
  );
}

function AgentSettings({
  activeModelVariantId,
  models,
  promptProfile,
  providers,
  selectedProjectId,
  selectedProviderId,
  selectedProviderModels,
  onActiveModelChange,
  onModelsChange,
  onProfileChange,
  onProviderChange,
  onProvidersChange,
  onError,
}: AgentSettingsProps) {
  const [providerMode, setProviderMode] = useState<"create" | "update">("create");
  const [providerForm, setProviderForm] = useState(emptyProviderForm);
  const [modelForm, setModelForm] = useState(emptyModelForm);
  const [isSavingProvider, setIsSavingProvider] = useState(false);
  const [isSavingModel, setIsSavingModel] = useState(false);
  const [isSavingProfile, setIsSavingProfile] = useState(false);

  const selectedProvider = providers.find((provider) => provider.id === selectedProviderId) ?? null;

  useEffect(() => {
    if (providerMode === "update" && selectedProvider) {
      setProviderForm({ name: selectedProvider.name, baseUrl: selectedProvider.baseUrl, apiKey: "" });
    }
    if (providerMode === "create") {
      setProviderForm(emptyProviderForm);
    }
  }, [providerMode, selectedProvider?.id, selectedProvider?.baseUrl, selectedProvider?.name]);

  function handleProviderSubmit(event: React.FormEvent<HTMLFormElement>): void {
    event.preventDefault();
    setIsSavingProvider(true);
    const action =
      providerMode === "update" && selectedProvider
        ? updateProviderConfig(selectedProvider.id, { baseUrl: providerForm.baseUrl, apiKey: providerForm.apiKey })
        : createProviderConfig(providerForm);

    void action
      .then((provider) => {
        const nextProviders =
          providerMode === "update"
            ? providers.map((entry) => (entry.id === provider.id ? provider : entry))
            : sortProvidersByName([...providers, provider]);
        onProvidersChange(nextProviders);
        onProviderChange(provider.id);
        setProviderMode("update");
        setProviderForm({ name: provider.name, baseUrl: provider.baseUrl, apiKey: "" });
        onError(null);
      })
      .catch((cause: unknown) => {
        onError(cause instanceof Error ? cause.message : "Provider save failed");
      })
      .finally(() => setIsSavingProvider(false));
  }

  function handleModelSubmit(event: React.FormEvent<HTMLFormElement>): void {
    event.preventDefault();
    if (!selectedProviderId) {
      onError("Select or create a provider before adding a model.");
      return;
    }
    setIsSavingModel(true);
    let compatibilityJson: unknown;
    try {
      compatibilityJson = parseCompatibilityJson(modelForm.compatibilityJson);
    } catch {
      onError("Compatibility JSON is invalid.");
      setIsSavingModel(false);
      return;
    }
    void createModelVariant(selectedProviderId, { ...modelForm, compatibilityJson })
      .then((model) => {
        onModelsChange(sortModelsByName([...models, model]));
        onActiveModelChange(model.id);
        setModelForm(emptyModelForm);
        onError(null);
      })
      .catch((cause: unknown) => {
        onError(cause instanceof Error ? cause.message : "Model save failed");
      })
      .finally(() => setIsSavingModel(false));
  }

  function handleProfileSubmit(event: React.FormEvent<HTMLFormElement>): void {
    event.preventDefault();
    setIsSavingProfile(true);
    void upsertPromptProfile(selectedProjectId, promptProfile)
      .then((profile) => {
        onProfileChange({
          genre: profile.genre,
          tense: profile.tense,
          pov: profile.pov,
          voice: profile.voice,
          instructionsMarkdown: profile.instructionsMarkdown,
          promptRecordRetentionDays: profile.promptRecordRetentionDays,
        });
        onError(null);
      })
      .catch((cause: unknown) => {
        onError(cause instanceof Error ? cause.message : "Prompt profile save failed");
      })
      .finally(() => setIsSavingProfile(false));
  }

  return (
    <section className="agent-settings" aria-label="Agent settings">
      <header className="agent-settings-header">
        <div>
          <h2>Agent settings</h2>
          <p>{models.length === 0 ? "No configured model variant." : `${models.length} model variant available.`}</p>
        </div>
        <label className="field inline-field">
          <span>Active model</span>
          <select
            value={activeModelVariantId ?? ""}
            onChange={(event) => onActiveModelChange(event.target.value || null)}
            disabled={models.length === 0}
          >
            <option value="">No model</option>
            {models.map((model) => (
              <option key={model.id} value={model.id}>
                {model.name} / {model.model}
              </option>
            ))}
          </select>
        </label>
      </header>

      <div className="settings-grid">
        <form className="settings-card" onSubmit={handleProviderSubmit}>
          <header>
            <h3>Provider</h3>
            <select value={providerMode} onChange={(event) => setProviderMode(event.target.value as "create" | "update")}>
              <option value="create">Create</option>
              <option value="update" disabled={!selectedProvider}>
                Update selected
              </option>
            </select>
          </header>
          <label className="field">
            <span>Provider config</span>
            <select value={selectedProviderId ?? ""} onChange={(event) => onProviderChange(event.target.value || null)}>
              <option value="">Select provider</option>
              {providers.map((provider) => (
                <option key={provider.id} value={provider.id}>
                  {provider.name}
                </option>
              ))}
            </select>
          </label>
          <div className="field-row">
            <label className="field">
              <span>Name</span>
              <input
                value={providerForm.name}
                onChange={(event) => setProviderForm((form) => ({ ...form, name: event.target.value }))}
                disabled={providerMode === "update"}
                required={providerMode === "create"}
              />
            </label>
            <label className="field">
              <span>Base URL</span>
              <input
                value={providerForm.baseUrl}
                onChange={(event) => setProviderForm((form) => ({ ...form, baseUrl: event.target.value }))}
                placeholder="https://api.provider.test"
                required
              />
            </label>
          </div>
          <label className="field">
            <span>API key</span>
            <input
              type="password"
              value={providerForm.apiKey}
              onChange={(event) => setProviderForm((form) => ({ ...form, apiKey: event.target.value }))}
              placeholder="Stored after save; never displayed"
              required
            />
          </label>
          <button type="submit" disabled={isSavingProvider}>
            {isSavingProvider ? "Saving..." : "Save provider"}
          </button>
        </form>

        <form className="settings-card" onSubmit={handleModelSubmit}>
          <header>
            <h3>Model variant</h3>
            <span>{selectedProviderModels.length} listed</span>
          </header>
          <div className="model-list">
            {selectedProviderModels.length === 0 ? (
              <p>No variants for this provider.</p>
            ) : (
              selectedProviderModels.map((model) => (
                <button
                  key={model.id}
                  type="button"
                  data-active={model.id === activeModelVariantId}
                  onClick={() => onActiveModelChange(model.id)}
                >
                  <span>{model.name}</span>
                  <small>
                    {model.model} · in {model.inputPricePerMillion}/out {model.outputPricePerMillion}
                  </small>
                </button>
              ))
            )}
          </div>
          <div className="field-row">
            <label className="field">
              <span>Name</span>
              <input value={modelForm.name} onChange={(event) => setModelForm((form) => ({ ...form, name: event.target.value }))} required />
            </label>
            <label className="field">
              <span>Model</span>
              <input value={modelForm.model} onChange={(event) => setModelForm((form) => ({ ...form, model: event.target.value }))} required />
            </label>
          </div>
          <div className="field-row price-row">
            <NumberField label="Input / 1M" value={modelForm.inputPricePerMillion} onChange={(value) => setModelForm((form) => ({ ...form, inputPricePerMillion: value }))} />
            <NumberField label="Output / 1M" value={modelForm.outputPricePerMillion} onChange={(value) => setModelForm((form) => ({ ...form, outputPricePerMillion: value }))} />
            <NumberField label="Cache read / 1M" value={modelForm.cacheReadPricePerMillion} onChange={(value) => setModelForm((form) => ({ ...form, cacheReadPricePerMillion: value }))} />
            <NumberField label="Cache write / 1M" value={modelForm.cacheWritePricePerMillion} onChange={(value) => setModelForm((form) => ({ ...form, cacheWritePricePerMillion: value }))} />
          </div>
          <div className="field-row">
            <NumberField label="Temperature" value={modelForm.temperature} step={0.1} onChange={(value) => setModelForm((form) => ({ ...form, temperature: value }))} />
            <NumberField label="Max output" value={modelForm.maxOutputTokens} step={1} onChange={(value) => setModelForm((form) => ({ ...form, maxOutputTokens: value }))} />
            <NumberField label="Context" value={modelForm.contextWindowTokens} step={1} onChange={(value) => setModelForm((form) => ({ ...form, contextWindowTokens: value }))} />
          </div>
          <div className="field-row">
            <label className="field">
              <span>Token field</span>
              <input value={modelForm.requestTokenField} onChange={(event) => setModelForm((form) => ({ ...form, requestTokenField: event.target.value }))} />
            </label>
            <label className="field">
              <span>Reasoning format</span>
              <input value={modelForm.reasoningFormat} onChange={(event) => setModelForm((form) => ({ ...form, reasoningFormat: event.target.value }))} />
            </label>
          </div>
          <label className="field">
            <span>Compatibility JSON</span>
            <input value={modelForm.compatibilityJson} onChange={(event) => setModelForm((form) => ({ ...form, compatibilityJson: event.target.value }))} />
          </label>
          <button type="submit" disabled={isSavingModel || !selectedProviderId}>
            {isSavingModel ? "Adding..." : "Add model variant"}
          </button>
        </form>

        <form className="settings-card prompt-card" onSubmit={handleProfileSubmit}>
          <header>
            <h3>Prompt profile</h3>
            <span>{promptProfile.promptRecordRetentionDays} day records</span>
          </header>
          <div className="field-row">
            <label className="field">
              <span>Genre</span>
              <input value={promptProfile.genre} onChange={(event) => onProfileChange({ ...promptProfile, genre: event.target.value })} />
            </label>
            <label className="field">
              <span>Tense</span>
              <input value={promptProfile.tense} onChange={(event) => onProfileChange({ ...promptProfile, tense: event.target.value })} />
            </label>
            <label className="field">
              <span>Point of view</span>
              <input value={promptProfile.pov} onChange={(event) => onProfileChange({ ...promptProfile, pov: event.target.value })} />
            </label>
            <label className="field">
              <span>Voice</span>
              <input value={promptProfile.voice} onChange={(event) => onProfileChange({ ...promptProfile, voice: event.target.value })} />
            </label>
          </div>
          <label className="field">
            <span>Writing instructions</span>
            <textarea
              value={promptProfile.instructionsMarkdown}
              onChange={(event) => onProfileChange({ ...promptProfile, instructionsMarkdown: event.target.value })}
            />
          </label>
          <NumberField
            label="Prompt Record retention days"
            value={promptProfile.promptRecordRetentionDays}
            step={1}
            onChange={(value) => onProfileChange({ ...promptProfile, promptRecordRetentionDays: value })}
          />
          <button type="submit" disabled={isSavingProfile}>
            {isSavingProfile ? "Saving..." : "Save prompt profile"}
          </button>
        </form>
      </div>
    </section>
  );
}

type NumberFieldProps = {
  label: string;
  value: number;
  onChange: (value: number) => void;
  step?: number;
};

function NumberField({ label, value, onChange, step = 0.01 }: NumberFieldProps) {
  return (
    <label className="field">
      <span>{label}</span>
      <input type="number" min="0" step={step} value={value} onChange={(event) => onChange(Number(event.target.value))} />
    </label>
  );
}

type AgentPanelProps = {
  activeSessionId: string | null;
  activeModel: ModelVariant | null;
  activityEvents: ActivityEvent[];
  agentError: string | null;
  chatMessages: AgentMessage[];
  currentModelLabel: string;
  hasConfiguredModelVariant: boolean;
  promptRecords: PromptRecord[];
  projectId: string;
  selectedContent: ContentItem | null;
  skills: WriterSkill[];
  sessionSpend: { tokens: number; cost: number };
  sessions: AgentSession[];
  onActiveSessionChange: (sessionId: string | null) => void;
  onChatMessagesChange: React.Dispatch<React.SetStateAction<AgentMessage[]>>;
  onContentUpdate: (content: ContentItem) => void;
  onError: (message: string | null) => void;
  onRefreshTrail: (projectId: string) => void;
  onSessionsChange: React.Dispatch<React.SetStateAction<AgentSession[]>>;
  textSelection: { start: number; end: number } | null;
  onTextSelectionClear: () => void;
};

type AgentPanelScope = {
  projectId: string;
  contentId: string | null;
  modelVariantId: string | null;
};

function AgentPanel({
  activeSessionId,
  activeModel,
  activityEvents,
  agentError,
  chatMessages,
  currentModelLabel,
  hasConfiguredModelVariant,
  promptRecords,
  projectId,
  selectedContent,
  skills,
  sessionSpend,
  sessions,
  onActiveSessionChange,
  onChatMessagesChange,
  onContentUpdate,
  onError,
  onRefreshTrail,
  onSessionsChange,
  textSelection,
  onTextSelectionClear,
}: AgentPanelProps) {
  const [applyMode, setApplyMode] = useState<ApplyMode>("preview");
  const [messageInput, setMessageInput] = useState("");
  const [isSendingMessage, setIsSendingMessage] = useState(false);
  const [isRunningAction, setIsRunningAction] = useState(false);
  const [showActivity, setShowActivity] = useState(false);
  const [continuationUnits, setContinuationUnits] = useState<"word" | "sentence">("word");
  const [continuationCount, setContinuationCount] = useState(120);
  const [continuationInsert, setContinuationInsert] = useState(false);
  const [guidance, setGuidance] = useState("");
  const [previewCandidate, setPreviewCandidate] = useState<GenerationCandidate | null>(null);
  const [readCheckReport, setReadCheckReport] = useState<string | null>(null);
  const [selectedSkillIds, setSelectedSkillIds] = useState<string[]>([]);
  const [skillSelectionModelId, setSkillSelectionModelId] = useState<string | null>(null);
  const [isSkillSelectionLoading, setIsSkillSelectionLoading] = useState(false);
  const [chatCursorPosition, setChatCursorPosition] = useState(0);
  const [highlightedSkillIndex, setHighlightedSkillIndex] = useState(0);
  const [dismissedSkillMentionKey, setDismissedSkillMentionKey] = useState<string | null>(null);
  const chatTextareaRef = useRef<HTMLTextAreaElement | null>(null);
  const skillSelectionLoadRef = useRef(0);
  const skillSelectionDirtyRef = useRef(false);

  const selectedContentId = selectedContent?.id ?? null;
  const activeModelId = activeModel?.id ?? null;
  const latestScopeRef = useRef<AgentPanelScope>({ projectId, contentId: selectedContentId, modelVariantId: activeModelId });
  latestScopeRef.current = { projectId, contentId: selectedContentId, modelVariantId: activeModelId };

  useEffect(() => {
    setPreviewCandidate(null);
    setReadCheckReport(null);
    setIsSendingMessage(false);
    setIsRunningAction(false);
  }, [projectId, selectedContentId, activeModelId]);

  const activeChatSession = sessions.find((session) => session.actionKind === "chat" && session.modelVariantId === activeModel?.id) ?? null;
  const activeChatSessionId = activeChatSession?.id ?? null;
  const scopedSelectedSkillIds = skillSelectionModelId === activeModelId ? selectedSkillIds : [];
  const actionsDisabled = !activeModel || !selectedContent || selectedContent.kind !== "chapter" || isRunningAction;
  const activeSessionRecord = activeSessionId ? newestRecordForSession(promptRecords, activeSessionId) : null;
  const activeDollarSkillToken = useMemo(
    () => findDollarSkillToken(messageInput, chatCursorPosition),
    [chatCursorPosition, messageInput],
  );
  const activeDollarSkillTokenKey = activeDollarSkillToken
    ? `${activeDollarSkillToken.start}:${activeDollarSkillToken.end}:${activeDollarSkillToken.query}`
    : null;
  const skillMentionSuggestions = useMemo(() => {
    if (!activeDollarSkillToken) {
      return [];
    }
    return skills.filter((skill) => matchesSkillQuery(skill, activeDollarSkillToken.query)).slice(0, 6);
  }, [activeDollarSkillToken, skills]);
  const showSkillMentionSuggestions =
    skills.length > 0 &&
    activeDollarSkillToken !== null &&
    activeDollarSkillTokenKey !== dismissedSkillMentionKey &&
    skillMentionSuggestions.length > 0;
  const visibleHighlightedSkillIndex =
    showSkillMentionSuggestions && highlightedSkillIndex < skillMentionSuggestions.length ? highlightedSkillIndex : 0;
  const highlightedSkill = showSkillMentionSuggestions ? skillMentionSuggestions[visibleHighlightedSkillIndex] : null;

  useEffect(() => {
    setHighlightedSkillIndex(0);
  }, [activeDollarSkillTokenKey, skillMentionSuggestions.length]);

  useEffect(() => {
    const installedSkillIds = new Set(skills.map((skill) => skill.id));
    setSelectedSkillIds((current) => current.filter((skillId) => installedSkillIds.has(skillId)));
  }, [skills]);

  useEffect(() => {
    skillSelectionDirtyRef.current = false;
    const loadId = skillSelectionLoadRef.current + 1;
    skillSelectionLoadRef.current = loadId;
    setSkillSelectionModelId(null);
    setSelectedSkillIds([]);

    if (!activeChatSessionId) {
      setSkillSelectionModelId(activeModelId);
      setIsSkillSelectionLoading(false);
      return;
    }

    const abortController = new AbortController();
    setIsSkillSelectionLoading(true);
    void listSessionSkills(projectId, activeChatSessionId, abortController.signal)
      .then((sessionSkills) => {
        if (skillSelectionLoadRef.current !== loadId || skillSelectionDirtyRef.current) {
          return;
        }
        setSkillSelectionModelId(activeModelId);
        setSelectedSkillIds(sessionSkills.map((skill) => skill.id));
        onError(null);
      })
      .catch((cause: unknown) => {
        if (cause instanceof DOMException && cause.name === "AbortError") {
          return;
        }
        onError(cause instanceof Error ? cause.message : "Session skill list failed");
      })
      .finally(() => {
        if (skillSelectionLoadRef.current === loadId) {
          setIsSkillSelectionLoading(false);
        }
      });

    return () => {
      abortController.abort();
    };
  }, [projectId, activeChatSessionId, activeModelId, onError]);

  function currentScope(): AgentPanelScope {
    return { projectId, contentId: selectedContentId, modelVariantId: activeModelId };
  }

  function isCurrentScope(scope: AgentPanelScope): boolean {
    const latest = latestScopeRef.current;
    return (
      latest.projectId === scope.projectId &&
      latest.contentId === scope.contentId &&
      latest.modelVariantId === scope.modelVariantId
    );
  }

  function toggleSkill(skillId: string, checked: boolean): void {
    skillSelectionDirtyRef.current = true;
    setSkillSelectionModelId(activeModelId);
    setSelectedSkillIds((current) => {
      if (checked) {
        return current.includes(skillId) ? current : [...current, skillId];
      }
      return current.filter((entry) => entry !== skillId);
    });
  }

  function updateChatCursorPosition(element: HTMLTextAreaElement): void {
    setChatCursorPosition(element.selectionStart);
  }

  function acceptSkillMention(skill: WriterSkill): void {
    if (!activeDollarSkillToken) {
      return;
    }
    const beforeToken = messageInput.slice(0, activeDollarSkillToken.start);
    const afterToken = messageInput.slice(activeDollarSkillToken.end);
    const suffix = afterToken.length === 0 || isTokenBoundary(afterToken[0]) ? "" : " ";
    const mention = `$${skill.name}`;
    const nextInput = `${beforeToken}${mention}${suffix}${afterToken}`;
    const nextCursorPosition = beforeToken.length + mention.length + suffix.length;

    toggleSkill(skill.id, true);
    setMessageInput(nextInput);
    setChatCursorPosition(nextCursorPosition);
    setDismissedSkillMentionKey(null);
    window.requestAnimationFrame(() => {
      chatTextareaRef.current?.focus();
      chatTextareaRef.current?.setSelectionRange(nextCursorPosition, nextCursorPosition);
    });
  }

  function handleChatKeyDown(event: React.KeyboardEvent<HTMLTextAreaElement>): void {
    if (!showSkillMentionSuggestions) {
      return;
    }

    if (event.key === "ArrowDown") {
      event.preventDefault();
      setHighlightedSkillIndex((current) => (current + 1) % skillMentionSuggestions.length);
      return;
    }
    if (event.key === "ArrowUp") {
      event.preventDefault();
      setHighlightedSkillIndex((current) => (current - 1 + skillMentionSuggestions.length) % skillMentionSuggestions.length);
      return;
    }
    if (event.key === "Escape") {
      event.preventDefault();
      setDismissedSkillMentionKey(activeDollarSkillTokenKey);
      return;
    }
    if (event.key === "Enter" || event.key === "Tab") {
      event.preventDefault();
      if (highlightedSkill) {
        acceptSkillMention(highlightedSkill);
      }
    }
  }

  async function ensureChatSession(scope: AgentPanelScope): Promise<AgentSession | null> {
    if (activeChatSession) {
      if (isCurrentScope(scope)) {
        onActiveSessionChange(activeChatSession.id);
      }
      return activeChatSession;
    }
    if (!activeModel) {
      throw new Error("Select a model variant before chatting.");
    }
    const session = await createSession(scope.projectId, {
      title: selectedContent ? `Chat: ${selectedContent.title}` : "Workspace chat",
      actionKind: "chat",
      modelVariantId: activeModel.id,
      applyMode,
      skillIds: scopedSelectedSkillIds,
    });
    if (!isCurrentScope(scope)) {
      return null;
    }
    onSessionsChange((current) => [session, ...current.filter((entry) => entry.id !== session.id)]);
    onActiveSessionChange(session.id);
    return session;
  }

  function handleChatSubmit(event: React.FormEvent<HTMLFormElement>): void {
    event.preventDefault();
    const body = messageInput.trim();
    if (!body) {
      return;
    }
    const launchScope = currentScope();
    setIsSendingMessage(true);
    void ensureChatSession(launchScope)
      .then((session) => {
        if (!session || !isCurrentScope(launchScope)) {
          return null;
        }
        return selectSessionSkills(launchScope.projectId, session.id, scopedSelectedSkillIds).then(() => session);
      })
      .then((session) => {
        if (!session || !isCurrentScope(launchScope)) {
          return null;
        }
        return createSessionMessage(launchScope.projectId, session.id, body);
      })
      .then((result) => {
        if (!result || !isCurrentScope(launchScope)) {
          return;
        }
        onChatMessagesChange((current) => [...current, result.userMessage, result.assistantMessage]);
        setMessageInput("");
        onError(null);
        onRefreshTrail(launchScope.projectId);
      })
      .catch((cause: unknown) => {
        if (!isCurrentScope(launchScope)) {
          return;
        }
        onError(cause instanceof Error ? cause.message : "Chat turn failed");
      })
      .finally(() => {
        if (isCurrentScope(launchScope)) {
          setIsSendingMessage(false);
        }
      });
  }

  function handleQuickAction(action: "continuation" | "rewrite" | "read_check"): void {
    if (!activeModel || !selectedContent) {
      onError("Select a chapter and model variant before running an action.");
      return;
    }
    const launchScope = currentScope();
    const expectedRevision = selectedContent.currentRevision;
    const wholeSelectionEnd = byteLength(selectedContent.bodyMarkdown);
    setIsRunningAction(true);
    setPreviewCandidate(null);
    setReadCheckReport(null);

    const actionRequest =
      action === "continuation"
        ? runContinuation(launchScope.projectId, {
            contentId: selectedContent.id,
            modelVariantId: activeModel.id,
            applyMode,
            guidance,
            expectedRevision,
            insertPosition: wholeSelectionEnd,
            insert: continuationInsert,
            continuationUnits,
            continuationCount,
            skillIds: scopedSelectedSkillIds,
          })
        : action === "rewrite"
          ? runRewrite(launchScope.projectId, {
              contentId: selectedContent.id,
              modelVariantId: activeModel.id,
              applyMode,
              guidance,
              expectedRevision,
              selectionStart: textSelection ? textSelection.start : 0,
              selectionEnd: textSelection ? textSelection.end : wholeSelectionEnd,
              skillIds: scopedSelectedSkillIds,
            })
          : runReadAndCheck(launchScope.projectId, {
              contentId: selectedContent.id,
              modelVariantId: activeModel.id,
              applyMode,
              guidance,
              expectedRevision,
              selectionStart: textSelection ? textSelection.start : 0,
              selectionEnd: textSelection ? textSelection.end : wholeSelectionEnd,
              skillIds: scopedSelectedSkillIds,
            });

    void actionRequest
      .then((result) => {
        if (!isCurrentScope(launchScope)) {
          return;
        }
        if ("candidate" in result && result.candidate) {
          setPreviewCandidate(result.candidate);
        }
        if ("content" in result && result.content) {
          onContentUpdate(result.content);
        }
        if ("assistantMessage" in result) {
          setReadCheckReport(result.assistantMessage.bodyMarkdown);
        }
        onActiveSessionChange(result.session.id);
        onSessionsChange((current) => [result.session, ...current.filter((session) => session.id !== result.session.id)]);
        onError(null);
        onRefreshTrail(launchScope.projectId);
      })
      .catch((cause: unknown) => {
        if (!isCurrentScope(launchScope)) {
          return;
        }
        onError(cause instanceof Error ? cause.message : "Quick action failed");
      })
      .finally(() => {
        if (isCurrentScope(launchScope)) {
          setIsRunningAction(false);
        }
      });
  }

  function handleAccept(candidate: GenerationCandidate): void {
    const launchScope = currentScope();
    setIsRunningAction(true);
    void acceptCandidate(launchScope.projectId, candidate.id)
      .then((result) => {
        if (!isCurrentScope(launchScope)) {
          return;
        }
        onContentUpdate(result.content);
        setPreviewCandidate(null);
        onError(null);
        onRefreshTrail(launchScope.projectId);
      })
      .catch((cause: unknown) => {
        if (!isCurrentScope(launchScope)) {
          return;
        }
        onError(cause instanceof Error ? cause.message : "Accept candidate failed");
      })
      .finally(() => {
        if (isCurrentScope(launchScope)) {
          setIsRunningAction(false);
        }
      });
  }

  function handleReject(candidate: GenerationCandidate): void {
    const launchScope = currentScope();
    setIsRunningAction(true);
    void rejectCandidate(launchScope.projectId, candidate.id)
      .then(() => {
        if (!isCurrentScope(launchScope)) {
          return;
        }
        setPreviewCandidate(null);
        onError(null);
        onRefreshTrail(launchScope.projectId);
      })
      .catch((cause: unknown) => {
        if (!isCurrentScope(launchScope)) {
          return;
        }
        onError(cause instanceof Error ? cause.message : "Reject candidate failed");
      })
      .finally(() => {
        if (isCurrentScope(launchScope)) {
          setIsRunningAction(false);
        }
      });
  }

  return (
    <aside className="agent-panel" aria-label="Agent panel">
      <header className="agent-panel-header">
        <div>
          <h2>Agent</h2>
          <p>{currentModelLabel}</p>
        </div>
        <button className="pill-button" type="button" onClick={() => setShowActivity((value) => !value)} aria-expanded={showActivity}>
          Activity {activityEvents.length}
        </button>
      </header>

      {!hasConfiguredModelVariant ? <p className="agent-warning">Configure and select a model variant to enable Agent actions.</p> : null}
      {agentError ? <p className="agent-error" role="alert">{agentError}</p> : null}

      <div className="agent-spend">
        <span>Session: {sessionSpend.tokens} tokens, est. {formatMoney(sessionSpend.cost)}</span>
        <span>
          Last request:{" "}
          {activeSessionRecord ? `${activeSessionRecord.usage.totalTokens} tokens, est. ${formatMoney(activeSessionRecord.usage.totalCost)}` : "none"}
        </span>
      </div>

      {skills.length > 0 ? (
        <SkillSelector
          skills={skills}
          selectedSkillIds={scopedSelectedSkillIds}
          disabled={isSkillSelectionLoading}
          onToggle={toggleSkill}
        />
      ) : null}

      {showActivity ? <ActivityRows events={activityEvents} records={promptRecords} /> : null}

      <section className="chat-box" aria-label="Chat transcript">
        <div className="chat-transcript">
          {chatMessages.length === 0 ? (
            <p>No chat messages in this browser session.</p>
          ) : (
            chatMessages.map((message) => (
              <article key={message.id} className="chat-message" data-role={message.role}>
                <strong>{message.role}</strong>
                <p>{message.bodyMarkdown}</p>
              </article>
            ))
          )}
        </div>
        <form className="chat-form" onSubmit={handleChatSubmit}>
          <div className="chat-input-wrap">
            <textarea
              ref={chatTextareaRef}
              value={messageInput}
              onChange={(event) => {
                setMessageInput(event.target.value);
                updateChatCursorPosition(event.target);
              }}
              onClick={(event) => updateChatCursorPosition(event.currentTarget)}
              onKeyDown={handleChatKeyDown}
              onKeyUp={(event) => updateChatCursorPosition(event.currentTarget)}
              onSelect={(event) => updateChatCursorPosition(event.currentTarget)}
              aria-autocomplete="list"
              aria-expanded={showSkillMentionSuggestions}
              aria-controls={showSkillMentionSuggestions ? skillMentionListboxId : undefined}
              aria-activedescendant={highlightedSkill ? skillMentionOptionId(highlightedSkill.id) : undefined}
              placeholder="Ask about this project..."
              disabled={!activeModel || isSendingMessage}
            />
            {showSkillMentionSuggestions ? (
              <div id={skillMentionListboxId} className="skill-mention-list" role="listbox" aria-label="Skill suggestions">
                {skillMentionSuggestions.map((skill, index) => (
                  <button
                    id={skillMentionOptionId(skill.id)}
                    key={skill.id}
                    type="button"
                    role="option"
                    aria-selected={index === visibleHighlightedSkillIndex}
                    data-active={index === visibleHighlightedSkillIndex}
                    onMouseDown={(event) => {
                      event.preventDefault();
                      acceptSkillMention(skill);
                    }}
                  >
                    <strong>${skill.name}</strong>
                    <span>{skill.displayName}</span>
                  </button>
                ))}
              </div>
            ) : null}
          </div>
          <button type="submit" disabled={!activeModel || isSendingMessage || messageInput.trim() === ""}>
            {isSendingMessage ? "Sending..." : "Send"}
          </button>
        </form>
      </section>

      <section className="quick-actions" aria-label="Quick actions">
        <header>
          <h3>Quick actions</h3>
          <span>{currentModelLabel}</span>
        </header>
        <label className="field inline-field">
          <span>Apply mode</span>
          <select value={applyMode} onChange={(event) => setApplyMode(event.target.value as ApplyMode)}>
            <option value="preview">Preview</option>
            <option value="direct_apply">Direct apply</option>
          </select>
        </label>
        <div className="field-row">
          <label className="field">
            <span>Target type</span>
            <select value={continuationUnits} onChange={(event) => setContinuationUnits(event.target.value as "word" | "sentence")}>
              <option value="word">Words</option>
              <option value="sentence">Sentences</option>
            </select>
          </label>
          <NumberField label="Target count" value={continuationCount} step={1} onChange={setContinuationCount} />
          <label className="field">
            <span>
              <input type="checkbox" checked={continuationInsert} onChange={(e) => setContinuationInsert(e.target.checked)} />
              {" "}Insert at end (instead of append)
            </span>
          </label>
        </div>
        <label className="field">
          <span>Guidance</span>
          <textarea value={guidance} onChange={(event) => setGuidance(event.target.value)} />
        </label>
        <div className="action-buttons">
          <button type="button" disabled={actionsDisabled} onClick={() => handleQuickAction("continuation")}>
            Continuation
          </button>
          <button type="button" disabled={actionsDisabled} onClick={() => handleQuickAction("rewrite")}>
            Rewrite
          </button>
          <button type="button" disabled={actionsDisabled} onClick={() => handleQuickAction("read_check")}>
            Read and Check
          </button>
        </div>
        {selectedContent && selectedContent.kind !== "chapter" ? <p className="muted">Quick actions currently target chapters.</p> : null}
      </section>

      {previewCandidate ? (
        <section className="preview-result" aria-label="Preview result">
          <header>
            <h3>{previewCandidate.actionKind.replace("_", " ")} preview</h3>
            <span>{previewCandidate.status}</span>
          </header>
          {previewCandidate.originalMarkdown ? (
            <details>
              <summary>Original text</summary>
              <pre>{previewCandidate.originalMarkdown}</pre>
            </details>
          ) : null}
          <pre>{previewCandidate.generatedMarkdown}</pre>
          <div className="action-buttons">
            <button type="button" disabled={isRunningAction} onClick={() => handleAccept(previewCandidate)}>
              Accept
            </button>
            <button type="button" disabled={isRunningAction} onClick={() => handleReject(previewCandidate)}>
              Reject
            </button>
          </div>
        </section>
      ) : null}

      {readCheckReport ? (
        <section className="preview-result" aria-label="Read and check result">
          <header>
            <h3>Read and Check</h3>
            <span>note stored</span>
          </header>
          <pre>{readCheckReport}</pre>
          <button type="button" onClick={() => setReadCheckReport(null)}>
            Dismiss
          </button>
        </section>
      ) : null}
    </aside>
  );
}

function SkillSelector({
  skills,
  selectedSkillIds,
  disabled,
  onToggle,
}: {
  skills: WriterSkill[];
  selectedSkillIds: string[];
  disabled: boolean;
  onToggle: (skillId: string, checked: boolean) => void;
}) {
  return (
    <fieldset className="skill-selector">
      <legend>Skills</legend>
      {skills.map((skill) => (
        <label key={skill.id}>
          <input
            type="checkbox"
            checked={selectedSkillIds.includes(skill.id)}
            disabled={disabled}
            onChange={(event) => onToggle(skill.id, event.target.checked)}
          />
          <span>{skill.displayName}</span>
        </label>
      ))}
    </fieldset>
  );
}

function ActivityRows({ events, records }: { events: ActivityEvent[]; records: PromptRecord[] }) {
  return (
    <div className="activity-rows">
      {events.length === 0 ? (
        <p>No activity yet.</p>
      ) : (
        events.map((event) => {
          const record = event.sessionId ? newestRecordForSession(records, event.sessionId) : null;
          return (
            <article key={event.id} className="activity-row">
              <div>
                <strong>{event.summary}</strong>
                <span>{event.eventType}</span>
              </div>
              {record ? (
                <small>
                  {record.providerName} / {record.modelName} · {record.usage.totalTokens} tokens · est. {formatMoney(record.usage.totalCost)}
                </small>
              ) : (
                <small>No usage estimate</small>
              )}
            </article>
          );
        })
      )}
    </div>
  );
}

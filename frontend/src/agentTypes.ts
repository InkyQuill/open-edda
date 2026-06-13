import type { ContentItem } from "./types";

export type ActionKind = "chat" | "continuation" | "rewrite" | "read_check";

export type ApplyMode = "preview" | "direct_apply";

export type MessageRole = "user" | "assistant" | "tool" | "system";

export type ProviderConfigSummary = {
  id: string;
  authorId: string;
  name: string;
  baseUrl: string;
  createdAt: string;
  updatedAt: string;
};

export type ProviderConfigRequest = {
  name: string;
  baseUrl: string;
  apiKey: string;
};

export type ProviderConfigUpdateRequest = {
  baseUrl: string;
  apiKey: string;
};

export type ModelVariant = {
  id: string;
  providerConfigId: string;
  name: string;
  model: string;
  temperature: number;
  maxOutputTokens: number;
  contextWindowTokens: number;
  inputPricePerMillion: number;
  outputPricePerMillion: number;
  cacheReadPricePerMillion: number;
  cacheWritePricePerMillion: number;
  requestTokenField: string;
  reasoningFormat: string;
  compatibilityJson: string;
  createdAt: string;
  updatedAt: string;
};

export type ModelVariantRequest = {
  name: string;
  model: string;
  temperature: number;
  maxOutputTokens: number;
  contextWindowTokens: number;
  inputPricePerMillion: number;
  outputPricePerMillion: number;
  cacheReadPricePerMillion: number;
  cacheWritePricePerMillion: number;
  requestTokenField: string;
  reasoningFormat: string;
  compatibilityJson: unknown;
};

export type PromptProfile = {
  id: string;
  projectId: string;
  genre: string;
  tense: string;
  pov: string;
  voice: string;
  instructionsMarkdown: string;
  promptRecordRetentionDays: number;
  createdAt: string;
  updatedAt: string;
};

export type PromptProfileRequest = {
  genre: string;
  tense: string;
  pov: string;
  voice: string;
  instructionsMarkdown: string;
  promptRecordRetentionDays: number;
};

export type AgentSession = {
  id: string;
  projectId: string;
  title: string;
  actionKind: ActionKind;
  modelVariantId: string;
  applyMode: ApplyMode;
  createdAt: string;
  updatedAt: string;
};

export type AgentMessage = {
  id: string;
  sessionId: string;
  role: MessageRole;
  bodyMarkdown: string;
  metadataJson: string;
  createdAt: string;
};

export type ChatTurnResult = {
  userMessage: AgentMessage;
  assistantMessage: AgentMessage;
  promptRecordId?: string;
};

export type GenerationCandidate = {
  id: string;
  projectId: string;
  sessionId: string;
  contentItemId: string;
  actionKind: ActionKind;
  operationKind: string;
  expectedRevision: number;
  selectionStart?: number;
  selectionEnd?: number;
  insertPosition?: number;
  originalMarkdown?: string;
  generatedMarkdown: string;
  reason?: string;
  modelVariantId?: string;
  status: string;
  createdAt: string;
  updatedAt: string;
};

export type ContinuationRequest = {
  contentId: string;
  modelVariantId: string;
  applyMode: ApplyMode;
  guidance: string;
  expectedRevision: number;
  insertPosition: number;
  insert: boolean;
  continuationUnits: "word" | "sentence";
  continuationCount: number;
};

export type RewriteRequest = {
  contentId: string;
  modelVariantId: string;
  applyMode: ApplyMode;
  guidance: string;
  expectedRevision: number;
  selectionStart: number;
  selectionEnd: number;
};

export type ContinuationResult = {
  session: AgentSession;
  candidate?: GenerationCandidate;
  content?: ContentItem;
};

export type RewriteResult = {
  session: AgentSession;
  candidate?: GenerationCandidate;
  content?: ContentItem;
};

export type AttachedNote = {
  id: string;
  projectId: string;
  contentItemId: string;
  selectionStart: number;
  selectionEnd: number;
  title: string;
  bodyMarkdown: string;
  source: string;
  createdAt: string;
  updatedAt: string;
};

export type ReadAndCheckResult = {
  session: AgentSession;
  assistantMessage: AgentMessage;
  note: AttachedNote;
};

export type AcceptCandidateResult = {
  candidate: GenerationCandidate;
  content: ContentItem;
};

export type ActivityEvent = {
  id: string;
  projectId: string;
  sessionId: string;
  eventType: string;
  summary: string;
  metadataJson: string;
  createdAt: string;
};

export type Usage = {
  inputTokens: number;
  outputTokens: number;
  cacheReadTokens: number;
  cacheWriteTokens: number;
  totalTokens: number;
  inputCost: number;
  outputCost: number;
  cacheReadCost: number;
  cacheWriteCost: number;
  totalCost: number;
};

export type PromptRecord = {
  id: string;
  projectId: string;
  sessionId: string;
  providerName: string;
  modelName: string;
  actionKind: ActionKind;
  requestJson: string;
  responseJson: string;
  usage: Usage;
  createdAt: string;
};

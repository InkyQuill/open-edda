import { createAsyncThunk } from "@reduxjs/toolkit";

import { createSession, createSessionMessage, listSessionMessages, listSessions } from "../../agentApi";

export const loadAssistantSessions = createAsyncThunk(
  "assistant/loadSessions",
  async ({ projectId }: { projectId: string }, { signal }) => listSessions(projectId, 20, signal),
);

export const startAssistantSession = createAsyncThunk(
  "assistant/startSession",
  async ({ projectId, modelVariantId }: { projectId: string; modelVariantId: string }, { signal }) =>
    createSession(
      projectId,
      {
        title: "Workspace chat",
        actionKind: "chat",
        modelVariantId,
        applyMode: "preview",
      },
      signal,
    ),
);

export const loadAssistantMessages = createAsyncThunk(
  "assistant/loadMessages",
  async ({ projectId, sessionId }: { projectId: string; sessionId: string }, { signal }) =>
    listSessionMessages(projectId, sessionId, signal),
);

export const sendAssistantMessage = createAsyncThunk(
  "assistant/sendMessage",
  async (
    {
      projectId,
      sessionId,
      bodyMarkdown,
    }: {
      projectId: string;
      sessionId: string;
      bodyMarkdown: string;
    },
    { signal },
  ) => createSessionMessage(projectId, sessionId, bodyMarkdown, signal),
);

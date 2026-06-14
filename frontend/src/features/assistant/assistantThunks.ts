import { createAsyncThunk } from "@reduxjs/toolkit";

import { createSession, createSessionMessage, listSessions } from "../../agentApi";

export const loadAssistantSessions = createAsyncThunk(
  "assistant/loadSessions",
  async ({ projectId }: { projectId: string }) => listSessions(projectId, 20),
);

export const startAssistantSession = createAsyncThunk(
  "assistant/startSession",
  async ({ projectId, contentId }: { projectId: string; contentId: string | null }) =>
    createSession(projectId, {
      contentId,
      title: "Workspace chat",
      actionKind: "chat",
    }),
);

export const sendAssistantMessage = createAsyncThunk(
  "assistant/sendMessage",
  async ({
    projectId,
    sessionId,
    bodyMarkdown,
  }: {
    projectId: string;
    sessionId: string;
    bodyMarkdown: string;
  }) => createSessionMessage(projectId, sessionId, bodyMarkdown),
);

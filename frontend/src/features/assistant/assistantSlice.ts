import { createSlice, type PayloadAction } from "@reduxjs/toolkit";

import type { AgentMessage, AgentSession } from "../../agentTypes";
import {
  loadAssistantMessages,
  loadAssistantSessions,
  sendAssistantMessage,
  startAssistantSession,
} from "./assistantThunks";

export type AsyncStatus = "idle" | "pending" | "succeeded" | "failed";

export type AssistantState = {
  projectId: string | null;
  activeSessionId: string | null;
  draftMessage: string;
  sessions: AgentSession[];
  messagesBySessionId: Record<string, AgentMessage[]>;
  sessionsStatus: AsyncStatus;
  messagesStatus: AsyncStatus;
  sessionsRequestId: string | null;
  startSessionRequestId: string | null;
  messagesRequestId: string | null;
  sendMessageRequestId: string | null;
  error: string | null;
};

export const initialAssistantState: AssistantState = {
  projectId: null,
  activeSessionId: null,
  draftMessage: "",
  sessions: [],
  messagesBySessionId: {},
  sessionsStatus: "idle",
  messagesStatus: "idle",
  sessionsRequestId: null,
  startSessionRequestId: null,
  messagesRequestId: null,
  sendMessageRequestId: null,
  error: null,
};

function readableError(action: { error: { message?: string } }): string {
  return action.error.message ?? "Assistant request failed";
}

function matchesProject(state: AssistantState, projectId: string): boolean {
  return state.projectId === projectId;
}

const assistantSlice = createSlice({
  name: "assistant",
  initialState: initialAssistantState,
  reducers: {
    setDraftMessage(state, action: PayloadAction<string>) {
      state.draftMessage = action.payload;
    },
    setActiveSessionId(state, action: PayloadAction<string | null>) {
      state.activeSessionId = action.payload;
    },
    resetForProject() {
      return initialAssistantState;
    },
  },
  extraReducers: (builder) => {
    builder
      .addCase(loadAssistantSessions.pending, (state, action) => {
        state.projectId = action.meta.arg.projectId;
        state.sessionsRequestId = action.meta.requestId;
        state.sessionsStatus = "pending";
        state.error = null;
      })
      .addCase(loadAssistantSessions.fulfilled, (state, action) => {
        if (!matchesProject(state, action.meta.arg.projectId) || state.sessionsRequestId !== action.meta.requestId) {
          return;
        }
        state.sessionsStatus = "succeeded";
        state.sessionsRequestId = null;
        state.sessions = action.payload;
        state.error = null;

        if (!state.activeSessionId || !action.payload.some((session) => session.id === state.activeSessionId)) {
          state.activeSessionId = action.payload[0]?.id ?? null;
        }
      })
      .addCase(loadAssistantSessions.rejected, (state, action) => {
        if (!matchesProject(state, action.meta.arg.projectId) || state.sessionsRequestId !== action.meta.requestId) {
          return;
        }
        state.sessionsStatus = "failed";
        state.sessionsRequestId = null;
        state.error = readableError(action);
      })
      .addCase(startAssistantSession.pending, (state, action) => {
        state.projectId = action.meta.arg.projectId;
        state.startSessionRequestId = action.meta.requestId;
        state.sessionsStatus = "pending";
        state.error = null;
      })
      .addCase(startAssistantSession.fulfilled, (state, action) => {
        if (!matchesProject(state, action.meta.arg.projectId) || state.startSessionRequestId !== action.meta.requestId) {
          return;
        }
        state.sessionsStatus = "succeeded";
        state.startSessionRequestId = null;
        state.sessions = [action.payload, ...state.sessions.filter((session) => session.id !== action.payload.id)];
        state.activeSessionId = action.payload.id;
        state.messagesBySessionId[action.payload.id] = [];
        state.error = null;
      })
      .addCase(startAssistantSession.rejected, (state, action) => {
        if (!matchesProject(state, action.meta.arg.projectId) || state.startSessionRequestId !== action.meta.requestId) {
          return;
        }
        state.sessionsStatus = "failed";
        state.startSessionRequestId = null;
        state.error = readableError(action);
      })
      .addCase(loadAssistantMessages.pending, (state, action) => {
        state.projectId = action.meta.arg.projectId;
        state.messagesRequestId = action.meta.requestId;
        state.messagesStatus = "pending";
        state.error = null;
      })
      .addCase(loadAssistantMessages.fulfilled, (state, action) => {
        if (
          !matchesProject(state, action.meta.arg.projectId) ||
          state.messagesRequestId !== action.meta.requestId ||
          state.activeSessionId !== action.meta.arg.sessionId
        ) {
          return;
        }
        state.messagesStatus = "succeeded";
        state.messagesRequestId = null;
        state.messagesBySessionId[action.meta.arg.sessionId] = action.payload;
        state.error = null;
      })
      .addCase(loadAssistantMessages.rejected, (state, action) => {
        if (
          !matchesProject(state, action.meta.arg.projectId) ||
          state.messagesRequestId !== action.meta.requestId ||
          state.activeSessionId !== action.meta.arg.sessionId
        ) {
          return;
        }
        state.messagesStatus = "failed";
        state.messagesRequestId = null;
        state.error = readableError(action);
      })
      .addCase(sendAssistantMessage.pending, (state, action) => {
        state.projectId = action.meta.arg.projectId;
        // Ignore in-flight transcript loads that predate the optimistic send.
        state.messagesRequestId = null;
        state.sendMessageRequestId = action.meta.requestId;
        state.messagesStatus = "pending";
        state.error = null;
      })
      .addCase(sendAssistantMessage.fulfilled, (state, action) => {
        if (!matchesProject(state, action.meta.arg.projectId) || state.sendMessageRequestId !== action.meta.requestId) {
          return;
        }
        state.messagesStatus = "succeeded";
        state.sendMessageRequestId = null;
        state.draftMessage = "";
        state.error = null;

        const sessionId = action.meta.arg.sessionId;
        state.messagesBySessionId[sessionId] = [
          ...(state.messagesBySessionId[sessionId] ?? []),
          action.payload.userMessage,
          action.payload.assistantMessage,
        ];
      })
      .addCase(sendAssistantMessage.rejected, (state, action) => {
        if (!matchesProject(state, action.meta.arg.projectId) || state.sendMessageRequestId !== action.meta.requestId) {
          return;
        }
        state.messagesStatus = "failed";
        state.sendMessageRequestId = null;
        state.error = readableError(action);
      });
  },
});

export const assistantActions = assistantSlice.actions;
export const assistantReducer = assistantSlice.reducer;

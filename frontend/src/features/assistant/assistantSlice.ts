import { createSlice, type PayloadAction } from "@reduxjs/toolkit";

import type { AgentMessage, AgentSession } from "../../agentTypes";
import {
  loadAssistantSessions,
  sendAssistantMessage,
  startAssistantSession,
} from "./assistantThunks";

export type AsyncStatus = "idle" | "pending" | "succeeded" | "failed";

export type AssistantState = {
  activeSessionId: string | null;
  draftMessage: string;
  sessions: AgentSession[];
  messagesBySessionId: Record<string, AgentMessage[]>;
  sessionsStatus: AsyncStatus;
  messagesStatus: AsyncStatus;
  error: string | null;
};

export const initialAssistantState: AssistantState = {
  activeSessionId: null,
  draftMessage: "",
  sessions: [],
  messagesBySessionId: {},
  sessionsStatus: "idle",
  messagesStatus: "idle",
  error: null,
};

function readableError(action: { error: { message?: string } }): string {
  return action.error.message ?? "Assistant request failed";
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
      .addCase(loadAssistantSessions.pending, (state) => {
        state.sessionsStatus = "pending";
        state.error = null;
      })
      .addCase(loadAssistantSessions.fulfilled, (state, action) => {
        state.sessionsStatus = "succeeded";
        state.sessions = action.payload;
        state.error = null;

        if (!state.activeSessionId || !action.payload.some((session) => session.id === state.activeSessionId)) {
          state.activeSessionId = action.payload[0]?.id ?? null;
        }
      })
      .addCase(loadAssistantSessions.rejected, (state, action) => {
        state.sessionsStatus = "failed";
        state.error = readableError(action);
      })
      .addCase(startAssistantSession.pending, (state) => {
        state.sessionsStatus = "pending";
        state.error = null;
      })
      .addCase(startAssistantSession.fulfilled, (state, action) => {
        state.sessionsStatus = "succeeded";
        state.sessions = [action.payload, ...state.sessions.filter((session) => session.id !== action.payload.id)];
        state.activeSessionId = action.payload.id;
        state.error = null;
      })
      .addCase(startAssistantSession.rejected, (state, action) => {
        state.sessionsStatus = "failed";
        state.error = readableError(action);
      })
      .addCase(sendAssistantMessage.pending, (state) => {
        state.messagesStatus = "pending";
        state.error = null;
      })
      .addCase(sendAssistantMessage.fulfilled, (state, action) => {
        state.messagesStatus = "succeeded";
        state.draftMessage = "";
        state.error = null;

        const sessionId = action.payload.userMessage.sessionId;
        state.messagesBySessionId[sessionId] = [
          ...(state.messagesBySessionId[sessionId] ?? []),
          action.payload.userMessage,
          action.payload.assistantMessage,
        ];
      })
      .addCase(sendAssistantMessage.rejected, (state, action) => {
        state.messagesStatus = "failed";
        state.error = readableError(action);
      });
  },
});

export const assistantActions = assistantSlice.actions;
export const assistantReducer = assistantSlice.reducer;

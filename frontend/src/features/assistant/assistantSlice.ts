import { createSlice } from "@reduxjs/toolkit";

export type AsyncStatus = "idle" | "pending" | "succeeded" | "failed";

export type AssistantState = {
  activeSessionId: string | null;
  draftMessage: string;
  sessionsStatus: AsyncStatus;
  messagesStatus: AsyncStatus;
  error: string | null;
};

export const initialAssistantState: AssistantState = {
  activeSessionId: null,
  draftMessage: "",
  sessionsStatus: "idle",
  messagesStatus: "idle",
  error: null,
};

const assistantSlice = createSlice({
  name: "assistant",
  initialState: initialAssistantState,
  reducers: {
    resetForProject() {
      return initialAssistantState;
    },
  },
});

export const assistantActions = assistantSlice.actions;
export const assistantReducer = assistantSlice.reducer;

import { createAsyncThunk, createSlice } from "@reduxjs/toolkit";

import {
  acceptCandidate,
  rejectCandidate,
  runContinuation,
  runReadAndCheck,
  runRewrite,
} from "../../agentApi";
import type {
  GenerationCandidate,
  ReadAndCheckResult,
} from "../../agentTypes";

export type AssistantActionKind = "generate" | "rewrite" | "check";
export type AssistantActionStatus = "idle" | "running" | "succeeded" | "failed";

export interface AssistantActionState {
  actionKind: AssistantActionKind | null;
  status: AssistantActionStatus;
  acceptStatus: AssistantActionStatus;
  rejectStatus: AssistantActionStatus;
  projectId: string | null;
  contentId: string | null;
  requestKey: string | null;
  requestToken: string | null;
  candidate: GenerationCandidate | null;
  checkResult: ReadAndCheckResult | null;
  error: string | null;
  acceptError: string | null;
  rejectError: string | null;
}

export interface RunGenerateArgs {
  projectId: string;
  contentId: string;
  modelVariantId: string;
  expectedRevision: number;
  cursorByte: number;
  instructions: string;
  skillIds?: string[];
  requestKey: string;
  requestToken: string;
}

export interface RunRewriteArgs {
  projectId: string;
  contentId: string;
  modelVariantId: string;
  expectedRevision: number;
  selectionStartByte: number;
  selectionEndByte: number;
  instructions: string;
  skillIds?: string[];
  requestKey: string;
  requestToken: string;
}

export type RunCheckArgs = RunRewriteArgs;

export interface AcceptAssistantCandidateArgs {
  projectId: string;
  candidateId: string;
  requestKey: string;
  requestToken: string;
}

export type RejectAssistantCandidateArgs = AcceptAssistantCandidateArgs;

export const initialAssistantActionsState: AssistantActionState = {
  actionKind: null,
  status: "idle",
  acceptStatus: "idle",
  rejectStatus: "idle",
  projectId: null,
  contentId: null,
  requestKey: null,
  requestToken: null,
  candidate: null,
  checkResult: null,
  error: null,
  acceptError: null,
  rejectError: null,
};

export const runGenerate = createAsyncThunk(
  "assistantActions/runGenerate",
  async (args: RunGenerateArgs) =>
    runContinuation(args.projectId, {
      contentId: args.contentId,
      modelVariantId: args.modelVariantId,
      expectedRevision: args.expectedRevision,
      insertPosition: args.cursorByte,
      guidance: args.instructions,
      skillIds: args.skillIds,
      applyMode: "preview",
      insert: true,
      continuationUnits: "paragraph",
      continuationCount: 1,
    }),
);

export const runRewriteAction = createAsyncThunk(
  "assistantActions/runRewrite",
  async (args: RunRewriteArgs) =>
    runRewrite(args.projectId, {
      contentId: args.contentId,
      modelVariantId: args.modelVariantId,
      expectedRevision: args.expectedRevision,
      selectionStart: args.selectionStartByte,
      selectionEnd: args.selectionEndByte,
      guidance: args.instructions,
      skillIds: args.skillIds,
      applyMode: "preview",
    }),
);

export const runCheck = createAsyncThunk(
  "assistantActions/runCheck",
  async (args: RunCheckArgs) =>
    runReadAndCheck(args.projectId, {
      contentId: args.contentId,
      modelVariantId: args.modelVariantId,
      expectedRevision: args.expectedRevision,
      selectionStart: args.selectionStartByte,
      selectionEnd: args.selectionEndByte,
      guidance: args.instructions,
      skillIds: args.skillIds,
      applyMode: "preview",
    }),
);

export const acceptAssistantCandidate = createAsyncThunk(
  "assistantActions/acceptCandidate",
  async (args: AcceptAssistantCandidateArgs) =>
    acceptCandidate(args.projectId, args.candidateId),
);

export const rejectAssistantCandidate = createAsyncThunk(
  "assistantActions/rejectCandidate",
  async (args: RejectAssistantCandidateArgs) =>
    rejectCandidate(args.projectId, args.candidateId),
);

function errorMessage(error: { message?: string } | unknown): string {
  if (
    typeof error === "object" &&
    error !== null &&
    "message" in error &&
    typeof error.message === "string" &&
    error.message.trim()
  ) {
    return error.message;
  }
  return "Action failed";
}

function matchesRequest(
  state: AssistantActionState,
  arg: { requestKey: string; requestToken: string },
): boolean {
  return state.requestKey === arg.requestKey && state.requestToken === arg.requestToken;
}

function markPending(
  state: AssistantActionState,
  actionKind: AssistantActionKind,
  arg: {
    projectId: string;
    contentId: string;
    requestKey: string;
    requestToken: string;
  },
) {
  state.actionKind = actionKind;
  state.status = "running";
  state.acceptStatus = "idle";
  state.rejectStatus = "idle";
  state.projectId = arg.projectId;
  state.contentId = arg.contentId;
  state.requestKey = arg.requestKey;
  state.requestToken = arg.requestToken;
  state.candidate = null;
  state.checkResult = null;
  state.error = null;
  state.acceptError = null;
  state.rejectError = null;
}

function markRejected(
  state: AssistantActionState,
  arg: { requestKey: string; requestToken: string },
  message: string,
) {
  if (!matchesRequest(state, arg)) return;
  state.status = "failed";
  state.candidate = null;
  state.checkResult = null;
  state.error = message;
}

function isRevisionConflictError(message: string): boolean {
  const normalized = message.toLowerCase();
  return (
    normalized.includes("revision") ||
    normalized.includes("conflict") ||
    normalized.includes("stale")
  );
}

function acceptErrorMessage(message: string): string {
  if (!isRevisionConflictError(message)) {
    return message;
  }
  return "Content changed before this preview was accepted. Review the latest draft, then run the action again.";
}

const assistantActionsSlice = createSlice({
  name: "assistantActions",
  initialState: initialAssistantActionsState,
  reducers: {
    clearAssistantActionResult() {
      return initialAssistantActionsState;
    },
  },
  extraReducers: (builder) => {
    builder
      .addCase(runGenerate.pending, (state, action) => {
        markPending(state, "generate", action.meta.arg);
      })
      .addCase(runGenerate.fulfilled, (state, action) => {
        if (!matchesRequest(state, action.meta.arg)) return;
        state.status = "succeeded";
        state.candidate = action.payload.candidate ?? null;
        state.checkResult = null;
        state.error = null;
      })
      .addCase(runGenerate.rejected, (state, action) => {
        markRejected(state, action.meta.arg, errorMessage(action.error));
      })
      .addCase(runRewriteAction.pending, (state, action) => {
        markPending(state, "rewrite", action.meta.arg);
      })
      .addCase(runRewriteAction.fulfilled, (state, action) => {
        if (!matchesRequest(state, action.meta.arg)) return;
        state.status = "succeeded";
        state.candidate = action.payload.candidate ?? null;
        state.checkResult = null;
        state.error = null;
      })
      .addCase(runRewriteAction.rejected, (state, action) => {
        markRejected(state, action.meta.arg, errorMessage(action.error));
      })
      .addCase(runCheck.pending, (state, action) => {
        markPending(state, "check", action.meta.arg);
      })
      .addCase(runCheck.fulfilled, (state, action) => {
        if (!matchesRequest(state, action.meta.arg)) return;
        state.status = "succeeded";
        state.candidate = null;
        state.checkResult = action.payload;
        state.error = null;
      })
      .addCase(runCheck.rejected, (state, action) => {
        markRejected(state, action.meta.arg, errorMessage(action.error));
      })
      .addCase(acceptAssistantCandidate.pending, (state, action) => {
        if (!matchesRequest(state, action.meta.arg)) return;
        state.acceptStatus = "running";
        state.acceptError = null;
      })
      .addCase(acceptAssistantCandidate.fulfilled, (state, action) => {
        if (!matchesRequest(state, action.meta.arg)) return;
        state.candidate = null;
        state.acceptStatus = "succeeded";
        state.acceptError = null;
      })
      .addCase(acceptAssistantCandidate.rejected, (state, action) => {
        if (!matchesRequest(state, action.meta.arg)) return;
        state.acceptStatus = "failed";
        state.acceptError = acceptErrorMessage(errorMessage(action.error));
      })
      .addCase(rejectAssistantCandidate.pending, (state, action) => {
        if (!matchesRequest(state, action.meta.arg)) return;
        state.rejectStatus = "running";
        state.rejectError = null;
      })
      .addCase(rejectAssistantCandidate.fulfilled, (state, action) => {
        if (!matchesRequest(state, action.meta.arg)) return;
        state.candidate = null;
        state.rejectStatus = "succeeded";
        state.rejectError = null;
      })
      .addCase(rejectAssistantCandidate.rejected, (state, action) => {
        if (!matchesRequest(state, action.meta.arg)) return;
        state.rejectStatus = "failed";
        state.rejectError = errorMessage(action.error);
      });
  },
});

export const assistantActions = assistantActionsSlice.actions;
export const assistantActionsReducer = assistantActionsSlice.reducer;

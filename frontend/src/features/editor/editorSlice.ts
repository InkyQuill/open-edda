import { createSlice, type PayloadAction } from "@reduxjs/toolkit";

import type { ContentKind } from "../../types";

export type EditorActionKind = "rewrite" | "check" | "note";
export type EditorSaveStatus = "idle" | "pending" | "succeeded" | "failed";

export type EditorContentContext = {
  projectId: string;
  contentId: string;
  contentKind: ContentKind;
  revision: number;
};

export type EditorSelection = {
  startByte: number;
  endByte: number;
  previewText: string;
} | null;

export type EditorActionModal = {
  kind: EditorActionKind;
  instructions: string;
} | null;

export type EditorState = {
  contentContext: EditorContentContext | null;
  draftMarkdown: string;
  persistedMarkdown: string;
  dirty: boolean;
  saveStatus: EditorSaveStatus;
  saveError: string | null;
  cursorByte: number | null;
  selection: EditorSelection;
  actionModal: EditorActionModal;
  generateInstructions: string;
};

export const initialEditorState: EditorState = {
  contentContext: null,
  draftMarkdown: "",
  persistedMarkdown: "",
  dirty: false,
  saveStatus: "idle",
  saveError: null,
  cursorByte: null,
  selection: null,
  actionModal: null,
  generateInstructions: "",
};

const editorSlice = createSlice({
  name: "editor",
  initialState: initialEditorState,
  reducers: {
    hydrateEditorContext(
      state,
      action: PayloadAction<EditorContentContext & { bodyMarkdown: string }>,
    ) {
      state.contentContext = {
        projectId: action.payload.projectId,
        contentId: action.payload.contentId,
        contentKind: action.payload.contentKind,
        revision: action.payload.revision,
      };
      state.draftMarkdown = action.payload.bodyMarkdown;
      state.persistedMarkdown = action.payload.bodyMarkdown;
      state.dirty = false;
      state.saveStatus = "idle";
      state.saveError = null;
      state.cursorByte = null;
      state.selection = null;
      state.actionModal = null;
      state.generateInstructions = "";
    },
    setDraftMarkdown(state, action: PayloadAction<string>) {
      state.draftMarkdown = action.payload;
      state.dirty = state.draftMarkdown !== state.persistedMarkdown;
      state.saveStatus = "idle";
      state.saveError = null;
    },
    saveStarted(state) {
      state.saveStatus = "pending";
      state.saveError = null;
    },
    saveSucceeded(state, action: PayloadAction<{ revision: number; bodyMarkdown: string }>) {
      if (state.contentContext) {
        state.contentContext.revision = action.payload.revision;
      }
      state.persistedMarkdown = action.payload.bodyMarkdown;
      state.draftMarkdown = action.payload.bodyMarkdown;
      state.dirty = false;
      state.saveStatus = "succeeded";
      state.saveError = null;
      state.cursorByte = null;
      state.selection = null;
      state.actionModal = null;
    },
    saveFailed(state, action: PayloadAction<string>) {
      state.saveStatus = "failed";
      state.saveError = action.payload;
      state.dirty = state.draftMarkdown !== state.persistedMarkdown;
    },
    setCursorByte(state, action: PayloadAction<number | null>) {
      state.cursorByte = action.payload;
    },
    setSelection(state, action: PayloadAction<EditorSelection>) {
      state.selection = action.payload;
      if (!action.payload) {
        state.actionModal = null;
      }
    },
    openActionModal(
      state,
      action: PayloadAction<{
        kind: EditorActionKind;
        instructions?: string;
      }>,
    ) {
      state.actionModal = {
        kind: action.payload.kind,
        instructions: action.payload.instructions ?? "",
      };
    },
    setActionInstructions(state, action: PayloadAction<string>) {
      if (state.actionModal) {
        state.actionModal.instructions = action.payload;
      }
    },
    closeActionModal(state) {
      state.actionModal = null;
    },
    setGenerateInstructions(state, action: PayloadAction<string>) {
      state.generateInstructions = action.payload;
    },
    resetEditorContext(state) {
      state.contentContext = null;
      state.draftMarkdown = "";
      state.persistedMarkdown = "";
      state.dirty = false;
      state.saveStatus = "idle";
      state.saveError = null;
      state.cursorByte = null;
      state.selection = null;
      state.actionModal = null;
      state.generateInstructions = "";
    },
  },
});

export const editorActions = editorSlice.actions;
export const editorReducer = editorSlice.reducer;

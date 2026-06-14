import { createSlice, type PayloadAction } from "@reduxjs/toolkit";

export type EditorActionKind = "rewrite" | "check" | "note";

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
  cursorByte: number | null;
  selection: EditorSelection;
  actionModal: EditorActionModal;
  generateInstructions: string;
};

export const initialEditorState: EditorState = {
  cursorByte: null,
  selection: null,
  actionModal: null,
  generateInstructions: "",
};

const editorSlice = createSlice({
  name: "editor",
  initialState: initialEditorState,
  reducers: {
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
      state.cursorByte = null;
      state.selection = null;
      state.actionModal = null;
      state.generateInstructions = "";
    },
  },
});

export const editorActions = editorSlice.actions;
export const editorReducer = editorSlice.reducer;

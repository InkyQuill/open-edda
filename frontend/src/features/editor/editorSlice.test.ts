import { describe, expect, it } from "vitest";
import { editorActions, editorReducer, initialEditorState } from "./editorSlice";

describe("editorSlice", () => {
  it("stores selection and opens rewrite modal with draft instructions", () => {
    const selected = editorReducer(
      initialEditorState,
      editorActions.setSelection({
        startByte: 4,
        endByte: 16,
        previewText: "selected prose",
      }),
    );
    const modal = editorReducer(
      selected,
      editorActions.openActionModal({
        kind: "rewrite",
        instructions: "tighten",
      }),
    );

    expect(modal.selection?.previewText).toBe("selected prose");
    expect(modal.actionModal).toEqual({
      kind: "rewrite",
      instructions: "tighten",
    });
  });

  it("closes action modal when selection is cleared", () => {
    const selected = editorReducer(
      initialEditorState,
      editorActions.setSelection({
        startByte: 4,
        endByte: 16,
        previewText: "selected prose",
      }),
    );
    const modal = editorReducer(
      selected,
      editorActions.openActionModal({ kind: "check" }),
    );
    const cleared = editorReducer(modal, editorActions.setSelection(null));

    expect(cleared.selection).toBeNull();
    expect(cleared.actionModal).toBeNull();
  });

  it("resets generate instructions with editor context", () => {
    const prompted = editorReducer(
      initialEditorState,
      editorActions.setGenerateInstructions("continue in this tone"),
    );
    const reset = editorReducer(prompted, editorActions.resetEditorContext());

    expect(reset.generateInstructions).toBe("");
  });
});

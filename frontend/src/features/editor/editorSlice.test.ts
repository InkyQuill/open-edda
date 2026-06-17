import { describe, expect, it } from "vitest";
import { editorActions, editorReducer, initialEditorState } from "./editorSlice";

describe("editorSlice", () => {
  it("hydrates content context with clean draft markdown", () => {
    const state = editorReducer(
      initialEditorState,
      editorActions.hydrateEditorContext({
        projectId: "project-1",
        contentId: "chapter-1",
        contentKind: "chapter",
        revision: 3,
        bodyMarkdown: "Opening text.",
      }),
    );

    expect(state.contentContext).toEqual({
      projectId: "project-1",
      contentId: "chapter-1",
      contentKind: "chapter",
      revision: 3,
    });
    expect(state.persistedMarkdown).toBe("Opening text.");
    expect(state.draftMarkdown).toBe("Opening text.");
    expect(state.dirty).toBe(false);
    expect(state.saveStatus).toBe("idle");
    expect(state.saveError).toBeNull();
  });

  it("tracks local draft edits without changing persisted markdown", () => {
    const hydrated = editorReducer(
      initialEditorState,
      editorActions.hydrateEditorContext({
        projectId: "project-1",
        contentId: "chapter-1",
        contentKind: "chapter",
        revision: 3,
        bodyMarkdown: "Opening text.",
      }),
    );
    const edited = editorReducer(hydrated, editorActions.setDraftMarkdown("Opening text.\n\nMore."));

    expect(edited.persistedMarkdown).toBe("Opening text.");
    expect(edited.draftMarkdown).toBe("Opening text.\n\nMore.");
    expect(edited.dirty).toBe(true);
  });

  it("records save success as the new clean revision", () => {
    const hydrated = editorReducer(
      initialEditorState,
      editorActions.hydrateEditorContext({
        projectId: "project-1",
        contentId: "chapter-1",
        contentKind: "chapter",
        revision: 3,
        bodyMarkdown: "Opening text.",
      }),
    );
    const saving = editorReducer(
      editorReducer(hydrated, editorActions.setDraftMarkdown("Revised text.")),
      editorActions.saveStarted(),
    );
    const saved = editorReducer(
      saving,
      editorActions.saveSucceeded({
        revision: 4,
        bodyMarkdown: "Revised text.",
      }),
    );

    expect(saved.contentContext?.revision).toBe(4);
    expect(saved.persistedMarkdown).toBe("Revised text.");
    expect(saved.draftMarkdown).toBe("Revised text.");
    expect(saved.dirty).toBe(false);
    expect(saved.saveStatus).toBe("succeeded");
    expect(saved.saveError).toBeNull();
  });

  it("records save failure without discarding the local draft", () => {
    const hydrated = editorReducer(
      initialEditorState,
      editorActions.hydrateEditorContext({
        projectId: "project-1",
        contentId: "chapter-1",
        contentKind: "chapter",
        revision: 3,
        bodyMarkdown: "Opening text.",
      }),
    );
    const edited = editorReducer(hydrated, editorActions.setDraftMarkdown("Unsaved text."));
    const failed = editorReducer(edited, editorActions.saveFailed("content revision conflict"));

    expect(failed.draftMarkdown).toBe("Unsaved text.");
    expect(failed.persistedMarkdown).toBe("Opening text.");
    expect(failed.dirty).toBe(true);
    expect(failed.saveStatus).toBe("failed");
    expect(failed.saveError).toBe("content revision conflict");
  });

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

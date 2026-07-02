import { describe, expect, it } from "vitest";

import type { EditorContentContext } from "./editorSlice";
import {
  buildCheckRequest,
  buildGenerateRequest,
  buildRewriteRequest,
  buildAssistantActionRequestKey,
  validateGenerateAction,
  validateSelectionAction,
} from "./assistantActionRequests";

const contentContext: EditorContentContext = {
  projectId: "project-1",
  contentId: "content-1",
  contentKind: "chapter",
  revision: 7,
};

describe("assistantActionRequests", () => {
  it("validates missing generate prerequisites in user-facing language", () => {
    expect(
      validateGenerateAction({
        contentContext: null,
        cursorByte: 12,
        activeModelVariantId: "model-1",
        dirty: false,
      }),
    ).toBe("Open a chapter or story-bible entry before generating.");
    expect(
      validateGenerateAction({
        contentContext,
        cursorByte: 12,
        activeModelVariantId: null,
        dirty: false,
      }),
    ).toBe("Choose an active model variant in Settings before generating.");
    expect(
      validateGenerateAction({
        contentContext,
        cursorByte: null,
        activeModelVariantId: "model-1",
        dirty: false,
      }),
    ).toBe("Place the cursor where generated text should be inserted.");
  });

  it("returns null when generate prerequisites are present", () => {
    expect(
      validateGenerateAction({
        contentContext,
        cursorByte: 12,
        activeModelVariantId: "model-1",
        dirty: false,
      }),
    ).toBeNull();
  });

  it("blocks assistant actions while the editor draft is unsaved", () => {
    expect(
      validateGenerateAction({
        contentContext,
        cursorByte: 12,
        activeModelVariantId: "model-1",
        dirty: true,
      }),
    ).toBe("Save the current draft before generating.");
    expect(
      validateSelectionAction({
        contentContext,
        selection: { startByte: 3, endByte: 9 },
        activeModelVariantId: "model-1",
        dirty: true,
      }),
    ).toBe("Save the current draft before running this action.");
  });

  it("builds stable request keys from route and content revision", () => {
    expect(buildAssistantActionRequestKey("/projects/project-1/content/chapter/content-1", contentContext)).toBe(
      "/projects/project-1/content/chapter/content-1:project-1:content-1:7",
    );
  });

  it("builds generate thunk args from editor state", () => {
    expect(
      buildGenerateRequest({
        contentContext,
        activeModelVariantId: "model-1",
        cursorByte: 12,
        instructions: " continue here ",
        skillIds: ["skill-1"],
        requestKey: "route-key",
        requestToken: "token-1",
      }),
    ).toEqual({
      projectId: "project-1",
      contentId: "content-1",
      modelVariantId: "model-1",
      expectedRevision: 7,
      cursorByte: 12,
      instructions: "continue here",
      skillIds: ["skill-1"],
      requestKey: "route-key",
      requestToken: "token-1",
    });
  });

  it("validates missing selection action prerequisites", () => {
    expect(
      validateSelectionAction({
        contentContext: null,
        selection: { startByte: 3, endByte: 9 },
        activeModelVariantId: "model-1",
        dirty: false,
      }),
    ).toBe("Open a chapter or story-bible entry before running this action.");
    expect(
      validateSelectionAction({
        contentContext,
        selection: { startByte: 3, endByte: 9 },
        activeModelVariantId: null,
        dirty: false,
      }),
    ).toBe("Choose an active model variant in Settings before running this action.");
    expect(
      validateSelectionAction({
        contentContext,
        selection: null,
        activeModelVariantId: "model-1",
        dirty: false,
      }),
    ).toBe("Select text before running this action.");
    expect(
      validateSelectionAction({
        contentContext,
        selection: { startByte: 9, endByte: 9 },
        activeModelVariantId: "model-1",
        dirty: false,
      }),
    ).toBe("Select text before running this action.");
  });

  it("builds rewrite and check thunk args from editor selection", () => {
    const base = {
      contentContext,
      activeModelVariantId: "model-1",
      selection: { startByte: 3, endByte: 18 },
      instructions: " focus on tone ",
      skillIds: ["skill-1"],
      requestKey: "route-key",
      requestToken: "token-1",
    };

    expect(buildRewriteRequest(base)).toEqual({
      projectId: "project-1",
      contentId: "content-1",
      modelVariantId: "model-1",
      expectedRevision: 7,
      selectionStartByte: 3,
      selectionEndByte: 18,
      instructions: "focus on tone",
      skillIds: ["skill-1"],
      requestKey: "route-key",
      requestToken: "token-1",
    });
    expect(buildCheckRequest(base)).toEqual(buildRewriteRequest(base));
  });
});

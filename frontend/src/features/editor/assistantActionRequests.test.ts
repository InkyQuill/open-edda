import { describe, expect, it } from "vitest";

import type { EditorContentContext } from "./editorSlice";
import {
  buildGenerateRequest,
  buildAssistantActionRequestKey,
  validateGenerateAction,
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
      }),
    ).toBe("Open a chapter or story-bible entry before generating.");
    expect(
      validateGenerateAction({
        contentContext,
        cursorByte: 12,
        activeModelVariantId: null,
      }),
    ).toBe("Choose an active model variant in Settings before generating.");
    expect(
      validateGenerateAction({
        contentContext,
        cursorByte: null,
        activeModelVariantId: "model-1",
      }),
    ).toBe("Place the cursor where generated text should be inserted.");
  });

  it("returns null when generate prerequisites are present", () => {
    expect(
      validateGenerateAction({
        contentContext,
        cursorByte: 12,
        activeModelVariantId: "model-1",
      }),
    ).toBeNull();
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
});

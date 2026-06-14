import { describe, expect, it } from "vitest";
import { initialReviewState, reviewActions, reviewReducer } from "./reviewSlice";

describe("reviewSlice", () => {
  it("resets project-scoped review state", () => {
    const loaded = reviewReducer(
      {
        ...initialReviewState,
        activityStatus: "succeeded",
        promptRecordsStatus: "failed",
        selectedPromptRecordId: "prompt-record-1",
      },
      reviewActions.resetForProject(),
    );

    expect(loaded).toEqual(initialReviewState);
  });
});

import { describe, expect, it } from "vitest";
import {
  initialModelSettingsState,
  modelSettingsActions,
  modelSettingsReducer,
} from "./modelSettingsSlice";

describe("modelSettingsSlice", () => {
  it("resets project-scoped model settings state", () => {
    const loaded = modelSettingsReducer(
      {
        ...initialModelSettingsState,
        selectedProviderId: "provider-1",
        activeModelVariantId: "model-1",
        providersStatus: "succeeded",
      },
      modelSettingsActions.resetForProject(),
    );

    expect(loaded).toEqual(initialModelSettingsState);
  });
});

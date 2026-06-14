import { describe, expect, it } from "vitest";
import type { ModelVariant, ProviderConfigSummary } from "../../agentTypes";
import { loadModelVariants, loadProviderConfigs } from "./modelSettingsThunks";
import {
  initialModelSettingsState,
  modelSettingsActions,
  modelSettingsReducer,
  type ModelSettingsState,
} from "./modelSettingsSlice";

const providerAlpha: ProviderConfigSummary = {
  id: "provider-alpha",
  authorId: "author-1",
  name: "Provider Alpha",
  baseUrl: "https://alpha.example",
  createdAt: "2026-06-14T00:00:00.000Z",
  updatedAt: "2026-06-14T00:00:00.000Z",
};

const providerBeta: ProviderConfigSummary = {
  id: "provider-beta",
  authorId: "author-1",
  name: "Provider Beta",
  baseUrl: "https://beta.example",
  createdAt: "2026-06-14T00:00:00.000Z",
  updatedAt: "2026-06-14T00:00:00.000Z",
};

function modelVariant(id: string, providerConfigId: string): ModelVariant {
  return {
    id,
    providerConfigId,
    name: id,
    model: `${id}-model`,
    temperature: 0.7,
    maxOutputTokens: 1200,
    contextWindowTokens: 8000,
    inputPricePerMillion: 1,
    outputPricePerMillion: 2,
    cacheReadPricePerMillion: 0,
    cacheWritePricePerMillion: 0,
    requestTokenField: "max_tokens",
    reasoningFormat: "none",
    compatibilityJson: "{}",
    createdAt: "2026-06-14T00:00:00.000Z",
    updatedAt: "2026-06-14T00:00:00.000Z",
  };
}

describe("modelSettingsSlice", () => {
  it("selects a provider", () => {
    const selected = modelSettingsReducer(
      {
        ...initialModelSettingsState,
        providers: [providerAlpha, providerBeta],
      },
      modelSettingsActions.setSelectedProviderId("provider-beta"),
    );

    expect(selected.selectedProviderId).toBe("provider-beta");
  });

  it("selects a model variant", () => {
    const selected = modelSettingsReducer(
      {
        ...initialModelSettingsState,
        selectedProviderId: "provider-alpha",
        modelVariantsByProviderId: {
          "provider-alpha": [modelVariant("model-alpha", "provider-alpha")],
        },
      },
      modelSettingsActions.setActiveModelVariantId("model-alpha"),
    );

    expect(selected.activeModelVariantId).toBe("model-alpha");
  });

  it("clears the selected model when the provider changes to a provider without that model", () => {
    const selected = modelSettingsReducer(
      {
        ...initialModelSettingsState,
        selectedProviderId: "provider-alpha",
        activeModelVariantId: "model-alpha",
        providers: [providerAlpha, providerBeta],
        modelVariantsByProviderId: {
          "provider-alpha": [modelVariant("model-alpha", "provider-alpha")],
          "provider-beta": [modelVariant("model-beta", "provider-beta")],
        },
      },
      modelSettingsActions.setSelectedProviderId("provider-beta"),
    );

    expect(selected.selectedProviderId).toBe("provider-beta");
    expect(selected.activeModelVariantId).toBeNull();
  });

  it("resets project-scoped model settings state", () => {
    const loaded = modelSettingsReducer(
      {
        ...initialModelSettingsState,
        selectedProviderId: "provider-1",
        activeModelVariantId: "model-1",
        providers: [providerAlpha],
        modelVariantsByProviderId: {
          "provider-1": [modelVariant("model-1", "provider-1")],
        },
        providersStatus: "succeeded",
        modelsStatus: "failed",
        error: "Could not load models",
      },
      modelSettingsActions.resetForProject(),
    );

    expect(loaded).toEqual(initialModelSettingsState);
  });

  it("defaults to the first provider when provider configs load without a current selection", () => {
    const loading = modelSettingsReducer(initialModelSettingsState, loadProviderConfigs.pending("request-1"));

    const loaded = modelSettingsReducer(loading, loadProviderConfigs.fulfilled([providerAlpha, providerBeta], "request-1"));

    expect(loaded.providers).toEqual([providerAlpha, providerBeta]);
    expect(loaded.selectedProviderId).toBe("provider-alpha");
  });

  it("ignores stale provider config loads", () => {
    const loadingFirst = modelSettingsReducer(initialModelSettingsState, loadProviderConfigs.pending("request-1"));
    const loadingSecond = modelSettingsReducer(loadingFirst, loadProviderConfigs.pending("request-2"));

    const loadedSecond = modelSettingsReducer(loadingSecond, loadProviderConfigs.fulfilled([providerBeta], "request-2"));

    const staleLoadedFirst = modelSettingsReducer(
      loadedSecond,
      loadProviderConfigs.fulfilled([providerAlpha], "request-1"),
    );

    expect(staleLoadedFirst.providers).toEqual([providerBeta]);
    expect(staleLoadedFirst.selectedProviderId).toBe("provider-beta");
    expect(staleLoadedFirst.providersStatus).toBe("succeeded");
    expect(staleLoadedFirst.error).toBeNull();
  });

  it("defaults to the first model for the selected provider when models load", () => {
    const state: ModelSettingsState = {
      ...initialModelSettingsState,
      selectedProviderId: "provider-alpha",
      providers: [providerAlpha],
    };

    const loading = modelSettingsReducer(
      state,
      loadModelVariants.pending("request-1", { providerId: "provider-alpha" }),
    );

    const loaded = modelSettingsReducer(
      loading,
      loadModelVariants.fulfilled(
        [modelVariant("model-alpha", "provider-alpha"), modelVariant("model-beta", "provider-alpha")],
        "request-1",
        { providerId: "provider-alpha" },
      ),
    );

    expect(loaded.modelVariantsByProviderId["provider-alpha"]).toHaveLength(2);
    expect(loaded.activeModelVariantId).toBe("model-alpha");
  });

  it("ignores stale model loads for a provider that is no longer selected", () => {
    const state: ModelSettingsState = {
      ...initialModelSettingsState,
      selectedProviderId: "provider-beta",
      activeModelVariantId: "model-beta",
      providers: [providerAlpha, providerBeta],
      modelVariantsByProviderId: {
        "provider-beta": [modelVariant("model-beta", "provider-beta")],
      },
    };

    const loaded = modelSettingsReducer(
      state,
      loadModelVariants.pending("request-1", { providerId: "provider-alpha" }),
    );

    const selectedOtherProvider = modelSettingsReducer(
      loaded,
      modelSettingsActions.setSelectedProviderId("provider-beta"),
    );

    const staleLoaded = modelSettingsReducer(
      selectedOtherProvider,
      loadModelVariants.fulfilled([modelVariant("model-alpha", "provider-alpha")], "request-1", {
        providerId: "provider-alpha",
      }),
    );

    expect(staleLoaded.selectedProviderId).toBe("provider-beta");
    expect(staleLoaded.activeModelVariantId).toBe("model-beta");
    expect(staleLoaded.modelsStatus).toBe("idle");
    expect(staleLoaded.loadingModelProviderId).toBeNull();
    expect(staleLoaded.modelVariantsByProviderId["provider-alpha"]).toBeUndefined();
  });

  it("ignores stale duplicate model loads for the same provider", () => {
    const state: ModelSettingsState = {
      ...initialModelSettingsState,
      selectedProviderId: "provider-alpha",
      providers: [providerAlpha],
    };

    const loadingFirst = modelSettingsReducer(
      state,
      loadModelVariants.pending("request-1", { providerId: "provider-alpha" }),
    );
    const loadingSecond = modelSettingsReducer(
      loadingFirst,
      loadModelVariants.pending("request-2", { providerId: "provider-alpha" }),
    );

    const staleRejectedFirst = modelSettingsReducer(
      loadingSecond,
      loadModelVariants.rejected(new Error("First request failed"), "request-1", {
        providerId: "provider-alpha",
      }),
    );

    const loadedSecond = modelSettingsReducer(
      staleRejectedFirst,
      loadModelVariants.fulfilled([modelVariant("model-alpha-latest", "provider-alpha")], "request-2", {
        providerId: "provider-alpha",
      }),
    );

    expect(loadedSecond.modelsStatus).toBe("succeeded");
    expect(loadedSecond.error).toBeNull();
    expect(loadedSecond.loadingModelProviderId).toBeNull();
    expect(loadedSecond.modelVariantsByProviderId["provider-alpha"]).toEqual([
      modelVariant("model-alpha-latest", "provider-alpha"),
    ]);
    expect(loadedSecond.activeModelVariantId).toBe("model-alpha-latest");
  });
});

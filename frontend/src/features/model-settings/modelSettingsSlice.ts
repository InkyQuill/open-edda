import { createSlice } from "@reduxjs/toolkit";

import type { ModelVariant, ProviderConfigSummary } from "../../agentTypes";
import { loadModelVariants, loadProviderConfigs } from "./modelSettingsThunks";

export type AsyncStatus = "idle" | "pending" | "succeeded" | "failed";

export type ModelSettingsState = {
  selectedProviderId: string | null;
  activeModelVariantId: string | null;
  providers: ProviderConfigSummary[];
  modelVariantsByProviderId: Record<string, ModelVariant[]>;
  providersStatus: AsyncStatus;
  modelsStatus: AsyncStatus;
  loadingModelProviderId: string | null;
  error: string | null;
};

export const initialModelSettingsState: ModelSettingsState = {
  selectedProviderId: null,
  activeModelVariantId: null,
  providers: [],
  modelVariantsByProviderId: {},
  providersStatus: "idle",
  modelsStatus: "idle",
  loadingModelProviderId: null,
  error: null,
};

function modelBelongsToProvider(
  modelVariantsByProviderId: Record<string, ModelVariant[]>,
  modelVariantId: string | null,
  providerId: string | null,
): boolean {
  if (!modelVariantId || !providerId) return false;
  return (modelVariantsByProviderId[providerId] ?? []).some((model) => model.id === modelVariantId);
}

function errorMessage(actionErrorMessage: string | undefined, fallback: string): string {
  return actionErrorMessage ?? fallback;
}

const modelSettingsSlice = createSlice({
  name: "modelSettings",
  initialState: initialModelSettingsState,
  reducers: {
    setSelectedProviderId(state, action: { payload: string | null }) {
      state.selectedProviderId = action.payload;
      if (!modelBelongsToProvider(state.modelVariantsByProviderId, state.activeModelVariantId, action.payload)) {
        state.activeModelVariantId = null;
      }
      if (state.loadingModelProviderId !== action.payload) {
        state.modelsStatus = action.payload && state.modelVariantsByProviderId[action.payload] ? "succeeded" : "idle";
      }
    },
    setActiveModelVariantId(state, action: { payload: string | null }) {
      if (!action.payload) {
        state.activeModelVariantId = null;
        return;
      }
      if (modelBelongsToProvider(state.modelVariantsByProviderId, action.payload, state.selectedProviderId)) {
        state.activeModelVariantId = action.payload;
      }
    },
    resetForProject() {
      return initialModelSettingsState;
    },
  },
  extraReducers: (builder) => {
    builder
      .addCase(loadProviderConfigs.pending, (state) => {
        state.providersStatus = "pending";
        state.error = null;
      })
      .addCase(loadProviderConfigs.fulfilled, (state, action) => {
        state.providers = action.payload;
        state.providersStatus = "succeeded";
        state.error = null;

        const selectedProviderStillExists = action.payload.some((provider) => provider.id === state.selectedProviderId);
        if (!state.selectedProviderId || !selectedProviderStillExists) {
          state.selectedProviderId = action.payload[0]?.id ?? null;
        }

        if (!modelBelongsToProvider(state.modelVariantsByProviderId, state.activeModelVariantId, state.selectedProviderId)) {
          state.activeModelVariantId = null;
        }
      })
      .addCase(loadProviderConfigs.rejected, (state, action) => {
        state.providersStatus = "failed";
        state.error = errorMessage(action.error.message, "Could not load providers");
      })
      .addCase(loadModelVariants.pending, (state, action) => {
        state.loadingModelProviderId = action.meta.arg.providerId;
        state.modelsStatus = "pending";
        state.error = null;
      })
      .addCase(loadModelVariants.fulfilled, (state, action) => {
        const providerId = action.meta.arg.providerId;
        if (providerId !== state.loadingModelProviderId) return;
        state.loadingModelProviderId = null;

        if (providerId !== state.selectedProviderId) {
          state.modelsStatus = "idle";
          return;
        }

        state.modelVariantsByProviderId[providerId] = action.payload;
        state.modelsStatus = "succeeded";
        state.error = null;

        if (!modelBelongsToProvider(state.modelVariantsByProviderId, state.activeModelVariantId, providerId)) {
          state.activeModelVariantId = action.payload[0]?.id ?? null;
        }
      })
      .addCase(loadModelVariants.rejected, (state, action) => {
        const providerId = action.meta.arg.providerId;
        if (providerId !== state.loadingModelProviderId) return;
        state.loadingModelProviderId = null;

        if (providerId !== state.selectedProviderId) return;
        state.modelsStatus = "failed";
        state.error = errorMessage(action.error.message, "Could not load models");
      });
  },
});

export const modelSettingsActions = modelSettingsSlice.actions;
export const modelSettingsReducer = modelSettingsSlice.reducer;

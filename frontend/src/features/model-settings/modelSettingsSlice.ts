import { createSlice } from "@reduxjs/toolkit";

export type AsyncStatus = "idle" | "pending" | "succeeded" | "failed";

export type ModelSettingsState = {
  selectedProviderId: string | null;
  activeModelVariantId: string | null;
  providersStatus: AsyncStatus;
  modelsStatus: AsyncStatus;
  error: string | null;
};

export const initialModelSettingsState: ModelSettingsState = {
  selectedProviderId: null,
  activeModelVariantId: null,
  providersStatus: "idle",
  modelsStatus: "idle",
  error: null,
};

const modelSettingsSlice = createSlice({
  name: "modelSettings",
  initialState: initialModelSettingsState,
  reducers: {
    resetForProject() {
      return initialModelSettingsState;
    },
  },
});

export const modelSettingsActions = modelSettingsSlice.actions;
export const modelSettingsReducer = modelSettingsSlice.reducer;

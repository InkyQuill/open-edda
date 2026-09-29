import { createSlice, type PayloadAction } from "@reduxjs/toolkit";

export type SettingsState = {
  selectedProjectId: string | null;
};

export const initialSettingsState: SettingsState = {
  selectedProjectId: null,
};

const settingsSlice = createSlice({
  name: "settings",
  initialState: initialSettingsState,
  reducers: {
    setSelectedProjectId(state, action: PayloadAction<string | null>) {
      state.selectedProjectId = action.payload;
    },
  },
});

export const settingsActions = settingsSlice.actions;
export const settingsReducer = settingsSlice.reducer;

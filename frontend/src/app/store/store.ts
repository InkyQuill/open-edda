import { configureStore } from "@reduxjs/toolkit";
import { assistantReducer } from "../../features/assistant/assistantSlice";
import { assistantActionsReducer } from "../../features/assistant-actions/assistantActionsSlice";
import { editorReducer } from "../../features/editor/editorSlice";
import { modelSettingsReducer } from "../../features/model-settings/modelSettingsSlice";
import { reviewReducer } from "../../features/review/reviewSlice";
import { settingsReducer } from "../../features/settings/settingsSlice";
import { scriptRuntimeReducer } from "../../features/script-runtime/scriptRuntimeSlice";
import { skillsReducer } from "../../features/skills/skillsSlice";
import { workspaceReducer } from "../../features/workspace/workspaceSlice";

export const store = configureStore({
  reducer: {
    assistant: assistantReducer,
    assistantActions: assistantActionsReducer,
    editor: editorReducer,
    modelSettings: modelSettingsReducer,
    review: reviewReducer,
    scriptRuntime: scriptRuntimeReducer,
    settings: settingsReducer,
    skills: skillsReducer,
    workspace: workspaceReducer,
  },
});

export type RootState = ReturnType<typeof store.getState>;
export type AppDispatch = typeof store.dispatch;

import { configureStore } from "@reduxjs/toolkit";
import { assistantReducer } from "../../features/assistant/assistantSlice";
import { editorReducer } from "../../features/editor/editorSlice";
import { modelSettingsReducer } from "../../features/model-settings/modelSettingsSlice";
import { reviewReducer } from "../../features/review/reviewSlice";
import { scriptRuntimeReducer } from "../../features/script-runtime/scriptRuntimeSlice";
import { skillsReducer } from "../../features/skills/skillsSlice";
import { workspaceReducer } from "../../features/workspace/workspaceSlice";

export const store = configureStore({
  reducer: {
    workspace: workspaceReducer,
    editor: editorReducer,
    assistant: assistantReducer,
    modelSettings: modelSettingsReducer,
    review: reviewReducer,
    scriptRuntime: scriptRuntimeReducer,
    skills: skillsReducer,
  },
});

export type RootState = ReturnType<typeof store.getState>;
export type AppDispatch = typeof store.dispatch;

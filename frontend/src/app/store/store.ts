import { configureStore } from "@reduxjs/toolkit";
import { editorReducer } from "../../features/editor/editorSlice";
import { workspaceReducer } from "../../features/workspace/workspaceSlice";

export const store = configureStore({
  reducer: {
    editor: editorReducer,
    workspace: workspaceReducer,
  },
});

export type RootState = ReturnType<typeof store.getState>;
export type AppDispatch = typeof store.dispatch;

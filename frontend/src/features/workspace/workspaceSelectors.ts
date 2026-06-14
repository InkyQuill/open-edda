import type { WorkspaceState } from "./workspaceSlice";

type WorkspaceRootState = {
  workspace: WorkspaceState;
};

export const selectWorkspaceMode = (state: WorkspaceRootState) =>
  state.workspace.mode;

export const selectWorkspaceProjectState = (
  state: WorkspaceRootState,
) => {
  const { mobileSheet: _mobileSheet, ...projectState } = state.workspace;
  return projectState;
};

export const selectIsRightDrawerVisible = (state: WorkspaceRootState) =>
  state.workspace.rightDrawerOpen;

export const selectActiveRightTab = (state: WorkspaceRootState) =>
  state.workspace.activeRightTab;

export const selectMobileSheet = (state: WorkspaceRootState) =>
  state.workspace.mobileSheet;

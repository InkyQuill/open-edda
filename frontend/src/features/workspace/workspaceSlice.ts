import { createSlice, type PayloadAction } from "@reduxjs/toolkit";

export type WorkspaceMode = "draft" | "assistant" | "review";
export type DrawerTab =
  | "contents"
  | "world"
  | "notes"
  | "chat"
  | "tools"
  | "revisions"
  | "model";
export type MobileSheet =
  | null
  | "contents"
  | "assistant"
  | "review"
  | "world-notes"
  | "model";

export type WorkspaceProjectState = {
  mode: WorkspaceMode;
  leftDrawerOpen: boolean;
  rightDrawerOpen: boolean;
  leftDrawerWidth: number;
  rightDrawerWidth: number;
  activeLeftTab: DrawerTab;
  activeRightTab: DrawerTab;
  lastContentKind: string;
  fallbackContentId: string | null;
};

export type WorkspaceState = WorkspaceProjectState & {
  mobileSheet: MobileSheet;
};

export const initialWorkspaceState: WorkspaceState = {
  mode: "assistant",
  leftDrawerOpen: true,
  rightDrawerOpen: true,
  leftDrawerWidth: 304,
  rightDrawerWidth: 380,
  activeLeftTab: "contents",
  activeRightTab: "chat",
  lastContentKind: "chapter",
  fallbackContentId: null,
  mobileSheet: null,
};

const initialWorkspaceProjectState: WorkspaceProjectState = {
  mode: initialWorkspaceState.mode,
  leftDrawerOpen: initialWorkspaceState.leftDrawerOpen,
  rightDrawerOpen: initialWorkspaceState.rightDrawerOpen,
  leftDrawerWidth: initialWorkspaceState.leftDrawerWidth,
  rightDrawerWidth: initialWorkspaceState.rightDrawerWidth,
  activeLeftTab: initialWorkspaceState.activeLeftTab,
  activeRightTab: initialWorkspaceState.activeRightTab,
  lastContentKind: initialWorkspaceState.lastContentKind,
  fallbackContentId: initialWorkspaceState.fallbackContentId,
};

export const modeDefaults: Record<
  WorkspaceMode,
  Pick<WorkspaceState, "leftDrawerOpen" | "rightDrawerOpen" | "activeRightTab">
> = {
  draft: {
    leftDrawerOpen: true,
    rightDrawerOpen: false,
    activeRightTab: "chat",
  },
  assistant: {
    leftDrawerOpen: true,
    rightDrawerOpen: true,
    activeRightTab: "chat",
  },
  review: {
    leftDrawerOpen: true,
    rightDrawerOpen: true,
    activeRightTab: "tools",
  },
};

const workspaceSlice = createSlice({
  name: "workspace",
  initialState: initialWorkspaceState,
  reducers: {
    setMode(state, action: PayloadAction<WorkspaceMode>) {
      state.mode = action.payload;
      Object.assign(state, modeDefaults[action.payload]);
    },
    setLeftDrawerOpen(state, action: PayloadAction<boolean>) {
      state.leftDrawerOpen = action.payload;
    },
    setRightDrawerOpen(state, action: PayloadAction<boolean>) {
      state.rightDrawerOpen = action.payload;
    },
    setDrawerWidths(
      state,
      action: PayloadAction<{ left?: number; right?: number }>,
    ) {
      if (action.payload.left !== undefined) {
        state.leftDrawerWidth = Math.min(
          420,
          Math.max(240, action.payload.left),
        );
      }
      if (action.payload.right !== undefined) {
        state.rightDrawerWidth = Math.min(
          520,
          Math.max(320, action.payload.right),
        );
      }
    },
    setActiveLeftTab(state, action: PayloadAction<DrawerTab>) {
      state.activeLeftTab = action.payload;
    },
    setActiveRightTab(state, action: PayloadAction<DrawerTab>) {
      state.activeRightTab = action.payload;
      state.rightDrawerOpen = true;
    },
    setMobileSheet(state, action: PayloadAction<MobileSheet>) {
      state.mobileSheet = action.payload;
    },
    setLastContentKind(state, action: PayloadAction<string>) {
      state.lastContentKind = action.payload;
      state.fallbackContentId = null;
    },
    setFallbackContentId(state, action: PayloadAction<string | null>) {
      state.fallbackContentId = action.payload;
    },
    hydrateProjectState(
      state,
      action: PayloadAction<Partial<WorkspaceProjectState>>,
    ) {
      Object.assign(state, {
        ...initialWorkspaceProjectState,
        ...action.payload,
        mobileSheet: state.mobileSheet,
      });
    },
  },
});

export const workspaceActions = workspaceSlice.actions;
export const workspaceReducer = workspaceSlice.reducer;

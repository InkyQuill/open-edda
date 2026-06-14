import type { WorkspaceMode, WorkspaceProjectState } from "./workspaceSlice";

const browserKey = "open_edda_workspace_browser";
const projectPrefix = "open_edda_workspace_project:";

export type BrowserWorkspaceState = {
  defaultMode: WorkspaceMode;
  density: "comfortable" | "compact";
};

const defaultBrowserWorkspaceState: BrowserWorkspaceState = {
  defaultMode: "assistant",
  density: "comfortable",
};

function getStorage(storage?: Storage): Storage | null {
  if (storage) return storage;
  if (typeof window === "undefined") return null;
  return window.localStorage;
}

function isWorkspaceMode(value: unknown): value is WorkspaceMode {
  return value === "draft" || value === "assistant" || value === "review";
}

function isDrawerTab(
  value: unknown,
): value is WorkspaceProjectState["activeLeftTab"] {
  return (
    value === "contents" ||
    value === "world" ||
    value === "notes" ||
    value === "chat" ||
    value === "tools" ||
    value === "revisions" ||
    value === "model"
  );
}

function numberInRange(
  value: unknown,
  min: number,
  max: number,
): value is number {
  return (
    typeof value === "number" &&
    Number.isFinite(value) &&
    value >= min &&
    value <= max
  );
}

function parseProjectWorkspaceState(
  value: unknown,
): Partial<WorkspaceProjectState> {
  if (!value || typeof value !== "object") {
    return {};
  }
  const parsed = value as Record<string, unknown>;
  const state: Partial<WorkspaceProjectState> = {};

  if (isWorkspaceMode(parsed.mode)) state.mode = parsed.mode;
  if (typeof parsed.leftDrawerOpen === "boolean") {
    state.leftDrawerOpen = parsed.leftDrawerOpen;
  }
  if (typeof parsed.rightDrawerOpen === "boolean") {
    state.rightDrawerOpen = parsed.rightDrawerOpen;
  }
  if (numberInRange(parsed.leftDrawerWidth, 240, 420)) {
    state.leftDrawerWidth = parsed.leftDrawerWidth;
  }
  if (numberInRange(parsed.rightDrawerWidth, 320, 520)) {
    state.rightDrawerWidth = parsed.rightDrawerWidth;
  }
  if (isDrawerTab(parsed.activeLeftTab)) state.activeLeftTab = parsed.activeLeftTab;
  if (isDrawerTab(parsed.activeRightTab)) state.activeRightTab = parsed.activeRightTab;
  if (
    typeof parsed.lastContentKind === "string" &&
    parsed.lastContentKind.length > 0
  ) {
    state.lastContentKind = parsed.lastContentKind;
  }
  if (
    typeof parsed.fallbackContentId === "string" ||
    parsed.fallbackContentId === null
  ) {
    state.fallbackContentId = parsed.fallbackContentId;
  }

  return state;
}

export function loadBrowserWorkspaceState(
  storage?: Storage,
): BrowserWorkspaceState {
  const availableStorage = getStorage(storage);
  if (!availableStorage) return defaultBrowserWorkspaceState;

  const raw = availableStorage.getItem(browserKey);
  if (!raw) return defaultBrowserWorkspaceState;

  try {
    const parsed = JSON.parse(raw) as Partial<BrowserWorkspaceState>;
    return {
      defaultMode: isWorkspaceMode(parsed.defaultMode) ? parsed.defaultMode : "assistant",
      density: parsed.density === "compact" ? "compact" : "comfortable",
    };
  } catch {
    return defaultBrowserWorkspaceState;
  }
}

export function saveBrowserWorkspaceState(
  value: BrowserWorkspaceState,
  storage?: Storage,
): void {
  getStorage(storage)?.setItem(browserKey, JSON.stringify(value));
}

export function loadProjectWorkspaceState(
  projectId: string,
  storage?: Storage,
): Partial<WorkspaceProjectState> {
  const availableStorage = getStorage(storage);
  if (!availableStorage) return {};

  const raw = availableStorage.getItem(`${projectPrefix}${projectId}`);
  if (!raw) return {};

  try {
    return parseProjectWorkspaceState(JSON.parse(raw));
  } catch {
    return {};
  }
}

export function saveProjectWorkspaceState(
  projectId: string,
  value: WorkspaceProjectState,
  storage?: Storage,
): void {
  getStorage(storage)?.setItem(
    `${projectPrefix}${projectId}`,
    JSON.stringify(value),
  );
}

import { describe, expect, it } from "vitest";
import {
  initialWorkspaceState,
  type WorkspaceProjectState,
  workspaceActions,
  workspaceReducer,
} from "./workspaceSlice";
import {
  loadProjectWorkspaceState,
  saveProjectWorkspaceState,
} from "./workspacePersistence";

class MemoryStorage implements Storage {
  readonly #items = new Map<string, string>();

  get length(): number {
    return this.#items.size;
  }

  clear(): void {
    this.#items.clear();
  }

  getItem(key: string): string | null {
    return this.#items.get(key) ?? null;
  }

  key(index: number): string | null {
    return Array.from(this.#items.keys())[index] ?? null;
  }

  removeItem(key: string): void {
    this.#items.delete(key);
  }

  setItem(key: string, value: string): void {
    this.#items.set(key, value);
  }
}

describe("workspaceSlice", () => {
  it("applies preset drawer defaults", () => {
    const draft = workspaceReducer(
      initialWorkspaceState,
      workspaceActions.setMode("draft"),
    );

    expect(draft.rightDrawerOpen).toBe(false);
    expect(draft.leftDrawerOpen).toBe(true);

    const review = workspaceReducer(draft, workspaceActions.setMode("review"));

    expect(review.rightDrawerOpen).toBe(true);
    expect(review.activeRightTab).toBe("tools");
  });

  it("clamps drawer widths", () => {
    const state = workspaceReducer(
      initialWorkspaceState,
      workspaceActions.setDrawerWidths({ left: 120, right: 900 }),
    );

    expect(state.leftDrawerWidth).toBe(240);
    expect(state.rightDrawerWidth).toBe(520);
  });

  it("hydrates project state from defaults instead of previous project state", () => {
    const projectA = workspaceReducer(
      initialWorkspaceState,
      workspaceActions.hydrateProjectState({
        mode: "review",
        rightDrawerOpen: true,
        activeRightTab: "tools",
        lastContentKind: "project_note",
        fallbackContentId: "project-a-note",
      }),
    );

    const projectB = workspaceReducer(
      projectA,
      workspaceActions.hydrateProjectState({}),
    );

    expect(projectB.mode).toBe(initialWorkspaceState.mode);
    expect(projectB.activeRightTab).toBe(initialWorkspaceState.activeRightTab);
    expect(projectB.lastContentKind).toBe(initialWorkspaceState.lastContentKind);
    expect(projectB.fallbackContentId).toBeNull();
  });

  it("clears fallback content when content kind changes", () => {
    const state = workspaceReducer(
      { ...initialWorkspaceState, fallbackContentId: "chapter-1" },
      workspaceActions.setLastContentKind("story_bible_entry"),
    );

    expect(state.lastContentKind).toBe("story_bible_entry");
    expect(state.fallbackContentId).toBeNull();
  });

  it("roundtrips project workspace state through provided storage", () => {
    const storage = new MemoryStorage();
    const projectState: WorkspaceProjectState = {
      mode: "review",
      leftDrawerOpen: true,
      rightDrawerOpen: true,
      leftDrawerWidth: 304,
      rightDrawerWidth: 380,
      activeLeftTab: "contents",
      activeRightTab: "tools",
      lastContentKind: "chapter",
      fallbackContentId: "content-1",
    };

    saveProjectWorkspaceState("project-1", projectState, storage);

    expect(loadProjectWorkspaceState("project-1", storage)).toEqual(
      projectState,
    );
  });

  it("drops invalid project workspace persistence fields", () => {
    const storage = new MemoryStorage();
    storage.setItem(
      "open_edda_workspace_project:project-1",
      JSON.stringify({
        mode: "broken",
        leftDrawerOpen: "yes",
        rightDrawerOpen: true,
        leftDrawerWidth: 999,
        rightDrawerWidth: 420,
        activeLeftTab: "unknown",
        activeRightTab: "chat",
        lastContentKind: "",
        fallbackContentId: 42,
        mobileSheet: "assistant",
      }),
    );

    expect(loadProjectWorkspaceState("project-1", storage)).toEqual({
      rightDrawerOpen: true,
      rightDrawerWidth: 420,
      activeRightTab: "chat",
    });
  });
});

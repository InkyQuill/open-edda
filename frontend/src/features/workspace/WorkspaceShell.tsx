import {
  Bot,
  ClipboardCheck,
  FolderOpen,
  Library,
  MessageSquare,
  PenLine,
  PanelLeftClose,
  PanelLeftOpen,
  Settings2,
  Menu,
} from "lucide-react";
import { useEffect } from "react";
import { useDispatch, useSelector } from "react-redux";
import { Link } from "react-router-dom";

import { AssistantDrawer } from "../assistant/AssistantDrawer";
import { EditorFrame } from "../editor/EditorFrame";
import { ContextDrawer } from "../notes/ContextDrawer";
import { ReviewDrawer } from "../review/ReviewDrawer";
import { Button } from "../../shared/ui/button";
import { Sheet, SheetContent, SheetHeader, SheetTitle } from "../../shared/ui/sheet";
import type { ContentItem, ContentKind } from "../../types";
import type { DrawerTab, MobileSheet, WorkspaceMode, WorkspaceState } from "./workspaceSlice";
import { workspaceActions } from "./workspaceSlice";

type WorkspaceRootState = {
  workspace: WorkspaceState;
};

type ContextTab = "contents" | "world" | "notes";

type WorkspaceShellProps = {
  projectId: string;
  projectTitle: string;
  contentItems: ContentItem[];
  contentLoading: boolean;
  contentError: string | null;
  contentCreateError: string | null;
  contentCreating: boolean;
  activeContentKind: ContentKind;
  selectedContent: ContentItem | null;
  onCreateContent: (kind: ContentKind) => void;
  onSelectContent: (item: ContentItem) => void;
  onContentKindChange: (kind: ContentKind) => void;
  onContentSaved: (item: ContentItem) => void;
};

const modeButtons: Array<{
  mode: WorkspaceMode;
  label: string;
  icon: typeof PenLine;
}> = [
  { mode: "draft", label: "Draft", icon: PenLine },
  { mode: "assistant", label: "Assistant", icon: Bot },
  { mode: "review", label: "Review", icon: ClipboardCheck },
];

const mobileButtons: Array<{
  sheet: NonNullable<MobileSheet>;
  label: string;
  icon: typeof FolderOpen;
}> = [
  { sheet: "contents", label: "Files", icon: FolderOpen },
  { sheet: "assistant", label: "Assistant", icon: MessageSquare },
  { sheet: "review", label: "Review", icon: ClipboardCheck },
  { sheet: "world-notes", label: "World/Notes", icon: Library },
];

function toContextTab(tab: DrawerTab): ContextTab {
  return tab === "world" || tab === "notes" ? tab : "contents";
}

function mobileSheetTitle(sheet: NonNullable<MobileSheet>): string {
  switch (sheet) {
    case "contents":
      return "Files";
    case "assistant":
      return "Assistant";
    case "review":
      return "Review";
    case "world-notes":
      return "World and notes";
  }
}

export function WorkspaceShell({
  projectId,
  projectTitle,
  contentItems,
  contentLoading,
  contentError,
  contentCreateError,
  contentCreating,
  activeContentKind,
  selectedContent,
  onCreateContent,
  onSelectContent,
  onContentKindChange,
  onContentSaved,
}: WorkspaceShellProps) {
  useEffect(() => {
    const viewport = window.visualViewport;
    const updateHeight = () => {
      document.documentElement.style.setProperty("--workspace-viewport-height", `${viewport?.height ?? window.innerHeight}px`);
    };
    updateHeight();
    viewport?.addEventListener("resize", updateHeight);
    window.addEventListener("resize", updateHeight);
    return () => {
      viewport?.removeEventListener("resize", updateHeight);
      window.removeEventListener("resize", updateHeight);
      document.documentElement.style.removeProperty("--workspace-viewport-height");
    };
  }, []);
  const dispatch = useDispatch();
  const workspace = useSelector((state: WorkspaceRootState) => state.workspace);
  const activeLeftTab = toContextTab(workspace.activeLeftTab);
  const rightDrawer =
    workspace.mode === "review" || workspace.activeRightTab === "tools" || workspace.activeRightTab === "revisions" ? (
      <ReviewDrawer projectId={projectId} />
    ) : (
      <AssistantDrawer projectId={projectId} />
    );

  const contextDrawer = (
    <ContextDrawer
      activeTab={activeLeftTab}
      contentItems={contentItems}
      contentLoading={contentLoading}
      contentError={contentError}
      contentCreateError={contentCreateError}
      contentCreating={contentCreating}
      contentCreationDisabled={contentCreating || contentLoading}
      activeContentKind={activeContentKind}
      selectedContentId={selectedContent?.id ?? null}
      onCreateContent={onCreateContent}
      onSelectContent={onSelectContent}
      onContentKindChange={onContentKindChange}
      onTabChange={(tab) => dispatch(workspaceActions.setActiveLeftTab(tab))}
    />
  );

  function renderMobileSheet(sheet: NonNullable<MobileSheet>) {
    if (sheet === "assistant") return <AssistantDrawer projectId={projectId} />;
    if (sheet === "review") return <ReviewDrawer projectId={projectId} />;
    return (
      <ContextDrawer
        activeTab={sheet === "world-notes" ? "world" : "contents"}
        contentItems={contentItems}
        contentLoading={contentLoading}
        contentError={contentError}
        contentCreateError={contentCreateError}
        contentCreating={contentCreating}
        contentCreationDisabled={contentCreating || contentLoading}
        activeContentKind={activeContentKind}
        selectedContentId={selectedContent?.id ?? null}
        onCreateContent={onCreateContent}
        onSelectContent={onSelectContent}
        onContentKindChange={onContentKindChange}
        onTabChange={(tab) => dispatch(workspaceActions.setActiveLeftTab(tab))}
      />
    );
  }

  return (
    <main className="writing-workspace flex h-dvh min-h-0 flex-col bg-background text-foreground">
      <header className="flex items-center justify-between gap-3 border-b border-border px-4 py-3">
        <div className="min-w-0">
          <Link to="/projects" className="hidden text-xs text-muted-foreground hover:text-foreground md:inline">Edda / Projects</Link>
          <h1 className="truncate text-sm font-semibold md:text-base">{projectTitle}</h1>
        </div>
        <details className="mobile-workspace-menu relative md:hidden">
          <summary className="flex min-h-10 cursor-pointer items-center gap-1 px-2 text-xs"><Menu className="size-4" aria-hidden="true" />Menu</summary>
          <nav className="absolute right-0 top-full z-30 flex w-56 flex-col gap-1 rounded-md border border-border bg-popover p-2" aria-label="Workspace menu">
            <Button asChild variant="ghost"><Link to="/projects">Projects</Link></Button>
            <Button asChild variant="ghost"><Link to={`/settings?projectId=${encodeURIComponent(projectId)}`}>Settings</Link></Button>
            {modeButtons.map(({ mode, label }) => <Button key={mode} variant={workspace.mode === mode ? "secondary" : "ghost"} aria-pressed={workspace.mode === mode} onClick={() => dispatch(workspaceActions.setMode(mode))}>{label}</Button>)}
          </nav>
        </details>
        <div className="hidden items-center gap-2 md:flex">
          <Button variant="ghost" size="icon" aria-label={workspace.leftDrawerOpen ? "Hide project navigation" : "Show project navigation"} aria-pressed={workspace.leftDrawerOpen} onClick={() => dispatch(workspaceActions.setLeftDrawerOpen(!workspace.leftDrawerOpen))}>{workspace.leftDrawerOpen ? <PanelLeftClose aria-hidden="true" /> : <PanelLeftOpen aria-hidden="true" />}</Button>
          <Button asChild type="button" variant="outline">
            <Link to={`/settings?projectId=${encodeURIComponent(projectId)}`}>
              <Settings2 data-icon="inline-start" aria-hidden="true" />
              Settings
            </Link>
          </Button>
          {modeButtons.map(({ mode, label, icon: Icon }) => (
            <Button
              key={mode}
              aria-pressed={workspace.mode === mode}
              type="button"
              variant={workspace.mode === mode ? "secondary" : "ghost"}
              onClick={() => dispatch(workspaceActions.setMode(mode))}
            >
              <Icon data-icon="inline-start" aria-hidden="true" />
              {label}
            </Button>
          ))}
        </div>
      </header>

      <div className="flex min-h-0 flex-1">
        {workspace.leftDrawerOpen ? (
          <aside className="workspace-navigation hidden min-h-0 shrink-0 border-r border-border bg-sidebar p-4 md:block" style={{ width: workspace.leftDrawerWidth }}>
            {contextDrawer}
          </aside>
        ) : null}

        <section className="workspace-editor-stage flex min-w-0 flex-1 justify-center overflow-auto bg-background px-4 py-6">
          <EditorFrame
            projectId={projectId}
            content={selectedContent}
            mode={workspace.mode}
            contentLoading={contentLoading}
            contentError={contentError}
            onContentSaved={onContentSaved}
          />
        </section>

        {workspace.rightDrawerOpen ? (
          <aside className="workspace-tools hidden min-h-0 shrink-0 border-l border-border bg-card p-4 md:block" style={{ width: workspace.rightDrawerWidth }}>
            {rightDrawer}
          </aside>
        ) : null}
      </div>

      <nav className="grid grid-cols-4 border-t border-border bg-background p-2 pb-[max(0.5rem,env(safe-area-inset-bottom))] md:hidden" aria-label="Workspace panels">
        {mobileButtons.map(({ sheet, label, icon: Icon }) => (
          <Button
            key={sheet}
            type="button"
            variant={workspace.mobileSheet === sheet ? "secondary" : "ghost"}
            className="h-auto min-h-12 flex-col gap-1 px-1 py-2 text-xs"
            onClick={() => dispatch(workspaceActions.setMobileSheet(sheet))}
          >
            <Icon data-icon="inline-start" aria-hidden="true" />
            <span>{label}</span>
          </Button>
        ))}
      </nav>

      <Sheet
        open={workspace.mobileSheet !== null}
        onOpenChange={(open) => {
          if (!open) dispatch(workspaceActions.setMobileSheet(null));
        }}
      >
        {workspace.mobileSheet ? (
          <SheetContent side="bottom" className="max-h-[85dvh] overflow-auto">
            <SheetHeader>
              <SheetTitle>{mobileSheetTitle(workspace.mobileSheet)}</SheetTitle>
            </SheetHeader>
            <div className="min-h-0 px-4 pb-4">{renderMobileSheet(workspace.mobileSheet)}</div>
          </SheetContent>
        ) : null}
      </Sheet>
    </main>
  );
}
